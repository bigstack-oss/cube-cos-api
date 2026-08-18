package gpu

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
)

type ResourceType string
type SupportResourceType string
type GpuStatus string

const (
	ResourceTypeUnset         ResourceType = "unset"
	ResourceTypePgpu          ResourceType = "pgpu"
	ResourceTypeSriovVgpu     ResourceType = "sriovVgpu"
	ResourceTypeMigBackedVgpu ResourceType = "migBackedVgpu"

	SupportResourceTypePgpu          SupportResourceType = "pgpu"
	SupportResourceTypeSriovVgpu     SupportResourceType = "sriovVgpu"
	SupportResourceTypeMigBackedVgpu SupportResourceType = "migBackedVgpu"

	GpuStatusUnassigned GpuStatus = "unassigned"
	GpuStatusIdle       GpuStatus = "idle"
	GpuStatusInUse      GpuStatus = "inUse"
)

type GpuFromHex struct {
	Id                         string                `json:"id"`
	Name                       string                `json:"name"`
	Type                       ResourceType          `json:"type"`
	SupportTypes               []SupportResourceType `json:"supportTypes"`
	PciAddress                 string                `json:"pciAddress"`
	SriovVgpuProfileCountLimit *int                  `json:"sriovVgpuProfileCountLimit"`
	Status                     GpuStatus             `json:"status"`
	Allocation                 *AllocationSummary    `json:"allocation"`
}

type VgpuProfileCollectionFromHex struct {
	Sriov     *[]VgpuProfileFromHex `json:"sriov"`
	MigBacked *[]VgpuProfileFromHex `json:"migBacked"`
}

// MiB is a whole-MiB quantity decoded from a hex_sdk JSON number that may be
// fractional. `hex_sdk gpu_vgpu_profile_list` derives a MIG-backed profile's
// vramMiB from the GiB column of `nvidia-smi mig -lgip` (GiB * 1024), so it can
// emit e.g. 23674.88 for a 23.12 GiB "1g.24gb" profile. Decoding that into a
// plain uint64 fails, and because encoding/json aborts the whole document on a
// single field error, one fractional profile drops *every* profile on the card.
// Rounding at the decode boundary keeps every caller working in whole MiB.
type MiB uint64

func (m *MiB) UnmarshalJSON(data []byte) error {
	// encoding/json calls UnmarshalJSON for a JSON null too; leave the value
	// untouched, as it would for a plain numeric field.
	if string(data) == "null" {
		return nil
	}

	var f float64
	if err := json.Unmarshal(data, &f); err != nil {
		return err
	}

	if math.IsNaN(f) || math.IsInf(f, 0) || f < 0 || f > math.MaxUint64 {
		return fmt.Errorf("vramMiB out of range: %v", f)
	}

	*m = MiB(math.Round(f))

	return nil
}

type VgpuProfileFromHex struct {
	Id           uint32  `json:"id"`
	Name         string  `json:"name"`
	VramMiB      MiB     `json:"vramMiB"`
	Count        int     `json:"count"`
	Alias        *string `json:"alias"`
	VmCountLimit *int    `json:"vmCountLimit"`
}

type PgpuAttachedInstanceFromHex struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type GpuCard struct {
	Id                         string                `json:"id"`
	Name                       string                `json:"name"`
	PciAddress                 string                `json:"pciAddress"`
	ResourceType               ResourceType          `json:"resourceType"`
	SupportResourceTypes       []SupportResourceType `json:"supportResourceTypes"`
	Vram                       VramInfo              `json:"vram"`
	Gpu                        GpuInfo               `json:"gpu"`
	AllocationSummary          *AllocationSummary    `json:"allocationSummary"`
	SriovVgpuProfileCountLimit *int                  `json:"sriovVgpuProfileCountLimit"`
	Profiles                   GpuProfileCollection  `json:"profiles"`
	AttachedInstances          *[]AttachedInstance   `json:"attachedInstances"`
	Links                      GpuCardLinks          `json:"links"`
	Status                     GpuStatusInfo         `json:"status"`
	// Degraded is true when runtime enrichment that this card should have
	// had (stats, attached instances) could not be obtained, so its capacity
	// numbers are not trustworthy. Consumers such as schedulers should not
	// allocate off a degraded card's reported capacity.
	Degraded bool `json:"degraded"`
}

// A nil field means the number does not exist for this card, not that it is
// zero. Two situations produce it and neither is a fault, so neither sets
// Degraded: a card handed to a VM as a pgpu is bound to vfio-pci and invisible
// to nvidia-smi, so it has no runtime stats at all; and a card with MIG enabled
// reports its framebuffer but `N/A` for utilization, because NVIDIA provides no
// device-level utilization once the card is partitioned (per-GPU-instance
// numbers would need DCGM). Reporting 0 instead would be indistinguishable from
// an idle card.
type VramInfo struct {
	AllocatedMiB       *int    `json:"allocatedMiB"`
	TotalMiB           *int    `json:"totalMiB"`
	UtilizationPercent *uint32 `json:"utilizationPercent"`
}

