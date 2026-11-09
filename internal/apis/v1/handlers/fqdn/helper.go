package fqdn

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	ostime "time"

	dephttp "github.com/bigstack-oss/bigstack-dependency-go/pkg/http"
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/bodies"
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/queries"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/base"
	deffqdn "github.com/bigstack-oss/cube-cos-api/internal/definition/v1/fqdn"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	log "go-micro.dev/v5/logger"
)

// ops wraps hex_sdk and the node list; replaced in tests
var ops = struct {
	getNode       func() (*deffqdn.Node, error)
	getFqdn       func() (string, error)
	verifyCert    func(name string) error
	installCert   func(dir string) (*deffqdn.Certificate, error)
	reloadCert    func() error
	setTuning     func(name string) error
	applyCluster  func(certOnly bool) (string, error)
	restartApi    func() error
	listNodes     func() []nodes.Node
	sendToPeer    func(node nodes.Node, method, url string, body any) ([]byte, error)
	isControlRole func(role string) bool
}{
	getNode:       cubecos.GetFqdnNode,
	getFqdn:       cubecos.GetFqdn,
	verifyCert:    cubecos.VerifyFqdnCertificate,
	installCert:   cubecos.InstallFqdnCertificate,
	reloadCert:    cubecos.ReloadFqdnCertificate,
	setTuning:     cubecos.SetFqdnTuning,
	applyCluster:  cubecos.ApplyFqdnCluster,
	restartApi:    cubecos.RestartApiSoon,
	listNodes:     nodes.List,
	sendToPeer:    sendToPeer,
	isControlRole: isControlRole,
}

// one FQDN change at a time on this node
var fqdnMu sync.Mutex

var errBusy = errors.New("another FQDN change is in progress on this node")

// peerHttp calls peers with no retries: a node's commit can take minutes, and repeating it
// would only restart its services again
var peerHttp = newPeerHttp()

func newPeerHttp() *dephttp.Helper {
	h, _ := dephttp.NewHelper(dephttp.Timeout(30*ostime.Minute), dephttp.RetryCount(0))
	return h
}

type httpErr struct {
	code int
	err  error
}

func (e *httpErr) Error() string {
	return e.err.Error()
}

func badRequest(format string, a ...any) error {
	return &httpErr{code: http.StatusBadRequest, err: fmt.Errorf(format, a...)}
}

func conflict(format string, a ...any) error {
	return &httpErr{code: http.StatusConflict, err: fmt.Errorf(format, a...)}
}

// setErr answers err in the shape bodies uses, with data -- the step log of a switch that
// stopped part way -- when there is any: the operator needs to see how far it got
func setErr(c *gin.Context, err error, data any) {
	code := http.StatusInternalServerError
	var he *httpErr
	switch {
	case errors.Is(err, errBusy):
		code = http.StatusConflict
	case errors.As(err, &he):
		code = he.code
	}

	resp := gin.H{
		bodies.Code:   code,
		bodies.Status: strings.ToLower(http.StatusText(code)),
		bodies.Msg:    err.Error(),
	}
	if data != nil {
		resp[bodies.Data] = data
	}
	c.JSON(code, resp)
}

// isNodeRequest is true for a request carrying a node token: the CLI on a node, or a peer.
// Changing the FQDN or installing a private key is only taken from there.
func isNodeRequest(c *gin.Context) bool {
	node := c.GetHeader("Node")
	return node != "" && c.GetHeader("Authorization") == "Bearer "+nodes.GenToken(node)
}

func isForwarded(c *gin.Context) bool {
	return isNodeRequest(c) && c.GetHeader(nodes.ForwardedHeader) != ""
}

func requireNode(c *gin.Context) error {
	if !isNodeRequest(c) {
		return &httpErr{code: http.StatusForbidden, err: errors.New("the FQDN is changed from the CLI on a control node")}
	}

	return nil
}

func requireForwarded(c *gin.Context) error {
	if !isForwarded(c) {
		return &httpErr{code: http.StatusForbidden, err: errors.New("only a control node calls this, for another node")}
	}

	return nil
}

