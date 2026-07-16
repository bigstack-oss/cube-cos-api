package gpu

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

type VgpuProfileFromHex struct {
	Id           uint32  `json:"id"`
	Name         string  `json:"name"`
	VramMiB      uint64  `json:"vramMiB"`
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
	Status                     GpuStatusInfo         `json:"status"`
	// Degraded is true when NVML runtime enrichment that this card should have
	// had (stats, attached instances) could not be obtained, so its capacity
	// numbers are not trustworthy. Consumers such as schedulers should not
	// allocate off a degraded card's reported capacity.
	Degraded bool `json:"degraded"`
}

type VramInfo struct {
	AllocatedMiB       int    `json:"allocatedMiB"`
	TotalMiB           int    `json:"totalMiB"`
	UtilizationPercent uint32 `json:"utilizationPercent"`
}

type GpuInfo struct {
	UtilizationPercent uint32 `json:"utilizationPercent"`
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

type AttachedInstance struct {
	Id                 string              `json:"id"`
	Name               string              `json:"name"`
	ProfileAlias       *string             `json:"profileAlias"`
	UtilizationPercent uint32              `json:"utilizationPercent"`
	MemoryUsage        InstanceMemoryUsage `json:"memoryUsage"`
	Links              InstanceLinks       `json:"links"`
}

type InstanceMemoryUsage struct {
	AllocatedMiB int `json:"allocatedMiB"`
	TotalMiB     int `json:"totalMiB"`
}

type InstanceLinks struct {
	Grafana string `json:"grafana"`
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
