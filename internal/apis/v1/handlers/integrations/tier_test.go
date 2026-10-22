package integrations

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/storages"
	"github.com/gin-gonic/gin"
)

type tierRig struct {
	t      *testing.T
	events []string
	state  map[string]string // tier name -> state; absent = not listed
	failOn map[string]bool   // event -> fail
	peers  []nodes.Node
}

func newTierRig(t *testing.T, peerNames ...string) *tierRig {
	r := &tierRig{t: t, state: map[string]string{}, failOn: map[string]bool{}}

	tierOps.get = func(name string) (*storages.Tier, error) {
		st, ok := r.state[name]
		if !ok {
			return nil, nil
		}
		return &storages.Tier{Name: name, State: st}, nil
	}
	tierOps.create = func(name string, ids []string) (string, error) {
		return r.local("create "+name+" "+strings.Join(ids, ","), "created")
	}
	tierOps.remove = func(name string) (string, error) {
		return r.local("remove "+name, "deleted")
	}
	tierOps.register = func(name string) (string, error) {
		return r.local("register "+name, "")
	}
	tierOps.unregister = func(name string) (string, error) {
		return r.local("unregister "+name, "")
	}
	tierOps.inUse = func(name string) bool {
		r.events = append(r.events, "inUse "+name)
		return r.failOn["inUse "+name]
	}
	tierOps.restart = func() error {
		_, err := r.local("restart", "")
		return err
	}
	tierOps.forget = func(name string) error {
		_, err := r.local("forget "+name, "")
		return err
	}

	for _, p := range peerNames {
		r.peers = append(r.peers, r.peer(p))
	}
	peerControls = func() ([]nodes.Node, error) { return r.peers, nil }
	return r
}

func (r *tierRig) local(ev, out string) (string, error) {
	r.events = append(r.events, ev)
	if r.failOn[ev] {
		return "boom: " + ev, errors.New("boom: " + ev)
	}
	return out, nil
}

