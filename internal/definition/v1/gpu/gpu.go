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
	Id                string                `json:"id"`
	Name              string                `json:"name"`
	Type              ResourceType          `json:"type"`
	SupportTypes      []SupportResourceType `json:"supportTypes"`
	PciAddress        string                `json:"pciAddress"`
	ProfileCountLimit *int                  `json:"profileCountLimit"`
	Status            GpuStatus             `json:"status"`
	Allocation        *AllocationSummary    `json:"allocation"`
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
	Id                   string                `json:"id"`
	Name                 string                `json:"name"`
	PciAddress           string                `json:"pciAddress"`
	ResourceType         ResourceType          `json:"resourceType"`
	SupportResourceTypes []SupportResourceType `json:"supportResourceTypes"`
	Vram                 VramInfo              `json:"vram"`
	Gpu                  GpuInfo               `json:"gpu"`
	AllocationSummary    *AllocationSummary    `json:"allocationSummary"`
	ProfileCountLimit    *int                  `json:"profileCountLimit"`
	Profiles             GpuProfileCollection  `json:"profiles"`
	AttachedInstances    *[]AttachedInstance   `json:"attachedInstances"`
	Status               GpuStatusInfo         `json:"status"`
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
	Console string `json:"console"`
}

type GpuStatusInfo struct {
	Current      GpuStatus `json:"current"`
	IsProcessing bool      `json:"isProcessing"`
}