func isControlRole(role string) bool {
	return slices.Contains(
		[]string{nodes.RoleControl, nodes.RoleControlConverged, nodes.RoleEdgeCore, nodes.RoleModerator},
		role,
	)
}

func requireControl() error {
	if !ops.isControlRole(base.CurrentRole) {
		return badRequest("run this on a control node, not on a %s node", base.CurrentRole)
	}

	return nil
}

// peers are every other node, controls first: they serve the web UIs, so they move before
// the computes, whose only part is the console URL
func peers() []nodes.Node {
	controls, others := []nodes.Node{}, []nodes.Node{}
	seen := map[string]bool{base.Hostname: true}
	for _, n := range ops.listNodes() {
		if n.IsLocal() || seen[n.Hostname] {
			continue
		}
		seen[n.Hostname] = true

		if ops.isControlRole(n.Role) {
			controls = append(controls, n)
		} else {
			others = append(others, n)
		}
	}

	return append(controls, others...)
}

func peerControls() []nodes.Node {
	controls := []nodes.Node{}
	for _, n := range peers() {
		if ops.isControlRole(n.Role) {
			controls = append(controls, n)
		}
	}

	return controls
}

func sendToPeer(node nodes.Node, method, url string, body any) ([]byte, error) {
	req := peerHttp.R().SetHeaders(nodes.GetForwardedHeaders())
	if body != nil {
		req.SetHeader("Content-Type", "application/json").SetBody(body)
	}

	resp, err := req.Execute(method, url)
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, errors.New(respMsg(resp))
	}

	return resp.Body(), nil
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

func reqId(c *gin.Context) string {
	return queries.GetReqId(c)
}

func getFqdn(c *gin.Context) {
	err := requireNode(c)
	if err != nil {
		setErr(c, err, nil)
		return
	}

	status, err := collectStatus()
	if err != nil {
		log.Errorf("fqdn(%s): %v", reqId(c), err)
		setErr(c, err, nil)
		return
	}

	bodies.SetOk(c, "fetch fqdn successfully", status)
}

// collectStatus is this node's view, with every other node's
func collectStatus() (*deffqdn.Status, error) {
	local, err := ops.getNode()
	if err != nil {
		return nil, err
	}

	status := &deffqdn.Status{
		Fqdn:        local.Fqdn,
		PublicHost:  local.PublicHost,
		Certificate: local.Certificate,
		Nodes:       []deffqdn.Node{*local},
	}

	for _, n := range peers() {
		status.Nodes = append(status.Nodes, peerStatus(n))
	}

	return status, nil
}

func peerStatus(n nodes.Node) deffqdn.Node {
	failed := deffqdn.Node{Hostname: n.Hostname, Role: n.Role}
	if n.IsDown() {
		failed.Error = "down"
		return failed
	}

	b, err := ops.sendToPeer(n, http.MethodGet, n.FqdnNodeUrl(), nil)
	if err != nil {
		failed.Error = err.Error()
		return failed
	}

	resp := struct {
		Data deffqdn.Node `json:"data"`
	}{}
	err = json.Unmarshal(b, &resp)
	if err != nil {
		failed.Error = fmt.Sprintf("unexpected answer: %v", err)
		return failed
	}

	return resp.Data
}

func setFqdn(c *gin.Context) {
	req := deffqdn.SetReq{}
	err := c.ShouldBindJSON(&req)
	if err != nil {
		setErr(c, badRequest("the request body is not valid: %v", err), nil)
		return
	}

	name := deffqdn.Normalize(req.Fqdn)
	if name == "" {
		setErr(c, badRequest("no FQDN given; reset it to serve the web UIs on the control VIP"), nil)
		return
	}

	switchFqdn(c, name)
}

func resetFqdn(c *gin.Context) {
	switchFqdn(c, "")
}

// switchFqdn moves the web UIs onto name, or back to the control VIP for ""
func switchFqdn(c *gin.Context, name string) {
	out, err := doSwitch(c, name)
	result := deffqdn.Result{Fqdn: name, Output: out}
	if err != nil {
		log.Errorf("fqdn(%s): switching to %q: %v", reqId(c), name, err)
		setErr(c, err, result)
		return
	}

	msg := fmt.Sprintf("the web UIs are served on %s", name)
	if name == "" {
		msg = "the FQDN is reset; the web UIs are served on the control VIP"
	}
	bodies.SetOk(c, msg, result)
}

