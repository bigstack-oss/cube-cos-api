package cubecos

import (
	"encoding/json"
	"testing"
)

func TestMovePreflightStatus(t *testing.T) {
	cases := []struct {
		code string
		want int
	}{
		{"OK", 200},
		{"E_HAS_SNAPSHOTS", 409},
		{"E_VM_NOT_RUNNING", 409},
		{"E_NO_SUCH_TYPE", 400},
		{"E_SAME_TYPE", 400},
		{"E_SAME_CLUSTER_SAME_POOL", 409},
	}
	for _, c := range cases {
		if got := MovePreflightStatus(c.code); got != c.want {
			t.Errorf("MovePreflightStatus(%q) = %d, want %d", c.code, got, c.want)
		}
	}
}

func TestMoveDispatchStatus(t *testing.T) {
	cases := []struct {
		code string
		want int
	}{
		{"OK", 202},
		{"E_HAS_SNAPSHOTS", 409},
		{"E_VM_NOT_RUNNING", 409},
		{"E_NO_SUCH_TYPE", 400},
		{"E_SAME_TYPE", 400},
		{"E_SAME_CLUSTER_SAME_POOL", 409},
		{"E_DISPATCH_FAILED", 500},
	}
	for _, c := range cases {
		if got := MoveDispatchStatus(c.code); got != c.want {
			t.Errorf("MoveDispatchStatus(%q) = %d, want %d", c.code, got, c.want)
		}
	}
}

// TestMoveDispatchGoldenSamples pins the wire contract between
// cinder_move_volume's actual stdout and the Go side that parses it: byte
// samples of what the shell emits, unmarshalled through the exact same
// snake_case-raw-then-convert path RunMoveVolume uses (rawMoveDispatch ->
// MoveDispatch), then mapped to an HTTP status. This is the seam that
// TestMoveDispatchStatus cannot cover, because it feeds MoveDispatchStatus
// the literal string "OK" rather than going through JSON at all — that gap
// is exactly how a success response with no "code" key shipped as a false
// 409 instead of 202.
func TestMoveDispatchGoldenSamples(t *testing.T) {
	cases := []struct {
		name       string
		json       string
		wantStatus int
		check      func(t *testing.T, md *MoveDispatch)
	}{
		{
			name:       "successful dispatch",
			json:       `{"ok":true,"code":"OK","reason":"ok","dispatched":true}`,
			wantStatus: 202,
		},
		{
			name:       "dispatch failure",
			json:       `{"ok":false,"code":"E_DISPATCH_FAILED","reason":"cinder retype command failed"}`,
			wantStatus: 500,
		},
		{
			name: "refusal carries every blocker and every preflight field",
			json: `{"ok":false,"code":"E_HAS_SNAPSHOTS","reason":"volume must not have snapshots",` +
				`"blockers":[{"code":"E_HAS_SNAPSHOTS","reason":"volume must not have snapshots"},` +
				`{"code":"E_VM_NOT_RUNNING","reason":"the attached instance must be running or paused"}],` +
				`"src_type":"ceph","dst_type":"tier-nvme","size_gb":100,"attached_to":"vm-1"}`,
			wantStatus: 409,
			check: func(t *testing.T, md *MoveDispatch) {
				if len(md.Blockers) != 2 {
					t.Fatalf("got %d blockers, want 2: %+v", len(md.Blockers), md.Blockers)
				}
				if md.Blockers[0].Code != "E_HAS_SNAPSHOTS" || md.Blockers[1].Code != "E_VM_NOT_RUNNING" {
					t.Errorf("blockers = %+v, want E_HAS_SNAPSHOTS then E_VM_NOT_RUNNING", md.Blockers)
				}
				if md.SrcType != "ceph" {
					t.Errorf("SrcType = %q, want ceph", md.SrcType)
				}
				if md.DstType != "tier-nvme" {
					t.Errorf("DstType = %q, want tier-nvme", md.DstType)
				}
				if md.SizeGb != 100 {
					t.Errorf("SizeGb = %d, want 100", md.SizeGb)
				}
				if md.AttachedTo != "vm-1" {
					t.Errorf("AttachedTo = %q, want vm-1", md.AttachedTo)
				}
			},
		},
		{
			// The bug that actually shipped: a success payload missing the
			// "code" key entirely must never be reported as success. Today
			// it falls through MoveDispatchStatus -> MovePreflightStatus("")
			// -> the 409 default. Pinned explicitly so a future emitter
			// change that drops the key again fails loudly here, not in
			// production.
			name:       "missing code key is never reported as success",
			json:       `{"ok":true,"dispatched":true}`,
			wantStatus: 409,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var raw rawMoveDispatch
			if err := json.Unmarshal([]byte(c.json), &raw); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			md := MoveDispatch(raw)

			if got := MoveDispatchStatus(md.Code); got != c.wantStatus {
				t.Errorf("MoveDispatchStatus(%q) = %d, want %d", md.Code, got, c.wantStatus)
			}
			if c.check != nil {
				c.check(t, &md)
			}
		})
	}
}
