package firmwares

import "github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"

const (
	Module = "firmwares"

	Db               = "firmwares"
	UploadCollection = "upload"

	TmpUploadDir       = "/tmp/firmwares"
	TmpPreCalculateMd5 = "precalculated.md5"
	DefaultMd5File     = "md5"

	UpdateDir     = "/var/update"
	UpdateHistory = "/var/appliance-db/update.history"

	// Per-node, on the A/B root partition: it does not survive an upgrade.
	// Only the non-rolling per-node update path uses it. Cluster-wide roll
	// progress comes from RollJob / hex_sdk power_roll_status_json.
	UpdateProgress      = "/var/lib/cube-cos-api/progress.json"
	ResolvedMarker      = "/var/lib/cube-cos-api/resolved"
	BootstrappingMarker = "/var/lib/cube-cos-api/bootstrapping"
	BootstrappingLog    = "/run/cube_bootstrap.log"

	// Cluster-wide roll state, on shared cephfs, written by the hex_sdk
	// power_roll_* state machine.
	RollJob = "/mnt/cephfs/rolling/job.json"
)

// Roll job states (job.json .state).
const (
	RollStateNone    = "none"
	RollStateRunning = "running"
	RollStatePaused  = "paused"
	RollStateDone    = "done"
	RollStateAborted = "aborted"
)

// Roll job kinds (job.json .kind).
const (
	RollKindRestart = "restart"
	RollKindUpgrade = "upgrade"
)

type ReqOpts struct {
	Id          string          `json:"id"`
	Hostname    string          `json:"hostname"`
	Version     string          `json:"version"`
	PkgPath     string          `json:"pkgPath"`
	AutoRolling bool            `json:"autoRolling"`
	Status      status.Firmware `json:"status"`
}

type Firmware struct {
	Version      string          `json:"version" bson:"version"`
	ReleaseNotes string          `json:"releaseNotes" bson:"releaseNotes"`
	UpdatedAt    string          `json:"updatedAt" bson:"updatedAt"`
	Status       status.Firmware `json:"status" bson:"status"`
}

type Upadte struct {
	Current  string `yaml:"current"`
	Rollback string `yaml:"rollback"`
	History  []Raw  `yaml:"history"`
}

type Raw struct {
	Image     string `yaml:"image"`
	Type      string `yaml:"type"`
	Version   string `yaml:"version"`
	Variant   string `yaml:"variant"`
	BuiltAt   string `yaml:"built-at"`
	CreatedAt string `yaml:"created-at"`
}

type Upgrade struct {
	Version          string     `json:"version"`
	IsRollingApplied bool       `json:"isRollingApplied"`
	Progresses       []Progress `json:"progresses"`
}

type Progress struct {
	Host   string                      `json:"host"`
	Phase  string                      `json:"phase"`
	Status status.SystemUpdateProgress `json:"status"`
}

// RollStatus is the envelope emitted by "hex_sdk power_roll_status_json".
// Progresses is shaped exactly like Upgrade.Progresses so it can be handed
// to the API response unchanged.
type RollStatus struct {
	IsRollingApplied bool       `json:"isRollingApplied"`
	State            string     `json:"state"`
	Reason           string     `json:"reason"`
	Progresses       []Progress `json:"progresses"`
}

// IsInFlight reports whether a roll is currently running or paused on a failure.
func (r *RollStatus) IsInFlight() bool {
	return r.State == RollStateRunning || r.State == RollStatePaused
}

// Roll is the subset of job.json the API reads directly. power_roll_status_json
// does not project the target package, so the version comes from here.
type Roll struct {
	Kind    string     `json:"kind"`
	Version string     `json:"version"`
	State   string     `json:"state"`
	Reason  string     `json:"reason"`
	Nodes   []RollNode `json:"nodes"`
}

type RollNode struct {
	Hostname string `json:"hostname"`
	Role     string `json:"role"`
	Status   string `json:"status"`
}

// The predicates are nil-safe: no roll has ever run on a fresh cluster.
func (r *Roll) IsInFlight() bool {
	return r != nil && (r.State == RollStateRunning || r.State == RollStatePaused)
}

func (r *Roll) IsUpgrade() bool {
	return r != nil && r.Kind == RollKindUpgrade
}

// AreAllNodesDone reports whether every node in the job reached its final phase.
func (r *Roll) AreAllNodesDone() bool {
	if r == nil || len(r.Nodes) == 0 {
		return false
	}

	for _, node := range r.Nodes {
		if node.Status != "done" {
			return false
		}
	}

	return true
}

type BootstrappingStatus struct {
	Node   string `json:"node"`
	Return string `json:"return"`
	Stdout string `json:"stdout"`
}

type ResolvedStatus struct {
	HasFailureBeenResolved bool `json:"hasFailureBeenResolved"`
}

func (u *ReqOpts) SetInstalling() {
	u.Status.Current = status.Installing
	u.Status.Desired = status.Installed
	u.Status.IsProcessing = true
}

func (u *ReqOpts) SetError(err string) {
	u.Status.Current = status.Error
	u.Status.IsProcessing = false
	u.Status.Description = err
}

func (u *ReqOpts) SetWaitingReboot() {
	u.Status.Current = status.WaitingReboot
	u.Status.IsProcessing = false
}
