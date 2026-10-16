package firmwares

import (
	"testing"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/base"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/firmwares"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"
)

func stubRoll(t *testing.T, roll *firmwares.Roll, rs *firmwares.RollStatus) {
	t.Helper()
	origRoll, origStatus := getRoll, getRollStatus
	getRoll = func() (*firmwares.Roll, error) { return roll, nil }
	getRollStatus = func() (*firmwares.RollStatus, error) { return rs, nil }
	t.Cleanup(func() { getRoll, getRollStatus = origRoll, origStatus })
}

func drainingStatus(state string) *firmwares.RollStatus {
	return &firmwares.RollStatus{
		IsRollingApplied: true,
		State:            state,
		Progresses: []firmwares.Progress{
			{Host: "cube47", Phase: "pending", Status: status.SystemUpdateProgress{Current: "pending"}},
			{Host: "cube46", Phase: "draining", Status: status.SystemUpdateProgress{Current: "draining", IsProcessing: true}},
		},
	}
}

func TestRollingProgressProjectsRestartRoll(t *testing.T) {
	base.ActiveFirmwareVersion = "3.1.20"
	stubRoll(t,
		&firmwares.Roll{Kind: firmwares.RollKindRestart, State: firmwares.RollStateRunning},
		drainingStatus(firmwares.RollStateRunning),
	)

	h := &helper{reqId: "test"}
	got, err := h.getRollingProgress()
	if err != nil {
		t.Fatal(err)
	}

	if got.Kind != firmwares.RollKindRestart || !got.IsRollingApplied || got.Version != "3.1.20" {
		t.Fatalf("got %+v, want a rolling restart on the active version", got)
	}

	if len(got.Progresses) != 2 || got.Progresses[0].Host != "cube46" || got.Progresses[0].Phase != "draining" {
		t.Fatalf("got progresses %+v, want cube46 draining first", got.Progresses)
	}
}

func TestRollingProgressProjectsUpgradeRoll(t *testing.T) {
	stubRoll(t,
		&firmwares.Roll{Kind: firmwares.RollKindUpgrade, State: firmwares.RollStatePaused},
		drainingStatus(firmwares.RollStatePaused),
	)

	h := &helper{reqId: "test"}
	for name, get := range map[string]func() (*firmwares.Upgrade, error){
		"rolling":  h.getRollingProgress,
		"firmware": h.getFirmwareUpgradeProgress,
	} {
		got, err := get()
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}

		if got.Kind != firmwares.RollKindUpgrade || len(got.Progresses) != 2 {
			t.Fatalf("%s: got %+v, want the upgrade projection", name, got)
		}
	}
}

func TestIsRollProjected(t *testing.T) {
	cases := []struct {
		name        string
		roll        *firmwares.Roll
		withRestart bool
		want        bool
	}{
		{"no roll", nil, true, false},
		{"restart running", &firmwares.Roll{Kind: "restart", State: "running"}, true, true},
		{"kindless running", &firmwares.Roll{State: "running"}, true, true},
		{"restart hidden from firmware logic", &firmwares.Roll{Kind: "restart", State: "running"}, false, false},
		{"restart done", &firmwares.Roll{Kind: "restart", State: "done"}, true, false},
		{"upgrade paused", &firmwares.Roll{Kind: "upgrade", State: "paused"}, false, true},
		{"upgrade aborted", &firmwares.Roll{Kind: "upgrade", State: "aborted"}, true, false},
	}

	for _, c := range cases {
		if got := isRollProjected(c.roll, c.withRestart); got != c.want {
			t.Errorf("%s: got %v, want %v", c.name, got, c.want)
		}
	}
}
