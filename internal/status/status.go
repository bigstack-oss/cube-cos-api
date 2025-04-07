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

	Completed = "completed"
	Updated   = "updated"
	Ok        = "ok"
	Ng        = "ng"
	Error     = "error"
)

type Details struct {
	Current string `json:"current,omitempty" bson:"current"`
	Desired string `json:"desired,omitempty" bson:"desired"`

	CreatedAt time.Time `json:"createdAt,omitzero" bson:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt,omitzero" bson:"updatedAt"`
	IsFixing  bool      `json:"isFixing" bson:"isFixing"`

	Description string `json:"description,omitempty" bson:"description"`
}

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
	Desired string `json:"-,omitempty" bson:"desired"`

	CreatedAt  string `json:"createdAt,omitzero" bson:"createdAt"`
	IsCreating bool   `json:"isCreating" bson:"isCreating"`
}

type License struct {
	Current    string `json:"current,omitempty" bson:"current"`
	IsExpiring bool   `json:"isExpiring" bson:"isExpiring"`
}

func NewHealthOk() *Health {
	return &Health{Current: Ok}
}

func (d *Details) ClearDesired() {
	d.Desired = None
}

func (d *Details) SetCurrentToCompleted() {
	d.Current = Completed
}

func (d *Details) SetCurrentToOk() {
	d.Current = Ok
}

func (d *Details) SetCurrentToPending() {
	d.Current = Pending
}

func (h *Health) SetCurrentToRepairing() {
	h.Current = Repairing
}

func (d *Details) SetCurrentToCheckingAndRepairing() {
	d.Current = CheckingAndRepairing
}

func (d *Details) SetDesiredToUpdate() {
	d.Desired = Update
}

func (d *Details) SetDesiredToCompleted() {
	d.Desired = Completed
}

func (d *Details) SetDesiredToOk() {
	d.Desired = Ok
}

func (d *Details) SetDesiredToDelete() {
	d.Desired = Delete
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
