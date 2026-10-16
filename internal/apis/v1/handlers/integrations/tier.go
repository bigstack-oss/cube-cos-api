package integrations

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	ostime "time"

	dephttp "github.com/bigstack-oss/bigstack-dependency-go/pkg/http"
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/bodies"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/base"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/storages"
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	log "go-micro.dev/v5/logger"
)

// Tier requests are synchronous. The Ceph part (device class, CRUSH rule, pool,
// volume type) runs once here; the registry entry and Cinder apply run on every control.

// tierOps wraps hex_sdk; replaced in tests
var tierOps = struct {
	get        func(name string) (*storages.Tier, error)
	list       func() ([]storages.Tier, error)
	create     func(name string, osdIds []string) (string, error)
	remove     func(name string) (string, error)
	register   func(name string) (string, error)
	unregister func(name string) (string, error)
	inUse      func(name string) bool
	restart    func() error
	forget     func(name string) error
}{
	get:        cubecos.GetStorageTier,
	list:       cubecos.ListStorageTiers,
	create:     cubecos.CreateStorageTier,
	remove:     cubecos.DeleteStorageTier,
	register:   cubecos.RegisterStorageTier,
	unregister: cubecos.UnregisterStorageTier,
	inUse:      cubecos.IsStorageTierInUse,
	restart:    cubecos.RestartCinderVolume,
	forget:     cubecos.ForgetStorageTierService,
}

// peerControls lists the other controls; none when HA is off
var peerControls = func() ([]nodes.Node, error) {
	if !base.IsHaEnabled {
		return nil, nil
	}

	return nodes.GetPeerControls()
}

// tierHttp calls peers with a 12m timeout and no retries
var tierHttp = newTierHttp()

func newTierHttp() *dephttp.Helper {
	h, _ := dephttp.NewHelper(dephttp.Timeout(12*ostime.Minute), dephttp.RetryCount(0))
	return h
}

// one tier change at a time on this node
var tierMu sync.Mutex

var errTierBusy = errors.New("another storage tier change is in progress on this node")

type tierHttpErr struct {
	code int
	err  error
}

func (e *tierHttpErr) Error() string {
	return e.err.Error()
}

func tierConflict(format string, a ...any) error {
	return &tierHttpErr{code: http.StatusConflict, err: fmt.Errorf(format, a...)}
}

func tierNotFound(format string, a ...any) error {
	return &tierHttpErr{code: http.StatusNotFound, err: fmt.Errorf(format, a...)}
}

func setTierErr(c *gin.Context, err error) {
	var he *tierHttpErr
	switch {
	case errors.Is(err, errTierBusy):
		bodies.SetConflict(c, err)
	case errors.As(err, &he) && he.code == http.StatusConflict:
		bodies.SetConflict(c, err)
	case errors.As(err, &he) && he.code == http.StatusNotFound:
		bodies.SetNotFound(c, err)
	default:
		bodies.SetInternalServerError(c, err)
	}
}

func listStorageTiers(c *gin.Context) {
	h, err := initHelper(c, "listStorageTiers")
	if err != nil {
		bodies.SetBadRequest(c, err, nil)
		return
	}

	tiers, err := tierOps.list()
	if err != nil {
		log.Errorf("storage tiers(%s): %v", h.reqId, err)
		bodies.SetInternalServerError(c, err)
		return
	}

	bodies.SetOk(c, "fetch storage tiers successfully", tiers)
}

func createStorageTier(c *gin.Context) {
	h, err := initHelper(c, "createStorageTier")
	if err != nil {
		bodies.SetBadRequest(c, err, nil)
		return
	}

	out, err := h.createStorageTier()
	if err != nil {
		log.Errorf("storage tiers(%s): create %s: %v", h.reqId, h.tierReqOpts.Name, err)
		setTierErr(c, err)
		return
	}

	bodies.SetOk(
		c,
		fmt.Sprintf("storage tier %s created", h.tierReqOpts.Name),
		storages.TierResult{Name: h.tierReqOpts.Name, Output: out},
	)
}

