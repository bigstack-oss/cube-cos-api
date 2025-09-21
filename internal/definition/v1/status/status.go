package status

import "time"

const (
	None = ""

	Create = "create"
	Update = "update"
	Delete = "delete"
	Reset  = "reset"

	Creating             = "creating"
	Adding               = "adding"
	Pending              = "pending"
	Updating             = "updating"
	Uploading            = "uploading"
	SettingToDefault     = "setting to default"
	Verifying            = "verifying"
	Importing            = "importing"
	Installing           = "installing"
	Repairing            = "repairing"
	Processing           = "processing"
	Partitioning         = "partitioning"
	Upgrading            = "upgrading"
	Syncing              = "syncing"
	Checking             = "checking"
	CheckDisabled        = "check disabled"
	CheckingAndRepairing = "checkingAndRepairing"
	Deleting             = "deleting"
	Removing             = "removing"
	Expairing            = "expairing"
	Fixing               = "fixing"
	RollingBack          = "rollingBack"
	PoweringOn           = "powering on"
	PoweringOff          = "powering off"
	PoweringCycle        = "powering cycle"
	Restarted            = "restarted"

	Completed  = "completed"
	Created    = "created"
	Imported   = "imported"
	Toggled    = "toggled"
	Added      = "added"
	Uploaded   = "uploaded"
	Upgraded   = "upgraded"
	Installed  = "installed"
	Rollbacked = "rollbacked"
	Removed    = "removed"
	Updated    = "updated"
	Defaulted  = "defaulted"
	Promoted   = "promoted"
	Demoted    = "demoted"
	Reweighted = "reweighted"
	Deleted    = "deleted"
	Failed     = "failed"
	Disabled   = "disabled"
	Ok         = "ok"
	Up         = "up"
	Ng         = "ng"
	Down       = "down"
	Available  = "available"
	InUse      = "in-use"
	System     = "system"

	Valid     = "valid"
	Unlicense = "unlicense"
	Expired   = "expired"
	Unknown   = "unknown"
	Error     = "error"
)

type Health struct {
	Current string `json:"current,omitempty" bson:"current"`
	Desired string `json:"desired,omitempty" bson:"desired"`

	CreatedAt *time.Time `json:"createdAt,omitempty" bson:"createdAt"`
	UpdatedAt *time.Time `json:"updatedAt,omitempty" bson:"updatedAt"`
	IsFixing  bool       `json:"isFixing" bson:"isFixing"`

	Description string `json:"description,omitempty" bson:"description"`
}

type Tuning struct {
	Current string `json:"current,omitempty" bson:"current"`
	Desired string `json:"desired,omitempty" bson:"desired"`

	CreatedAt string `json:"createdAt,omitzero" bson:"createdAt"`
	UpdatedAt string `json:"updatedAt,omitzero" bson:"updatedAt"`

	MaxPendingDuration int  `json:"maxPendingDuration,omitempty" bson:"maxPendingDuration"`
	IsUpdating         bool `json:"isUpdating" bson:"isUpdating"`
}

type Trigger struct {
	Current string `json:"current,omitempty" bson:"current"`
	Desired string `json:"-" bson:"desired"`

	CreatedAt string `json:"-" bson:"createdAt"`
	UpdatedAt string `json:"-" bson:"updatedAt"`

	IsProcessing bool `json:"isProcessing" bson:"isProcessing"`
}

type SupportFile struct {
	Current string `json:"current,omitempty" bson:"current"`
	Desired string `json:"-" bson:"desired"`

	CreatedAt  string `json:"createdAt,omitzero" bson:"createdAt"`
	IsCreating bool   `json:"isCreating" bson:"isCreating"`
}

type License struct {
	Current    string `json:"current,omitempty" bson:"current"`
	IsExpiring bool   `json:"isExpiring" bson:"isExpiring"`
}

type Settings struct {
	Current    string `json:"current,omitempty" bson:"current"`
	Desired    string `json:"-" bson:"desired"`
	CreatedAt  string `json:"-" bson:"createdAt"`
	IsUpdating bool   `json:"isUpdating" bson:"isUpdating"`
}

