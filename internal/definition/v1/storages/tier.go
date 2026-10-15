package storages

// Device tier states, as hex_sdk ceph_device_tier_names reports them
const (
	TierRegistered      = "registered"
	TierRegistryPartial = "registry-partial"
	TierRegistryOnly    = "registry-only"
	TierUnmanaged       = "unmanaged"
)

type Tier struct {
	Name  string `json:"name"`
	State string `json:"state"`
}

type TierReqOpts struct {
	Name   string   `json:"name"`
	OsdIds []string `json:"osdIds"`
}

type TierResult struct {
	Name   string `json:"name"`
	Output string `json:"output"`
}

// IsOwned is true for a tier the storage tier registry lists
func (t Tier) IsOwned() bool {
	switch t.State {
	case TierRegistered, TierRegistryPartial, TierRegistryOnly:
		return true
	}

	return false
}
