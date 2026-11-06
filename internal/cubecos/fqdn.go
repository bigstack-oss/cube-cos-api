package cubecos

import (
	"errors"

	defErrors "github.com/bigstack-oss/cube-cos-api/internal/definition/v1/errors"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/fqdn"
)

// GetFqdn returns cubesys.fqdn, the name the web UIs are served on, or "" when none is set.
func GetFqdn() (string, error) {
	value, err := GetTuningValue(fqdn.Tuning)
	if errors.Is(err, defErrors.ErrTuningNotFound) {
		return "", nil
	}
	if err != nil {
		return "", err
	}

	return value, nil
}
