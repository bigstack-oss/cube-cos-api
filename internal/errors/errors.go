package errors

import "errors"

const (
	Service = "service"
)

const (
	BadRequest          = "bad request"
	InternalServerError = "internal server error"
)

var (
	ServiceNotFound     = errors.New("service not found")
	TuningParamNotFound = errors.New("tuning parameter not found")
	LicensesNotFound    = errors.New("licenses not found")

	DataCenterIsNotReady  = errors.New("data center is not set ready")
	DataCenterIsRepairing = errors.New("data center is repairing")
)

type Template struct {
	Occurred bool
	Type     string
	Msg      string
	Raw      error
}

func (t Template) Error() string {
	return t.Msg
}

func ErrService(err error) Template {
	return Template{
		Occurred: true,
		Type:     Service,
		Msg:      "configuration operation failure",
		Raw:      err,
	}
}

func CombineErrors(errs []error) error {
	if len(errs) == 0 {
		return nil
	}

	errStr := ""
	for _, err := range errs {
		errStr += err.Error() + ", "
	}

	return errors.New(errStr)
}
