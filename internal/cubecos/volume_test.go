package cubecos

import "testing"

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
