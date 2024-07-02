package status

const (
	None = ""

	Create = "create"
	Update = "update"
	Delete = "delete"

	Pending   = "pending"
	Completed = "completed"
	Error     = "error"
)

type Details struct {
	Current string `json:"current" yaml:"current" bson:"current"`
	Desired string `json:"desired,omitempty" yaml:"desired,omitempty" bson:"desired,omitempty"`

	CreatedAt string `json:"createdAt" yaml:"createdAt" bson:"createdAt"`
	UpdatedAt string `json:"updatedAt" yaml:"updatedAt" bson:"updatedAt"`

	MaxPendingDuration int `json:"maxPendingDuration,omitempty" yaml:"maxPendingDuration,omitempty" bson:"maxPendingDuration,omitempty"`
}

func (s *Details) ClearDesired() {
	s.Desired = None
}

func (s *Details) SetCurrentToCompleted() {
	s.Current = Completed
}

func (s *Details) SetCurrentToPending() {
	s.Current = Pending
}

func (s *Details) SetDesiredToUpdate() {
	s.Desired = Update
}

func (s *Details) SetDesiredToDelete() {
	s.Desired = Delete
}
