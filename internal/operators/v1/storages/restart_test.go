package storages

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	ostime "time"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/storages"
)

const fakeHexSdk = `#!/bin/sh
echo "$1" >> "$HEX_LOG"
case " $HEX_FAIL " in *" $1 "*) echo "boom: $1"; exit 1 ;; esac
echo '{}'
`

// restartRig fakes hex_sdk on PATH, the request records in mongo, the
// clock and the result report to the VIP control
type restartRig struct {
	t       *testing.T
	o       *Operator
	log     string
	mu      sync.Mutex
	records map[string]bool // hostname -> record present
	now     ostime.Time
	start   ostime.Time
	sleeps  int
	onSleep map[int]func()
	reports chan storages.ReqOpts
	peerRes map[string]string // peer -> reported status
}

func newRestartRig(t *testing.T, peers ...string) *restartRig {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "hex_sdk"), []byte(fakeHexSdk), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+":"+os.Getenv("PATH"))
	t.Setenv("HEX_LOG", filepath.Join(dir, "calls"))
	t.Setenv("HEX_FAIL", "")

	r := &restartRig{
		t:       t,
		o:       &Operator{},
		log:     filepath.Join(dir, "calls"),
		records: map[string]bool{},
		now:     ostime.Date(2026, 10, 15, 0, 0, 0, 0, ostime.UTC),
		onSleep: map[int]func(){},
		reports: make(chan storages.ReqOpts, 1),
		peerRes: map[string]string{},
	}
	r.start = r.now
	for _, p := range peers {
		r.records[p] = true
	}

	saved, savedReport := restartOps, reportStorage
	t.Cleanup(func() { restartOps, reportStorage = saved, savedReport })

	restartOps.pending = func(name string, hosts []string) ([]string, error) {
		r.mu.Lock()
		defer r.mu.Unlock()
		left := []string{}
		for _, h := range hosts {
			if r.records[h] {
				left = append(left, h)
			}
		}
		return left, nil
	}
	restartOps.now = func() ostime.Time { return r.now }
	restartOps.sleep = func(d ostime.Duration) {
		r.now = r.now.Add(d)
		r.sleeps++
		if f := r.onSleep[r.sleeps]; f != nil {
			f()
		}
	}
	// the VIP control drops a host's record when its result arrives
	reportStorage = func(_ *Operator, req storages.ReqOpts) {
		r.mu.Lock()
		delete(r.records, req.Hostname)
		_, isPeer := r.peerRes[req.Hostname]
		if isPeer {
			r.peerRes[req.Hostname] = req.Status.Current
		}
		r.mu.Unlock()
		if !isPeer {
			r.reports <- req
		}
	}
	for _, p := range peers {
		r.peerRes[p] = ""
	}
	return r
}

func storageReq(host string, desired string, origin bool, peers ...string) *storages.ReqOpts {
	req := &storages.ReqOpts{ReqId: host + "-req", Hostname: host}
	req.CinderDetails.Name = "pure1"
	req.Status.Desired = desired
	req.IsOrigin = origin
	req.Peers = peers
	return req
}

// peerApplies runs a peer's own apply and report, as the peer's operator would
func (r *restartRig) peerApplies(host string, fail bool) func() {
	return func() {
		if fail {
			r.t.Setenv("HEX_FAIL", "cinder_put_storage")
			defer r.t.Setenv("HEX_FAIL", "")
		}
		r.o.processStorageReq(storageReq(host, status.Created, false))
	}
}

func (r *restartRig) result() storages.ReqOpts {
	r.t.Helper()
	select {
	case req := <-r.reports:
		return req
	case <-ostime.After(10 * ostime.Second):
		r.t.Fatal("no result reported")
	}
	return storages.ReqOpts{}
}

func (r *restartRig) calls() []string {
	b, _ := os.ReadFile(r.log)
	return strings.Fields(string(b))
}

