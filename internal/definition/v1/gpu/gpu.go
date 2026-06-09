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
	Id 								string 									`json:"id"`
	Name 							string 									`json:"name"`
	Type  						ResourceType 						`json:"type"`
	SupportTypes 			[]SupportResourceType 	`json:"supportTypes"`
	PciAddress 				string 									`json:"pciAddress"`
	ProfileCountLimit *int 										`json:"profileCountLimit"`
	Status 						GpuStatus 							`json:"status"`
	Allocation 				*AllocationSummary 			`json:"allocation"`
}

type VgpuProfileFromHex struct {
	Id 				uint32 	`json:"id"`
	Name      string  `json:"name"`
	Count     int     `json:"count"`
	Alias 		string 	`json:"alias"`
}

type PgpuAttachedInstanceFromHex struct {
	Id 		string `json:"id"`
	Name 	string `json:"name"`
}

type GpuCard struct {
	Id                   string                	`json:"id"`
	Name                 string                	`json:"name"`
	PciAddress           string                	`json:"pciAddress"`
	ResourceType         ResourceType          	`json:"resourceType"`
	SupportResourceTypes []SupportResourceType 	`json:"supportResourceTypes"`
	Vram                 *VramInfo             	`json:"vram"`
	Gpu                  *GpuInfo 							`json:"gpu"`
	AllocationSummary    *AllocationSummary    	`json:"allocationSummary"`
	VramLimitMiB 				 int 										`json:"vramLimitMiB"`
	ProfileCountLimit 	 *int 									`json:"profileCountLimit"`
	Profiles             *[]VgpuProfile         `json:"profiles"`
	AttachedInstances    *[]AttachedInstance    `json:"attachedInstances"`
	Status               GpuStatusInfo        	`json:"status"`
}

type VramInfo struct {
	AllocatedMiB       int     `json:"allocatedMiB"`
	TotalMiB           int     `json:"totalMiB"`
	UtilizationPercent float64 `json:"utilizationPercent"`
}

type GpuInfo struct {
	UtilizationPercent float64 `json:"utilizationPercent"`
}

type AllocationSummary struct {
	Current int `json:"current"`
	Total   int `json:"total"`
}

type VgpuProfile struct {
	Id        string  `json:"id"`
	Name      string  `json:"name"`
	VramMiB   uint64  `json:"vramMiB"`
	AliasName string 	`json:"aliasName"`
	Count     int     `json:"count"`
	Remaining int     `json:"remaining"`
}

type AttachedInstance struct {
	Id                 string        					`json:"id"`
	Name               string        					`json:"name"`
	ProfileAlias       *string       					`json:"profileAlias"`
	UtilizationPercent uint32       					`json:"utilizationPercent"`
	MemoryUsage        InstanceMemoryUsage   	`json:"memoryUsage"`
	Links              InstanceLinks 					`json:"links"`
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
	Current      GpuStatus 	`json:"current"`
	IsProcessing bool       `json:"isProcessing"`
}
