package status

import "time"

const (
	None = ""

	Create = "create"
	Update = "update"
	Delete = "delete"
	Reset  = "reset"

	Creating             = "creating"
	Pending              = "pending"
	Updating             = "updating"
	Repairing            = "repairing"
	Syncing              = "syncing"
	CheckingAndRepairing = "checkingAndRepairing"
	Deleting             = "deleting"
	Expairing            = "expairing"
	Fixing               = "fixing"
	PoweringOn           = "powering on"
	PoweringOff          = "powering off"
	PoweringCycle        = "powering cycle"

	Completed = "completed"
	Created   = "created"
	Updated   = "updated"
	Deleted   = "deleted"
	Failed    = "failed"
	Ok        = "ok"
	Up        = "up"
	Ng        = "ng"
	Down      = "down"
	InUse     = "in-use"
	System    = "system"

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
	Desired string `json:"desired,omitempty" bson:"desired"`

	CreatedAt string `json:"createdAt,omitzero" bson:"createdAt"`
	UpdatedAt string `json:"updatedAt,omitzero" bson:"updatedAt"`

	IsCreating bool `json:"isCreating" bson:"isCreating"`
	IsUpdating bool `json:"isUpdating" bson:"isUpdating"`
	IsDeleting bool `json:"isDeleting" bson:"isDeleting"`
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
	IsPromotable bool   `json:"isPromoted" bson:"isPromoted"`
	IsDemotable  bool   `json:"isDemoted" bson:"isDemoted"`
	Description  string `json:"description,omitempty" bson:"description"`
}

type Osd struct {
	Current string `json:"current,omitempty" bson:"current"`
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
