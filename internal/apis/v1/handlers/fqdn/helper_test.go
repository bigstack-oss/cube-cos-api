package fqdn

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/base"
	deffqdn "github.com/bigstack-oss/cube-cos-api/internal/definition/v1/fqdn"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"
	"github.com/gin-gonic/gin"
)

type rig struct {
	t       *testing.T
	events  []string
	failOn  map[string]error
	nodes   []nodes.Node
	fqdn    string
	restore func()
}

func newRig(t *testing.T) *rig {
	r := &rig{t: t, failOn: map[string]error{}}
	saved := ops
	host, role := base.Hostname, base.CurrentRole
	base.Hostname, base.CurrentRole = "cc1", nodes.RoleControl

	ev := func(e string) error {
		r.events = append(r.events, e)
		return r.failOn[e]
	}
	ops.getNode = func() (*deffqdn.Node, error) {
		return &deffqdn.Node{Hostname: "cc1", Fqdn: r.fqdn, PublicHost: "10.1.0.100"}, nil
	}
	ops.getFqdn = func() (string, error) { return r.fqdn, nil }
	ops.verifyCert = func(name string) error { return ev("verify " + name) }
	ops.installCert = func(dir string) (*deffqdn.Certificate, error) {
		return &deffqdn.Certificate{Fingerprint: "AA"}, ev("install")
	}
	ops.reloadCert = func() error { return ev("reload cc1") }
	ops.setTuning = func(name string) error { return ev("tuning cc1 " + name) }
	ops.applyCluster = func(certOnly bool) (string, error) {
		if certOnly {
			return "certs", ev("cluster cert")
		}
		return "cluster", ev("cluster")
	}
	ops.restartApi = func() error { return ev("restart cc1") }
	ops.listNodes = func() []nodes.Node { return r.nodes }
	ops.sendToPeer = func(n nodes.Node, method, url string, body any) ([]byte, error) {
		e := method + " " + n.Hostname + " " + url[strings.Index(url, "/fqdn"):]
		if req, ok := body.(deffqdn.SetReq); ok {
			e += " " + req.Fqdn
		}
		if err := ev(e); err != nil {
			return nil, err
		}
		return []byte(`{"data":{"hostname":"` + n.Hostname + `"}}`), nil
	}

	stage := certStageDir
	certStageDir = t.TempDir()

	r.restore = func() { ops = saved; base.Hostname, base.CurrentRole = host, role; certStageDir = stage }
	t.Cleanup(r.restore)
	return r
}

func (r *rig) node(name, role, st string) {
	r.nodes = append(r.nodes, nodes.Node{Hostname: name, Role: role, Status: st, Protocol: "http", Address: name + ":8082"})
}

