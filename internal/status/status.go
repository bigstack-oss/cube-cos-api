package status

const (
	None = ""

	Create = "create"
	Update = "update"
	Delete = "delete"

	Pending              = "pending"
	Repairing            = "repairing"
	CheckingAndRepairing = "checkingAndRepairing"

	Completed = "completed"
	Ok        = "ok"
	Error     = "error"
)

func NewOk() Details {
	return Details{Current: Ok}
}

type Details struct {
	Current string `json:"current,omitempty" bson:"current"`
	Desired string `json:"desired,omitempty" bson:"desired"`

	CreatedAt          string `json:"createdAt,omitempty" bson:"createdAt"`
	UpdatedAt          string `json:"updatedAt,omitempty" bson:"updatedAt"`
	MaxPendingDuration int    `json:"maxPendingDuration,omitempty" bson:"maxPendingDuration"`

	Description string `json:"description,omitempty" bson:"description"`
}

func (s *Details) ClearDesired() {
	s.Desired = None
}

func (s *Details) SetCurrentToCompleted() {
	s.Current = Completed
}

func (s *Details) SetCurrentToOk() {
	s.Current = Ok
}

func (s *Details) SetCurrentToPending() {
	s.Current = Pending
}

func (s *Details) SetCurrentToRepairing() {
	s.Current = Repairing
}

func (s *Details) SetCurrentToCheckingAndRepairing() {
	s.Current = CheckingAndRepairing
}

func (s *Details) SetDesiredToUpdate() {
	s.Desired = Update
}

func (s *Details) SetDesiredToCompleted() {
	s.Desired = Completed
}

func (s *Details) SetDesiredToOk() {
	s.Desired = Ok
}

func (s *Details) SetDesiredToDelete() {
	s.Desired = Delete
}

func (s *Details) SetDesiredToCheckingAndRepairing() {
	s.Desired = CheckingAndRepairing
}

func (s *Details) SetDesiredToRepairing() {
	s.Desired = Repairing
}

func (s *Details) SetCurrentToError(err error) {
	s.Current = Error
	if err != nil {
		s.Description = err.Error()
	}
}