func doSwitch(c *gin.Context, name string) (string, error) {
	err := requireNode(c)
	if err != nil {
		return "", err
	}

	err = requireControl()
	if err != nil {
		return "", err
	}

	if name != "" && !deffqdn.IsValid(name) {
		return "", badRequest("%s is not a valid FQDN", name)
	}

	if !fqdnMu.TryLock() {
		return "", errBusy
	}
	defer fqdnMu.Unlock()

	if name != "" {
		err = ops.verifyCert(name)
		if err != nil {
			return "", badRequest("%v", err)
		}
	}

	// every node carries the FQDN, so refuse while one cannot be reached rather than leave
	// the cluster half switched
	targets := peers()
	for _, n := range targets {
		if n.IsDown() {
			return "", conflict("node %s is down; every node takes the FQDN, nothing was changed", n.Hostname)
		}
	}

	steps := []string{}
	err = ops.setTuning(name)
	if err != nil {
		return "", fmt.Errorf("%s: %v; nothing else was changed", base.Hostname, err)
	}
	steps = append(steps, fmt.Sprintf("fqdn: %s committed", base.Hostname))

	for _, n := range targets {
		_, err = ops.sendToPeer(n, http.MethodPut, n.FqdnNodeUrl(), deffqdn.SetReq{Fqdn: name})
		if err != nil {
			return strings.Join(steps, "\n"), fmt.Errorf(
				"%s: %v; the nodes before it are switched, run the command again to finish", n.Hostname, err)
		}
		steps = append(steps, fmt.Sprintf("fqdn: %s committed", n.Hostname))
	}

	out, err := ops.applyCluster(false)
	steps = append(steps, out)
	if err != nil {
		return strings.Join(steps, "\n"), fmt.Errorf("%v; every node is switched, run the command again to finish", err)
	}

	// the public host, the SAML SP and the links are read when the API starts
	restartApis(c, targets)
	steps = append(steps, "fqdn: cube-cos-api restarts on every node")

	return strings.Join(steps, "\n"), nil
}

func restartApis(c *gin.Context, targets []nodes.Node) {
	for _, n := range targets {
		_, err := ops.sendToPeer(n, http.MethodPut, n.FqdnNodeApiUrl(), nil)
		if err != nil {
			log.Errorf("fqdn(%s): restarting cube-cos-api on %s: %v", reqId(c), n.Hostname, err)
		}
	}

	err := ops.restartApi()
	if err != nil {
		log.Errorf("fqdn(%s): restarting cube-cos-api here: %v", reqId(c), err)
	}
}

func installCertificate(c *gin.Context) {
	cert, out, err := doInstall(c)
	if err != nil {
		log.Errorf("fqdn(%s): installing the certificate: %v", reqId(c), err)
		setErr(c, err, deffqdn.Result{Output: out})
		return
	}

	bodies.SetOk(c, "the certificate is installed", cert)
}

func doInstall(c *gin.Context) (*deffqdn.Certificate, string, error) {
	err := requireNode(c)
	if err != nil {
		return nil, "", err
	}

	err = requireControl()
	if err != nil {
		return nil, "", err
	}

	req := deffqdn.CertificateReq{}
	err = c.ShouldBindJSON(&req)
	if err != nil {
		return nil, "", badRequest("the request body is not valid: %v", err)
	}
	if req.Certificate == "" || req.PrivateKey == "" || req.CaCertificate == "" {
		return nil, "", badRequest("the certificate, its private key and the CA certificate are all required")
	}

	if !fqdnMu.TryLock() {
		return nil, "", errBusy
	}
	defer fqdnMu.Unlock()

	dir, err := writeCertificateSet(req)
	if dir != "" {
		defer os.RemoveAll(dir)
	}
	if err != nil {
		return nil, "", err
	}

	cert, err := ops.installCert(dir)
	if err != nil {
		return nil, "", badRequest("%v", err)
	}

	name, err := ops.getFqdn()
	if err != nil || name == "" {
		// nothing serves it until an FQDN is set
		return cert, "", nil
	}

	// a replacement for the FQDN in use goes in front of browsers now
	out := []string{}
	err = ops.reloadCert()
	if err != nil {
		return cert, "", fmt.Errorf("the certificate is installed, but %s did not reload it: %v", base.Hostname, err)
	}
	for _, n := range peerControls() {
		_, err = ops.sendToPeer(n, http.MethodPut, n.FqdnNodeCertificateUrl(), nil)
		if err != nil {
			return cert, strings.Join(out, "\n"), fmt.Errorf("the certificate is installed, but %s did not reload it: %v", n.Hostname, err)
		}
		out = append(out, fmt.Sprintf("fqdn: %s reloaded", n.Hostname))
	}

	clusterOut, err := ops.applyCluster(true)
	out = append(out, clusterOut)
	if err != nil {
		return cert, strings.Join(out, "\n"), err
	}

	return cert, strings.Join(out, "\n"), nil
}