// peer answers like a control running the tier handlers
func (r *tierRig) peer(name string) nodes.Node {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Header.Get(nodes.ForwardedHeader) == "" {
			r.t.Errorf("%s: request without %s", name, nodes.ForwardedHeader)
		}
		body, _ := io.ReadAll(req.Body)
		opts := storages.TierReqOpts{}
		_ = json.Unmarshal(body, &opts)
		ev := ""
		switch req.Method {
		case http.MethodPost:
			ev = name + " register " + opts.Name
			if len(opts.OsdIds) > 0 {
				r.t.Errorf("%s: forwarded create carries OSD ids %v", name, opts.OsdIds)
			}
		case http.MethodDelete:
			ev = name + " unregister " + req.URL.Path[strings.LastIndex(req.URL.Path, "/")+1:]
		}
		r.events = append(r.events, ev)
		if r.failOn[ev] {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"code":500,"msg":"apply failed"}`))
			return
		}
		_, _ = w.Write([]byte(`{"code":200,"msg":"ok"}`))
	}))
	r.t.Cleanup(srv.Close)
	return nodes.Node{Hostname: name, Protocol: "http", Address: strings.TrimPrefix(srv.URL, "http://")}
}

func (r *tierRig) helper(method, body string, forwarded bool) *helper {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(method, "/api/v1/datacenters/dc/integrations/storages/tiers", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Request.Header.Set("Node", "cube45")
	if forwarded {
		c.Request.Header.Set(nodes.ForwardedHeader, "cube45")
		c.Set("isNode", true)
	}
	return &helper{c: c}
}

func (r *tierRig) create(name string, ids []string, forwarded bool) (string, error) {
	b, _ := json.Marshal(storages.TierReqOpts{Name: name, OsdIds: ids})
	h := r.helper(http.MethodPost, string(b), forwarded)
	if err := h.parseCreateStorageTierParams(); err != nil {
		return "", err
	}
	return h.createStorageTier()
}

func (r *tierRig) delete(name string, forwarded bool) (string, error) {
	h := r.helper(http.MethodDelete, "", forwarded)
	h.tierReqOpts.Name = name
	return h.deleteStorageTier()
}

func (r *tierRig) want(label string, want ...string) {
	r.t.Helper()
	if strings.Join(r.events, " | ") != strings.Join(want, " | ") {
		r.t.Errorf("%s:\n got  %q\n want %q", label, r.events, want)
	}
}

func wantErr(t *testing.T, err error, code int, has ...string) {
	t.Helper()
	if err == nil {
		t.Fatalf("want an error containing %q", has)
	}
	for _, s := range has {
		if !strings.Contains(err.Error(), s) {
			t.Errorf("error %q lacks %q", err, s)
		}
	}
	var he *tierHttpErr
	got := http.StatusInternalServerError
	if errors.As(err, &he) {
		got = he.code
	}
	if got != code {
		t.Errorf("error %q: code %d, want %d", err, got, code)
	}
}

func TestCreateStorageTierFansOutToEveryControl(t *testing.T) {
	r := newTierRig(t, "cube46", "cube47")
	out, err := r.create("gold", []string{"0", "1"}, false)
	if err != nil || out != "created" {
		t.Fatalf("create: %q %v", out, err)
	}
	r.want("Ceph once here, register on each peer, then one restart",
		"create gold 0,1", "cube46 register gold", "cube47 register gold", "restart")
}

func TestCreateStorageTierSingleNode(t *testing.T) {
	r := newTierRig(t)
	if _, err := r.create("gold", []string{"0"}, false); err != nil {
		t.Fatal(err)
	}
	r.want("no peers, no extra restart", "create gold 0")
}

func TestCreateStorageTierForwardedOnlyRegisters(t *testing.T) {
	r := newTierRig(t, "cube46")
	if _, err := r.create("gold", nil, true); err != nil {
		t.Fatal(err)
	}
	r.want("a peer runs the per-control part only", "register gold")
}

func TestCreateStorageTierNeedsOsds(t *testing.T) {
	r := newTierRig(t)
	if _, err := r.create("gold", nil, false); err == nil {
		t.Fatal("create without OSD ids was accepted")
	}
	r.want("nothing ran")
}

func TestCreateStorageTierPeerFailure(t *testing.T) {
	r := newTierRig(t, "cube46", "cube47")
	r.failOn["cube46 register gold"] = true
	_, err := r.create("gold", []string{"0"}, false)
	wantErr(t, err, http.StatusInternalServerError, "cube46", "apply failed", "run the create again")
	r.want("stops at the failed peer, no restart", "create gold 0", "cube46 register gold")
}

func TestCreateStorageTierRefusesUnmanaged(t *testing.T) {
	r := newTierRig(t, "cube46")
	r.state["gold"] = storages.TierUnmanaged
	_, err := r.create("gold", []string{"0"}, false)
	wantErr(t, err, http.StatusConflict, "not in the storage tier registry")
	r.want("nothing ran")
}

func TestDeleteStorageTierUnregistersEverywhereBeforeTeardown(t *testing.T) {
	r := newTierRig(t, "cube46", "cube47")
	r.state["gold"] = storages.TierRegistered
	if _, err := r.delete("gold", false); err != nil {
		t.Fatal(err)
	}
	r.want("every control drops it and cinder-volume restarts before the pool goes",
		"inUse gold", "unregister gold", "cube46 unregister gold", "cube47 unregister gold", "restart", "remove gold", "forget gold")
}

func TestDeleteStorageTierServiceRecordFailureOnlyWarns(t *testing.T) {
	r := newTierRig(t, "cube46")
	r.state["gold"] = storages.TierRegistered
	r.failOn["forget gold"] = true
	out, err := r.delete("gold", false)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "cube@gold was not removed") {
		t.Fatalf("no warning in %q", out)
	}
	r.want("deleted, the record left with a warning",
		"inUse gold", "unregister gold", "cube46 unregister gold", "restart", "remove gold", "forget gold")
}

func TestDeleteStorageTierInUse(t *testing.T) {
	r := newTierRig(t, "cube46")
	r.state["gold"] = storages.TierRegistered
	r.failOn["inUse gold"] = true
	_, err := r.delete("gold", false)
	wantErr(t, err, http.StatusConflict, "still in use", "nothing was deleted")
	r.want("refused before any change", "inUse gold")
}

func TestDeleteStorageTierNotFound(t *testing.T) {
	r := newTierRig(t, "cube46")
	_, err := r.delete("gold", false)
	wantErr(t, err, http.StatusNotFound, "no such storage tier")

	r.state["gold"] = storages.TierUnmanaged
	_, err = r.delete("gold", false)
	wantErr(t, err, http.StatusConflict, "not in the storage tier registry")
	r.want("nothing ran")
}

func TestDeleteStorageTierPeerFailurePutsRegistryBack(t *testing.T) {
	r := newTierRig(t, "cube46", "cube47")
	r.state["gold"] = storages.TierRegistered
	r.failOn["cube47 unregister gold"] = true
	_, err := r.delete("gold", false)
	wantErr(t, err, http.StatusInternalServerError, "cube47", "nothing was deleted", "put back")
	r.want("re-registered where it was taken out, nothing deleted",
		"inUse gold", "unregister gold", "cube46 unregister gold", "cube47 unregister gold",
		"register gold", "cube46 register gold", "restart")
}

func TestDeleteStorageTierCephRefusalPutsRegistryBack(t *testing.T) {
	r := newTierRig(t, "cube46")
	r.state["gold"] = storages.TierRegistered
	r.failOn["remove gold"] = true
	// unregistered here, the objects intact: the tier now reads unmanaged
	tierOps.unregister = func(name string) (string, error) {
		r.state[name] = storages.TierUnmanaged
		return r.local("unregister "+name, "")
	}
	_, err := r.delete("gold", false)
	wantErr(t, err, http.StatusInternalServerError, "boom: remove gold", "put back")
	r.want("registered back on every control",
		"inUse gold", "unregister gold", "cube46 unregister gold", "restart", "remove gold",
		"register gold", "cube46 register gold", "restart")
}

func TestDeleteStorageTierPartialDeleteLeavesRegistryOut(t *testing.T) {
	r := newTierRig(t, "cube46")
	r.state["gold"] = storages.TierRegistered
	r.failOn["remove gold"] = true
	// pool gone, rule kept: no longer listed
	tierOps.remove = func(name string) (string, error) {
		delete(r.state, name)
		return r.local("remove "+name, "")
	}
	_, err := r.delete("gold", false)
	wantErr(t, err, http.StatusInternalServerError, "boom: remove gold")
	r.want("no registry entry for a pool that is gone",
		"inUse gold", "unregister gold", "cube46 unregister gold", "restart", "remove gold")
}

func TestDeleteStorageTierRegistryOnly(t *testing.T) {
	r := newTierRig(t, "cube46")
	r.state["gold"] = storages.TierRegistryOnly
	if _, err := r.delete("gold", false); err != nil {
		t.Fatal(err)
	}
	r.want("registry entries only, no Ceph teardown",
		"unregister gold", "cube46 unregister gold", "restart", "forget gold")
}

func TestDeleteStorageTierForwardedOnlyUnregisters(t *testing.T) {
	r := newTierRig(t, "cube46")
	if _, err := r.delete("gold", true); err != nil {
		t.Fatal(err)
	}
	r.want("a peer runs the per-control part only", "unregister gold")
}

func TestStorageTierChangesAreSerialized(t *testing.T) {
	r := newTierRig(t)
	tierMu.Lock()
	defer tierMu.Unlock()
	_, err := r.create("gold", []string{"0"}, false)
	if !errors.Is(err, errTierBusy) {
		t.Fatalf("got %v, want errTierBusy", err)
	}
}

func TestStorageTierRoutesResolve(t *testing.T) {
	gin.SetMode(gin.TestMode)
	e := gin.New()
	hit := ""
	for _, hd := range Handlers {
		path := hd.Path
		e.Handle(hd.Method, path, func(c *gin.Context) { hit = c.FullPath() })
	}

	for _, tc := range []struct{ method, path, want string }{
		{http.MethodGet, "/integrations/storages/tiers", "/integrations/storages/tiers"},
		{http.MethodPost, "/integrations/storages/tiers", "/integrations/storages/tiers"},
		{http.MethodDelete, "/integrations/storages/tiers/gold", "/integrations/storages/tiers/:tierName"},
		{http.MethodDelete, "/integrations/storages/gold", "/integrations/storages/:storageName"},
	} {
		hit = ""
		e.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(tc.method, tc.path, nil))
		if hit != tc.want {
			t.Errorf("%s %s routed to %q, want %q", tc.method, tc.path, hit, tc.want)
		}
	}
}