func deleteStorageTier(c *gin.Context) {
	h, err := initHelper(c, "deleteStorageTier")
	if err != nil {
		bodies.SetBadRequest(c, err, nil)
		return
	}

	out, err := h.deleteStorageTier()
	if err != nil {
		log.Errorf("storage tiers(%s): delete %s: %v", h.reqId, h.tierReqOpts.Name, err)
		setTierErr(c, err)
		return
	}

	bodies.SetOk(
		c,
		fmt.Sprintf("storage tier %s deleted", h.tierReqOpts.Name),
		storages.TierResult{Name: h.tierReqOpts.Name, Output: out},
	)
}

func (h *helper) parseCreateStorageTierParams() error {
	err := h.c.ShouldBindJSON(&h.tierReqOpts)
	if err != nil {
		return fmt.Errorf("failed to parse storage tier creation req(%v)", err)
	}

	if h.tierReqOpts.Name == "" {
		return errors.New("storage tier name is required")
	}

	// a forwarded request only registers the tier
	if len(h.tierReqOpts.OsdIds) == 0 && !h.isPeerRequest() {
		return errors.New("at least one OSD id is required")
	}

	return nil
}

func (h *helper) parseDeleteStorageTierParams() error {
	h.tierReqOpts.Name = h.c.Param("tierName")
	if h.tierReqOpts.Name == "" {
		return errors.New("storage tier name is required")
	}

	return nil
}

// createStorageTier runs the full create here, then registers the tier on
// every other control and restarts cinder-volume once they all have it
func (h *helper) createStorageTier() (string, error) {
	if !tierMu.TryLock() {
		return "", errTierBusy
	}
	defer tierMu.Unlock()

	name := h.tierReqOpts.Name
	if h.isPeerRequest() {
		return tierOps.register(name)
	}

	tier, err := tierOps.get(name)
	if err != nil {
		return "", err
	}
	if tier != nil && tier.State == storages.TierUnmanaged {
		return "", tierConflict("%s is a device class, CRUSH rule and pool not in the storage tier registry; it was not created here", name)
	}

	peers, err := peerControls()
	if err != nil {
		return "", fmt.Errorf("cannot list the peer controls, nothing was changed: %v", err)
	}

	out, err := tierOps.create(name, h.tierReqOpts.OsdIds)
	if err != nil {
		return out, err
	}

	err = h.registerTierOn(peers)
	if err != nil {
		return out, fmt.Errorf("storage tier %s was created but %v; run the create again to finish", name, err)
	}

	return out, nil
}

// deleteStorageTier takes the tier out of every control's registry and has
// cinder-volume drop it before the Ceph teardown deletes its pool
func (h *helper) deleteStorageTier() (string, error) {
	if !tierMu.TryLock() {
		return "", errTierBusy
	}
	defer tierMu.Unlock()

	name := h.tierReqOpts.Name
	if h.isPeerRequest() {
		return tierOps.unregister(name)
	}

	tier, err := tierOps.get(name)
	if err != nil {
		return "", err
	}
	if tier == nil {
		return "", tierNotFound("no such storage tier: %s", name)
	}
	if !tier.IsOwned() {
		return "", tierConflict("%s is not in the storage tier registry; it was not created here", name)
	}

	registryOnly := tier.State == storages.TierRegistryOnly
	if !registryOnly && tierOps.inUse(name) {
		return "", tierConflict("volume type %s is still in use, or the volume list could not be read; nothing was deleted", name)
	}

	peers, err := peerControls()
	if err != nil {
		return "", fmt.Errorf("cannot list the peer controls, nothing was deleted: %v", err)
	}

	out, err := h.unregisterTierEverywhere(peers)
	if err != nil {
		return out, err
	}
	if !registryOnly {
		delOut, err := tierOps.remove(name)
		out = joinOutput(out, delOut)
		if err != nil {
			return out, h.restoreTierAfterFailedDelete(peers, err)
		}
	}

	// drop the removed backend's "down" service record
	err = tierOps.forget(name)
	if err != nil {
		out = joinOutput(out, fmt.Sprintf("Warning: cinder-volume service cube@%s was not removed: %v", name, err))
	}

	return out, nil
}