type GpuInfo struct {
	UtilizationPercent *uint32 `json:"utilizationPercent"`
}

type AllocationSummary struct {
	Current int `json:"current"`
	Total   int `json:"total"`
}

type GpuProfileCollection struct {
	SriovVgpu     []VgpuProfile `json:"sriovVgpu"`
	MigBackedVgpu []VgpuProfile `json:"migBackedVgpu"`
}

type VgpuProfile struct {
	Id         uint32  `json:"id"`
	Name       string  `json:"name"`
	VramMiB    uint64  `json:"vramMiB"`
	Count      int     `json:"count"`
	Remaining  *int    `json:"remaining"`
	AliasName  *string `json:"aliasName"`
	CountLimit *int    `json:"countLimit"`
}

// GpuCardLinks are the history charts for this one card. They are reported
// inline with the card because the UI offers them from its row, and a link built
// per node would be identical on every row of the same node.
type GpuCardLinks struct {
	WorkloadHistory string `json:"workloadHistory"`
	VramHistory     string `json:"vramHistory"`
}

// UtilizationPercent and MemoryUsage are nil for the same reasons as VramInfo's
// fields: a pgpu's card is invisible to the host, and a MIG-backed vGPU reports
// framebuffer but no utilization.
type AttachedInstance struct {
	Id                 string              `json:"id"`
	Name               string              `json:"name"`
	ProfileAlias       *string             `json:"profileAlias"`
	UtilizationPercent *uint32             `json:"utilizationPercent"`
	MemoryUsage        InstanceMemoryUsage `json:"memoryUsage"`
	Links              InstanceLinks       `json:"links"`
}

type InstanceMemoryUsage struct {
	AllocatedMiB *int `json:"allocatedMiB"`
	TotalMiB     *int `json:"totalMiB"`
}

type InstanceLinks struct {
	// Grafana is nil when the dashboard variables cannot be pinned. The instance
	// dashboard resolves its tenant, hostname and instance variables on load, so a
	// link built without the owning project can label this VM's chart with another
	// VM -- and a GPU passed through to a VM has no vGPU series to chart anyway.
	// Same rule as the stats above: what cannot be produced correctly is absent,
	// not zeroed.
	Grafana *string `json:"grafana"`
}

// InstanceConsole is the on-demand console link for an attached instance. It is
// minted lazily via its own endpoint rather than inline with the GPU listing so
// the list path does not create a Nova console token per instance per poll.
type InstanceConsole struct {
	Console string `json:"console"`
}

type GpuStatusInfo struct {
	Current      GpuStatus `json:"current"`
	IsProcessing bool      `json:"isProcessing"`
}

// Sentinel errors for the update-GPU-card flow. Wrap with %w so the HTTP
// handler can map them to status codes via errors.Is.
var (
	ErrGpuNotFound             = errors.New("gpu not found")
	ErrUnsupportedType         = errors.New("resource type not supported by gpu")
	ErrProfilesNotAllowed      = errors.New("profiles not allowed for this resource type")
	ErrProfileNotFound         = errors.New("vgpu profile not found")
	ErrExceedProfileCountLimit = errors.New("exceed gpu profile count limit")
	ErrExceedVramLimit         = errors.New("exceed vram limit MiB")
	ErrGpuInUse                = errors.New("gpu is in use")
	ErrInvalidProfileCount     = errors.New("invalid vgpu profile count")
	ErrDuplicateProfile        = errors.New("duplicate vgpu profile")
)

type UpdateGpuCardProfile struct {
	Id    uint32 `json:"id" bson:"id"`
	Count int    `json:"count" bson:"count"`
}

type UpdateGpuCardRequest struct {
	ResourceType ResourceType           `json:"resourceType" binding:"required"`
	Profiles     []UpdateGpuCardProfile `json:"profiles"`
}

// GpuReqRecord marks a GPU as being reconfigured. Its presence in
// ReqGpuCollection makes the list endpoint report Status.IsProcessing = true.
type GpuReqRecord struct {
	Hostname     string                 `json:"hostname" bson:"hostname"`
	GpuId        string                 `json:"gpuId" bson:"gpuId"`
	ReqId        string                 `json:"reqId" bson:"reqId"`
	ResourceType ResourceType           `json:"resourceType" bson:"resourceType"`
	Profiles     []UpdateGpuCardProfile `json:"profiles" bson:"profiles"`
}