func (r *rig) call(h gin.HandlerFunc, method, body string, node, forwarded bool) (int, map[string]any) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(method, "/api/v1/datacenters/dc/fqdn", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	if node {
		c.Request.Header.Set("Node", "cc9")
		c.Request.Header.Set("Authorization", "Bearer "+nodes.GenToken("cc9"))
	}
	if forwarded {
		c.Request.Header.Set(nodes.ForwardedHeader, "cc9")
	}
	h(c)

	resp := map[string]any{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	return w.Code, resp
}

func (r *rig) want(events ...string) {
	r.t.Helper()
	if strings.Join(r.events, "|") != strings.Join(events, "|") {
		r.t.Errorf("events:\n got  %q\n want %q", r.events, events)
	}
}

func TestSetFqdnNeedsNodeToken(t *testing.T) {
	r := newRig(t)
	code, _ := r.call(setFqdn, http.MethodPut, `{"fqdn":"cos.example.test"}`, false, false)
	if code != http.StatusForbidden {
		t.Errorf("code %d, want 403", code)
	}
	r.want()
}

func TestSetFqdnRejectsBadNames(t *testing.T) {
	for _, name := range []string{"10.1.0.100", "nodot", "-a.example.test", "a..test", strings.Repeat("a", 250) + ".test"} {
		r := newRig(t)
		code, _ := r.call(setFqdn, http.MethodPut, `{"fqdn":"`+name+`"}`, true, false)
		if code != http.StatusBadRequest {
			t.Errorf("%s: code %d, want 400", name, code)
		}
		r.want()
	}
}

func TestSetFqdnRefusedOnCompute(t *testing.T) {
	r := newRig(t)
	base.CurrentRole = nodes.RoleCompute
	code, _ := r.call(setFqdn, http.MethodPut, `{"fqdn":"cos.example.test"}`, true, false)
	if code != http.StatusBadRequest {
		t.Errorf("code %d, want 400", code)
	}
	r.want()
}

func TestSetFqdnNeedsCertificateCoveringIt(t *testing.T) {
	r := newRig(t)
	r.node("cc2", nodes.RoleControl, status.Up)
	r.failOn["verify cos.example.test"] = errors.New("the certificate does not cover cos.example.test")
	code, resp := r.call(setFqdn, http.MethodPut, `{"fqdn":"Cos.Example.Test "}`, true, false)
	if code != http.StatusBadRequest || !strings.Contains(resp["msg"].(string), "does not cover") {
		t.Errorf("code %d msg %v, want 400 naming the certificate", code, resp["msg"])
	}
	r.want("verify cos.example.test")
}

func TestSetFqdnRefusesWhileANodeIsDown(t *testing.T) {
	r := newRig(t)
	r.node("cc2", nodes.RoleControl, status.Up)
	r.node("cp1", nodes.RoleCompute, status.Down)
	code, _ := r.call(setFqdn, http.MethodPut, `{"fqdn":"cos.example.test"}`, true, false)
	if code != http.StatusConflict {
		t.Errorf("code %d, want 409", code)
	}
	r.want("verify cos.example.test")
}

func TestSetFqdnSwitchesEveryNodeThenTheCluster(t *testing.T) {
	r := newRig(t)
	r.node("cp1", nodes.RoleCompute, status.Up)
	r.node("cc2", nodes.RoleControl, status.Up)
	r.node("cc1", nodes.RoleControl, status.Up) // this node, as the node list has it
	r.node("cc3", nodes.RoleControlConverged, status.Up)
	code, resp := r.call(setFqdn, http.MethodPut, `{"fqdn":"cos.example.test"}`, true, false)
	if code != http.StatusOK {
		t.Fatalf("code %d (%v), want 200", code, resp["msg"])
	}
	r.want(
		"verify cos.example.test",
		"tuning cc1 cos.example.test",
		"PUT cc2 /fqdn/node cos.example.test",
		"PUT cc3 /fqdn/node cos.example.test",
		"PUT cp1 /fqdn/node cos.example.test",
		"cluster",
		"PUT cc2 /fqdn/node/api",
		"PUT cc3 /fqdn/node/api",
		"PUT cp1 /fqdn/node/api",
		"restart cc1",
	)
}

func TestSetFqdnStopsAtTheFirstNodeThatFails(t *testing.T) {
	r := newRig(t)
	r.node("cc2", nodes.RoleControl, status.Up)
	r.node("cc3", nodes.RoleControl, status.Up)
	r.failOn["PUT cc2 /fqdn/node cos.example.test"] = errors.New("commit failed")
	code, resp := r.call(setFqdn, http.MethodPut, `{"fqdn":"cos.example.test"}`, true, false)
	if code != http.StatusInternalServerError || !strings.Contains(resp["msg"].(string), "run the command again") {
		t.Errorf("code %d msg %v, want 500 asking for a re-run", code, resp["msg"])
	}
	r.want("verify cos.example.test", "tuning cc1 cos.example.test", "PUT cc2 /fqdn/node cos.example.test")
}

func TestSetFqdnReportsTheClusterPart(t *testing.T) {
	r := newRig(t)
	r.failOn["cluster"] = errors.New("the cluster-wide part failed: rancher")
	code, resp := r.call(setFqdn, http.MethodPut, `{"fqdn":"cos.example.test"}`, true, false)
	data, _ := resp["data"].(map[string]any)
	if code != http.StatusInternalServerError || data == nil || !strings.Contains(data["output"].(string), "cluster") {
		t.Errorf("code %d data %v, want 500 carrying the step log", code, resp["data"])
	}
	r.want("verify cos.example.test", "tuning cc1 cos.example.test", "cluster")
}

func TestResetFqdnSkipsTheCertificate(t *testing.T) {
	r := newRig(t)
	r.node("cc2", nodes.RoleControl, status.Up)
	code, _ := r.call(resetFqdn, http.MethodDelete, ``, true, false)
	if code != http.StatusOK {
		t.Fatalf("code %d, want 200", code)
	}
	r.want("tuning cc1 ", "PUT cc2 /fqdn/node ", "cluster", "PUT cc2 /fqdn/node/api", "restart cc1")
}

func TestInstallCertificateNeedsAllThree(t *testing.T) {
	r := newRig(t)
	code, _ := r.call(installCertificate, http.MethodPut, `{"certificate":"c","privateKey":"k"}`, true, false)
	if code != http.StatusBadRequest {
		t.Errorf("code %d, want 400", code)
	}
	r.want()
}

func TestInstallCertificateWithoutFqdnServesNothingYet(t *testing.T) {
	r := newRig(t)
	r.node("cc2", nodes.RoleControl, status.Up)
	code, resp := r.call(installCertificate, http.MethodPut, `{"certificate":"c","privateKey":"k","caCertificate":"a"}`, true, false)
	if code != http.StatusOK {
		t.Fatalf("code %d (%v), want 200", code, resp["msg"])
	}
	r.want("install")
}

func TestInstallCertificateReplacesTheOneInUse(t *testing.T) {
	r := newRig(t)
	r.fqdn = "cos.example.test"
	r.node("cc2", nodes.RoleControl, status.Up)
	r.node("cp1", nodes.RoleCompute, status.Up)
	code, resp := r.call(installCertificate, http.MethodPut, `{"certificate":"c","privateKey":"k","caCertificate":"a"}`, true, false)
	if code != http.StatusOK {
		t.Fatalf("code %d (%v), want 200", code, resp["msg"])
	}
	r.want("install", "reload cc1", "PUT cc2 /fqdn/node/certificate", "cluster cert")
}

func TestInstallCertificateRejectsWhatTheSdkRejects(t *testing.T) {
	r := newRig(t)
	r.failOn["install"] = errors.New("the private key does not belong to the certificate")
	code, resp := r.call(installCertificate, http.MethodPut, `{"certificate":"c","privateKey":"k","caCertificate":"a"}`, true, false)
	if code != http.StatusBadRequest || !strings.Contains(resp["msg"].(string), "private key") {
		t.Errorf("code %d msg %v, want 400 with the reason", code, resp["msg"])
	}
}

func TestNodeEndpointsNeedAForwardedNodeRequest(t *testing.T) {
	for name, h := range map[string]gin.HandlerFunc{
		"get": getNode, "set": setNode, "reload": reloadNodeCertificate, "restart": restartNodeApi,
	} {
		r := newRig(t)
		code, _ := r.call(h, http.MethodPut, `{"fqdn":"cos.example.test"}`, true, false)
		if code != http.StatusForbidden {
			t.Errorf("%s: code %d, want 403", name, code)
		}
		r.want()
	}
}

func TestSetNodeCommitsHere(t *testing.T) {
	r := newRig(t)
	code, _ := r.call(setNode, http.MethodPut, `{"fqdn":"cos.example.test"}`, true, true)
	if code != http.StatusOK {
		t.Errorf("code %d, want 200", code)
	}
	r.want("tuning cc1 cos.example.test")
}
