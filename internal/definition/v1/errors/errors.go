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
	ErrServiceNotFound              = errors.New("service not found")
	ErrTuningNotFound               = errors.New("tuning parameter not found")
	ErrTuningValueInvalid           = errors.New("tuning value is invalid")
	ErrLicensesNotFound             = errors.New("licenses not found")
	ErrLicenseUnknownStatus         = errors.New("unknown license status")
	ErrLicenseInternalSystemFailure = errors.New("internal license system error")
	ErrDataCenterIsNotReady         = errors.New("data center is not set ready")
	ErrDataCenterIsRepairing        = errors.New("data center is repairing")
	ErrEmailSenderHostInvalid       = errors.New("email sender host is invalid")
	ErrEmailSenderPortInvalid       = errors.New("email sender port is invalid")
	ErrLicenseAlreadyExpired        = errors.New("license is already expired")
	ErrLicenseNotInstalled          = errors.New("license is not installed")
	ErrLicenseInvalidHardware       = errors.New("license's hardware serial is not matched with the current system")
	ErrLicenseInvalidSignature      = errors.New("license's signature is invalid")
	ErrLicenseSystemCompromised     = errors.New("license system is compromised")
	ErrSdkExecutionFailure          = errors.New("sdk execution error")
	ErrUnknownSettingType           = errors.New("unknown setting type")
	ErrInvalidListenAddress         = errors.New("invalid listen address")
	ErrInvalidListenPort            = errors.New("invalid listen port")
	ErrInvalidTimeZone              = errors.New("invalid time zone")
	ErrInvalidNodeRole              = errors.New("invalid node role")
	ErrInvalidHostname              = errors.New("invalid hostname")
	ErrInvalidDataCenterName        = errors.New("invalid data center name")
	ErrInvalidManagementIp          = errors.New("invalid management ip")
	ErrNoRedirectPathFound          = errors.New("no redirect path found")
	ErrAlertSettingNotInited        = errors.New("alert setting is not initialized")
	ErrAlertSettingTaskTypeIsEmpty  = errors.New("alert setting task type is empty")
	ErrEmailSenderHostIsEmpty       = errors.New("email sender host is empty")
	ErrEmailRecipientNotFound       = errors.New("email recipient not found")
	ErrEmailRecipientIsEmpty        = errors.New("email recipient is empty")
	ErrSlackChannelNameIsEmpty      = errors.New("slack channel name is empty")
	ErrSessionIndexNotFound         = errors.New("session index not found in jwt session")
	ErrAuthMethodCannotGetUserInfo  = errors.New("authed method not support to fetch user info")
	ErrMetricTypeInvalid            = errors.New("metricType should be cpuUsage, memoryUsage, diskUsage, diskBandwidth, diskIops, diskLatency, diskReadIops, diskWriteIops, networkTrafficIn, or networkTrafficOut")
	ErrViewTypeInvalid              = errors.New("viewType should be summary, history, or rank")
	ErrEntityTypeInvalid            = errors.New("entityType should be hosts or vms")
	ErrLimitInvalid                 = errors.New("limit should be an integer and greater than 0")
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

func Is(target, compare error) bool {
	return errors.Is(target, compare)
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