// registerTierOn registers the tier on each peer in turn, then restarts cinder-volume
func (h *helper) registerTierOn(peers []nodes.Node) error {
	if len(peers) == 0 {
		return nil
	}

	body := storages.TierReqOpts{Name: h.tierReqOpts.Name}
	for _, node := range peers {
		err := h.sendTierReqToPeer(node, http.MethodPost, node.PostStorageTierUrl(), body)
		if err != nil {
			return fmt.Errorf("registering it on %s failed: %v", node.Hostname, err)
		}
	}

	err := tierOps.restart()
	if err != nil {
		return fmt.Errorf("restarting cinder-volume failed: %v", err)
	}

	return nil
}

// unregisterTierEverywhere unregisters the tier here and on every peer, then
// restarts cinder-volume; on failure it re-registers it where it was removed
func (h *helper) unregisterTierEverywhere(peers []nodes.Node) (string, error) {
	name := h.tierReqOpts.Name
	out, err := tierOps.unregister(name)
	if err != nil {
		return out, fmt.Errorf("cannot unregister storage tier %s here, nothing was deleted: %v", name, err)
	}

	done := []nodes.Node{}
	for _, node := range peers {
		err = h.sendTierReqToPeer(node, http.MethodDelete, node.DeleteStorageTierUrl(name), nil)
		if err != nil {
			err = fmt.Errorf("cannot unregister storage tier %s on %s, nothing was deleted: %v", name, node.Hostname, err)
			return out, h.reregisterTier(done, true, err)
		}
		done = append(done, node)
	}

	if len(peers) > 0 {
		err = tierOps.restart()
		if err != nil {
			err = fmt.Errorf("restarting cinder-volume failed, nothing was deleted: %v", err)
			return out, h.reregisterTier(done, true, err)
		}
	}

	return out, nil
}

// restoreTierAfterFailedDelete re-registers the tier on every control while its
// objects are still there
func (h *helper) restoreTierAfterFailedDelete(peers []nodes.Node, cause error) error {
	tier, err := tierOps.get(h.tierReqOpts.Name)
	if err != nil {
		return fmt.Errorf("%v; its state could not be read afterwards, so the registry was not put back on the other controls: %v", cause, err)
	}
	if tier == nil || tier.State == storages.TierRegistryOnly {
		return cause
	}

	// unmanaged: objects intact, only the local registry entry is missing
	local := tier.State == storages.TierUnmanaged
	return h.reregisterTier(peers, local, cause)
}

func (h *helper) reregisterTier(peers []nodes.Node, local bool, cause error) error {
	name := h.tierReqOpts.Name
	failed := []string{}
	if local {
		_, err := tierOps.register(name)
		if err != nil {
			failed = append(failed, base.Hostname)
		}
	}

	err := h.registerTierOn(peers)
	if err != nil {
		failed = append(failed, err.Error())
	}

	if len(failed) > 0 {
		return fmt.Errorf("%v; putting the registry entry back also failed (%s), re-register it with 'storage tier create %s'", cause, strings.Join(failed, "; "), name)
	}

	return fmt.Errorf("%v; its registry entry was put back", cause)
}

func (h *helper) sendTierReqToPeer(node nodes.Node, method, url string, body any) error {
	req := tierHttp.R().SetHeaders(nodes.GetForwardedHeaders())
	if body != nil {
		req.SetHeader("Content-Type", "application/json").SetBody(body)
	}

	resp, err := req.Execute(method, url)
	if err != nil {
		return err
	}
	if resp.IsError() {
		return fmt.Errorf("%s", respMsg(resp))
	}

	return nil
}

func respMsg(resp *resty.Response) string {
	msg := struct {
		Msg string `json:"msg"`
	}{}
	err := json.Unmarshal(resp.Body(), &msg)
	if err == nil && msg.Msg != "" {
		return msg.Msg
	}

	return fmt.Sprintf("%s %s", resp.Status(), strings.TrimSpace(resp.String()))
}

func joinOutput(a, b string) string {
	if a == "" {
		return b
	}
	if b == "" {
		return a
	}

	return a + "\n" + b
}