type BlockDevice struct {
	Current      string `json:"current,omitempty" bson:"current"`
	Desired      string `json:"-" bson:"desired"`
	IsPromotable bool   `json:"isPromotable" bson:"isPromotable"`
	IsDemotable  bool   `json:"isDemotable" bson:"isDemotable"`
	IsProcessing bool   `json:"isProcessing" bson:"isProcessing"`
	Description  string `json:"description,omitempty" bson:"description"`
}

type Osd struct {
	Current      string `json:"current,omitempty" bson:"current"`
	Desired      string `json:"-" bson:"desired"`
	IsProcessing bool   `json:"isProcessing" bson:"isProcessing"`
}

type Image struct {
	Current        string  `json:"current,omitempty" bson:"current"`
	Desired        string  `json:"-" bson:"desired"`
	IsProcessing   bool    `json:"isProcessing" bson:"isProcessing"`
	ProcessPercent float64 `json:"processPercent" bson:"processPercent"`
}

type Volume struct {
	Current        string  `json:"current,omitempty" bson:"current"`
	Desired        string  `json:"-" bson:"desired"`
	IsProcessing   bool    `json:"isProcessing" bson:"isProcessing"`
	ProcessPercent float64 `json:"processPercent" bson:"processPercent"`
}

type Integration struct {
	Current      string `json:"current,omitempty" bson:"current"`
	IsProcessing bool   `json:"isProcessing" bson:"isProcessing"`
}

type Firmware struct {
	Current      string `json:"current,omitempty" bson:"current"`
	Desired      string `json:"-" bson:"desired"`
	IsUpdatable  bool   `json:"isUpdatable" bson:"isUpdatable"`
	IsProcessing bool   `json:"isProcessing" bson:"isProcessing"`
	IsRemovable  bool   `json:"isRemovable" bson:"isRemovable"`
	Description  string `json:"description,omitempty" bson:"description"`
}

type Fixpack struct {
	Current        string `json:"current,omitempty" bson:"current"`
	Desired        string `json:"-" bson:"desired"`
	IsInstallable  bool   `json:"isInstallable" bson:"isInstallable"`
	IsRollbackable bool   `json:"isRollbackable" bson:"isRollbackable"`
	IsProcessing   bool   `json:"isProcessing" bson:"isProcessing"`
	IsRemovable    bool   `json:"isRemovable" bson:"isRemovable"`
	Description    string `json:"description,omitempty" bson:"description"`
}

type Storage struct {
	Current      string `json:"current,omitempty" bson:"current"`
	Desired      string `json:"-" bson:"desired"`
	IsProcessing bool   `json:"isProcessing" bson:"isProcessing"`
	UpdatedAt    string `json:"updatedAt,omitempty" bson:"updatedAt"`
}

type Model struct {
	Current      string `json:"current,omitempty" bson:"current"`
	Desired      string `json:"-" bson:"desired"`
	IsProcessing bool   `json:"isProcessing" bson:"isProcessing"`
	UpdatedAt    string `json:"updatedAt,omitempty" bson:"updatedAt"`
}

type SystemUpdateProgress struct {
	Current        string  `json:"current" bson:"current"`
	IsProcessing   bool    `json:"isProcessing" bson:"isProcessing"`
	ProcessPercent float64 `json:"processPercent" bson:"processPercent"`
	Description    string  `json:"description" bson:"description"`
}

func NewHealthOk() *Health {
	return &Health{Current: Ok}
}

func (h *Health) SetCurrentToRepairing() {
	h.Current = Repairing
}

func (h *Health) SetDesiredToCheckingAndRepairing() {
	h.Desired = CheckingAndRepairing
}

func (h *Health) SetDesiredToRepairing() {
	h.Desired = Repairing
}

func (h *Health) SetCurrentToError(err error) {
	h.Current = Error
	if err != nil {
		h.Description = err.Error()
	}
}

func (s *Settings) SetOk() {
	s.Current = Ok
	s.IsUpdating = false
}
