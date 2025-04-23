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
	ServiceNotFound    = errors.New("service not found")
	TuningNotFound     = errors.New("tuning parameter not found")
	TuningValueInvalid = errors.New("tuning value is invalid")
	LicensesNotFound   = errors.New("licenses not found")

	DataCenterIsNotReady  = errors.New("data center is not set ready")
	DataCenterIsRepairing = errors.New("data center is repairing")

	EmailSenderHostInvalid = errors.New("email sender host is invalid")
	EmailSenderPortInvalid = errors.New("email sender port is invalid")

	LicenseAlreadyExpired     = errors.New("license is already expired")
	LicenseNotInstalled       = errors.New("license is not installed")
	LicenseInvalidHardware    = errors.New("license's hardware serial is not matched with the current system")
	LicenseInvalidSignature   = errors.New("license's signature is invalid")
	LicenseSysytemCompromised = errors.New("license system is compromised")

	SdkExecutionError  = errors.New("sdk execution error")
	UnknownSettingType = errors.New("unknown setting type")

	InvalidListenAddress  = errors.New("invalid listen address")
	InvalidListenPort     = errors.New("invalid listen port")
	InvalidTimeZone       = errors.New("invalid time zone")
	InvalidNodeRole       = errors.New("invalid node role")
	InvalidHostname       = errors.New("invalid hostname")
	InvalidDataCenterName = errors.New("invalid data center name")
	InvalidManagementIp   = errors.New("invalid management ip")
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