// certStageDir holds the request's PEMs until hex_sdk installs them: tmpfs, so the private
// key never reaches a disk here; replaced in tests
var certStageDir = "/run"

// writeCertificateSet puts the request's PEMs where hex_sdk fqdn_cert_install reads them, in
// a directory only root can read; the caller removes it
func writeCertificateSet(req deffqdn.CertificateReq) (string, error) {
	dir, err := os.MkdirTemp(certStageDir, "cube-fqdn-cert-")
	if err != nil {
		return "", err
	}

	err = os.Chmod(dir, 0700)
	if err != nil {
		return dir, err
	}

	files := map[string]string{
		deffqdn.CertificateFile:   req.Certificate,
		deffqdn.PrivateKeyFile:    req.PrivateKey,
		deffqdn.CaCertificateFile: req.CaCertificate,
	}
	for name, content := range files {
		err = os.WriteFile(filepath.Join(dir, name), []byte(content), 0600)
		if err != nil {
			return dir, err
		}
	}

	return dir, nil
}

func getNode(c *gin.Context) {
	err := requireForwarded(c)
	if err != nil {
		setErr(c, err, nil)
		return
	}

	node, err := ops.getNode()
	if err != nil {
		setErr(c, err, nil)
		return
	}

	bodies.SetOk(c, "fetch fqdn successfully", node)
}

// setNode commits the FQDN a control asked for on this node, "" removing it
func setNode(c *gin.Context) {
	err := requireForwarded(c)
	if err != nil {
		setErr(c, err, nil)
		return
	}

	req := deffqdn.SetReq{}
	err = c.ShouldBindJSON(&req)
	if err != nil {
		setErr(c, badRequest("the request body is not valid: %v", err), nil)
		return
	}

	name := deffqdn.Normalize(req.Fqdn)
	if name != "" && !deffqdn.IsValid(name) {
		setErr(c, badRequest("%s is not a valid FQDN", name), nil)
		return
	}

	if !fqdnMu.TryLock() {
		setErr(c, errBusy, nil)
		return
	}
	defer fqdnMu.Unlock()

	err = ops.setTuning(name)
	if err != nil {
		log.Errorf("fqdn(%s): committing %q here: %v", reqId(c), name, err)
		setErr(c, err, nil)
		return
	}

	bodies.SetOk(c, fmt.Sprintf("%s committed", base.Hostname), nil)
}

func reloadNodeCertificate(c *gin.Context) {
	err := requireForwarded(c)
	if err != nil {
		setErr(c, err, nil)
		return
	}

	err = ops.reloadCert()
	if err != nil {
		setErr(c, err, nil)
		return
	}

	bodies.SetOk(c, fmt.Sprintf("%s reloaded the certificate", base.Hostname), nil)
}

func restartNodeApi(c *gin.Context) {
	err := requireForwarded(c)
	if err != nil {
		setErr(c, err, nil)
		return
	}

	err = ops.restartApi()
	if err != nil {
		setErr(c, err, nil)
		return
	}

	bodies.SetOk(c, fmt.Sprintf("cube-cos-api restarts on %s", base.Hostname), nil)
}