func (r *restartRig) wantCalls(want ...string) {
	r.t.Helper()
	if got := r.calls(); strings.Join(got, " ") != strings.Join(want, " ") {
		r.t.Errorf("hex_sdk calls:\n got  %q\n want %q", got, want)
	}
}

func TestRestartOnceAfterEveryPeerReported(t *testing.T) {
	r := newRestartRig(t, "cube46", "cube47")
	r.onSleep[1] = r.peerApplies("cube46", false)
	r.onSleep[3] = r.peerApplies("cube47", false)

	r.o.processStorageReq(storageReq("cube45", status.Created, true, "cube46", "cube47"))
	res := r.result()

	r.wantCalls("cinder_put_storage", "cinder_put_storage", "cinder_put_storage", "cinder_restart_volume")
	if res.Status.Current != status.Completed {
		t.Errorf("origin status = %q, want %q", res.Status.Current, status.Completed)
	}
	if r.sleeps != 3 {
		t.Errorf("polled %d times, want 3", r.sleeps)
	}
}

func TestRestartAfterBoundWhenPeerNeverReports(t *testing.T) {
	r := newRestartRig(t, "cube46", "cube47")
	r.onSleep[1] = r.peerApplies("cube46", false)

	r.o.processStorageReq(storageReq("cube45", status.Deleted, true, "cube46", "cube47"))
	res := r.result()

	r.wantCalls("cinder_delete_storage", "cinder_put_storage", "cinder_restart_volume")
	if waited := r.now.Sub(r.start); waited < peerWaitBound || waited > peerWaitBound+peerPollEvery {
		t.Errorf("restarted after %s, want the %s bound", waited, peerWaitBound)
	}
	if res.Status.Current != status.Completed {
		t.Errorf("origin status = %q, want %q", res.Status.Current, status.Completed)
	}
}

func TestPeerFailureStillRestartsOnce(t *testing.T) {
	r := newRestartRig(t, "cube46", "cube47")
	r.onSleep[1] = r.peerApplies("cube46", false)
	r.onSleep[2] = r.peerApplies("cube47", true)

	r.o.processStorageReq(storageReq("cube45", status.Updated, true, "cube46", "cube47"))
	res := r.result()

	r.wantCalls("cinder_put_storage", "cinder_put_storage", "cinder_put_storage", "cinder_restart_volume")
	if res.Status.Current != status.Completed {
		t.Errorf("origin status = %q, want %q", res.Status.Current, status.Completed)
	}
	if r.peerRes["cube46"] != status.Completed || r.peerRes["cube47"] != status.Failed {
		t.Errorf("peer results = %v, want cube46 completed, cube47 failed", r.peerRes)
	}
}

func TestLocalOrRestartFailureMarksOriginFailed(t *testing.T) {
	for _, fail := range []string{"cinder_put_storage", "cinder_restart_volume"} {
		r := newRestartRig(t)
		t.Setenv("HEX_FAIL", fail)

		r.o.processStorageReq(storageReq("cube45", status.Created, true))
		res := r.result()

		r.wantCalls("cinder_put_storage", "cinder_restart_volume")
		if res.Status.Current != status.Failed {
			t.Errorf("%s fails: origin status = %q, want %q", fail, res.Status.Current, status.Failed)
		}
	}
}

func TestNoRestartWithoutOriginOrBackendChange(t *testing.T) {
	tests := []struct {
		name string
		req  *storages.ReqOpts
		call string
	}{
		{"forwarded create", storageReq("cube46", status.Created, false), "cinder_put_storage"},
		{"forwarded delete", storageReq("cube46", status.Deleted, false), "cinder_delete_storage"},
		{"set as default", storageReq("cube45", status.Defaulted, true), "cinder_set_default_storage"},
	}

	for _, tt := range tests {
		r := newRestartRig(t)
		r.o.processStorageReq(tt.req)
		select {
		case <-r.reports:
		default:
			t.Fatalf("%s: result not reported synchronously", tt.name)
		}
		r.wantCalls(tt.call)
	}
}
