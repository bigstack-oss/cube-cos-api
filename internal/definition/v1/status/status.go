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
	CheckingAndRepairing = "checkingAndRepairing"
	Deleting             = "deleting"

	Completed = "completed"
	Created   = "created"
	Updated   = "updated"
	Deleted   = "deleted"
	Ok        = "ok"
	Up        = "up"
	Ng        = "ng"
	Down      = "down"

	Valid     = "valid"
	Unlicense = "unlicense"
	Expired   = "expired"

	Error = "error"
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

	IsUpdating bool `json:"isUpdating" bson:"isUpdating"`
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
	Current     string `json:"current,omitempty" bson:"current"`
	Description string `json:"description,omitempty" bson:"description"`
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

func (s *Settings) InitOkStatus() {
	s.Current = Ok
	s.IsUpdating = false
}
