package cubecos

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	ostime "time"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/storages"
	log "go-micro.dev/v5/logger"
)

var runHexSdk = func(ctx context.Context, args ...string) ([]byte, error) {
	return exec.CommandContext(ctx, "hex_sdk", args...).CombinedOutput()
}

func hexSdkTier(timeout ostime.Duration, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	out, err := runHexSdk(ctx, args...)
	output := strings.TrimSpace(string(out))
	if err != nil {
		log.Errorf("storage tiers: hex_sdk %s failed(%v): %s", strings.Join(args, " "), err, output)
		if output == "" {
			output = err.Error()
		}
		return output, fmt.Errorf("%s", output)
	}

	return output, nil
}

// ListStorageTiers returns every device tier with its registry state
func ListStorageTiers() ([]storages.Tier, error) {
	out, err := hexSdkTier(3*ostime.Minute, "ceph_device_tier_names")
	if err != nil {
		return nil, fmt.Errorf("failed to read the device tiers: %s", out)
	}

	return parseStorageTiers(out)
}

func parseStorageTiers(out string) ([]storages.Tier, error) {
	tiers := []storages.Tier{}
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		state, name, found := strings.Cut(line, "|")
		if !found || name == "" {
			return nil, fmt.Errorf("unexpected device tier line %q", line)
		}

		tiers = append(tiers, storages.Tier{Name: name, State: state})
	}

	return tiers, nil
}

// GetStorageTier returns the tier named, or nil if there is none
func GetStorageTier(name string) (*storages.Tier, error) {
	tiers, err := ListStorageTiers()
	if err != nil {
		return nil, err
	}

	for _, t := range tiers {
		if t.Name == name {
			return &t, nil
		}
	}

	return nil, nil
}

// CreateStorageTier builds the tier on the Ceph side and registers it on this node
func CreateStorageTier(name string, osdIds []string) (string, error) {
	args := append([]string{"ceph_device_tier_create", name}, osdIds...)
	return hexSdkTier(15*ostime.Minute, args...)
}

// DeleteStorageTier tears the tier down on the Ceph side and unregisters it on this node
func DeleteStorageTier(name string) (string, error) {
	return hexSdkTier(15*ostime.Minute, "ceph_device_tier_delete", name)
}

// RegisterStorageTier adds the tier to this node's registry and applies Cinder
func RegisterStorageTier(name string) (string, error) {
	return hexSdkTier(10*ostime.Minute, "ceph_device_tier_register", name)
}

// UnregisterStorageTier takes the tier out of this node's registry and applies Cinder
func UnregisterStorageTier(name string) (string, error) {
	return hexSdkTier(10*ostime.Minute, "ceph_device_tier_unregister", name)
}

// IsStorageTierInUse is true when volumes use the tier's volume type, or that cannot be told
func IsStorageTierInUse(name string) bool {
	_, err := hexSdkTier(3*ostime.Minute, "cinder_is_volume_type_in_use", name)
	return err == nil
}

// ForgetStorageTierService drops the tier's cinder-volume service record
func ForgetStorageTierService(name string) error {
	_, err := hexSdkTier(3*ostime.Minute, "cinder_remove_volume_service", "cube@"+name)
	return err
}

// RestartCinderVolume restarts the cinder-volume pacemaker singleton once
func RestartCinderVolume() error {
	_, err := hexSdkTier(5*ostime.Minute, "cinder_restart_volume")
	return err
}
