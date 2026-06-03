package cubecos

import (
	"fmt"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/base"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/gpu"
)

func GetMockGpus() []gpu.GpuCard {
	h100_2_20C := "H100-2-20C"
	h100_3_40C := "H100-3-40C"
	
	return []gpu.GpuCard{
		{
			Id: "GPU-28eca022-0c66-97c1-a527-17e900626924",
			Name: "[Mock] NVIDIA RTX A2000",
			PciAddress: "00000000:86:00.0",
			ResourceType: gpu.ResourceTypeUnset,
			SupportResourceTypes: []gpu.SupportResourceType{},
			Vram: nil,
			Gpu: nil,
			AllocationSummary: nil,
			Profiles: nil,
			AttachedInstances: nil,
			Status: gpu.GpuStatusInfo{
				Current: gpu.GpuStatusUnassigned,
				IsProcessing: false,
			},
		},
		{
			Id: "GPU-28eca022-0c66-97c1-a527-17e900626924",
			Name: "[Mock] NVIDIA RTX A2000",
			PciAddress: "00000000:87:00.0",
			ResourceType: gpu.ResourceTypePgpu,
			SupportResourceTypes: []gpu.SupportResourceType{
				gpu.SupportResourceTypePgpu,
			},
			Vram: &gpu.VramInfo{
				AllocatedMiB: 0,
				TotalMiB: 6138,
				UtilizationPercent: 0,
			},
			Gpu: &gpu.GpuInfo{
				UtilizationPercent: 0,
			},
			AllocationSummary: &gpu.AllocationSummary{
				Current: 0,
				Total: 1,
			},
			Profiles: nil,
			AttachedInstances: &[]gpu.AttachedInstance{},
			Status: gpu.GpuStatusInfo{
				Current: gpu.GpuStatusIdle,
				IsProcessing: false,
			},
		},
		{
			Id: "GPU-97205c71-e8c3-4156-ab31-eb410b0b3c01",
			Name: "[Mock] NVIDIA RTX A2000",
			PciAddress: "00000000:88:00.0",
			ResourceType: gpu.ResourceTypePgpu,
			SupportResourceTypes: []gpu.SupportResourceType{
				gpu.SupportResourceTypePgpu,
			},
			Vram: &gpu.VramInfo{
				AllocatedMiB: 1473,
				TotalMiB: 6138,
				UtilizationPercent: 24,
			},
			Gpu: &gpu.GpuInfo{
				UtilizationPercent: 12,
			},
			AllocationSummary: &gpu.AllocationSummary{
				Current: 1,
				Total: 1,
			},
			Profiles: nil,
			AttachedInstances: &[]gpu.AttachedInstance{
				{
					Id: "2dd6020e-bff8-4f04-8d41-5ee3a5ae3362",
					Name: "VM 01",
					ProfileAlias: nil,
					UtilizationPercent: 12,
					MemoryUsage: gpu.InstanceMemoryUsage{
						AllocatedMiB: 1473,
						TotalMiB: 6138,
					},
					Links: buildInstanceLinks("2dd6020e-bff8-4f04-8d41-5ee3a5ae3362"),
				},
			},
			Status: gpu.GpuStatusInfo{
				Current: gpu.GpuStatusInUse,
				IsProcessing: false,
			},
		},
		{
			Id: "GPU-0a4167a2-f1a8-4759-a790-70d70ca217b4",
			Name: "[Mock] NVIDIA H100",
			PciAddress: "00000000:89:00.0",
			ResourceType: gpu.ResourceTypeSriovVgpu,
			SupportResourceTypes: []gpu.SupportResourceType{
				gpu.SupportResourceTypePgpu,
				gpu.SupportResourceTypeSriovVgpu,
			},
			Vram: &gpu.VramInfo{
				AllocatedMiB: 6144,
				TotalMiB: 81920,
				UtilizationPercent: 7.5,
			},
			Gpu: &gpu.GpuInfo{
				UtilizationPercent: 15,
			},
			AllocationSummary: &gpu.AllocationSummary{
				Current: 5,
				Total: 7,
			},
			Profiles: &[]gpu.VgpuProfile{
				{
					Id: "517",
					Name: "H100-1-10C",
					VramMiB: 10240,
					AliasName: "H100-1-10C",
					Count: 0,
					Remaining: 0,
				},
				{
					Id: "518",
					Name: "H100-1-10CME",
					VramMiB: 10240,
					AliasName: "H100-1-10CME",
					Count: 0,
					Remaining: 0,
				},
				{
					Id: "519",
					Name: h100_2_20C,
					VramMiB: 20480,
					AliasName: h100_2_20C,
					Count: 1,
					Remaining: 0,
				},
				{
					Id: "520",
					Name: h100_3_40C,
					VramMiB: 40960,
					AliasName: h100_3_40C,
					Count: 1,
					Remaining: 0,
				},
			},
			AttachedInstances: &[]gpu.AttachedInstance{
				{
					Id: "23727ab3-2f33-4ab6-9a23-b27e958ba325",
					Name: "VM 01",
					ProfileAlias: &h100_2_20C,
					UtilizationPercent: 6,
					MemoryUsage: gpu.InstanceMemoryUsage{
						AllocatedMiB: 2048,
						TotalMiB: 20480,
					},
					Links: buildInstanceLinks("23727ab3-2f33-4ab6-9a23-b27e958ba325"),
				},
				{
					Id: "c9d4a964-3d93-43ed-968e-cd2c4dd97443",
					Name: "VM 02",
					ProfileAlias: &h100_3_40C,
					UtilizationPercent: 9,
					MemoryUsage: gpu.InstanceMemoryUsage{
						AllocatedMiB: 4096,
						TotalMiB: 40960,
					},
					Links: buildInstanceLinks("c9d4a964-3d93-43ed-968e-cd2c4dd97443"),
				},
			},
			Status: gpu.GpuStatusInfo{
				Current: gpu.GpuStatusInUse,
				IsProcessing: false,
			},
		},
		{
			Id: "GPU-eb12f26f-5af0-41e9-8bca-42ced02bdff4",
			Name: "[Mock] NVIDIA A100",
			PciAddress: "00000000:90:00.0",
			ResourceType: gpu.ResourceTypeMigBackedVgpu,
			SupportResourceTypes: []gpu.SupportResourceType{
				gpu.SupportResourceTypePgpu,
				gpu.SupportResourceTypeMigBackedVgpu,
			},
			Vram: &gpu.VramInfo{
				AllocatedMiB: 1234,
				TotalMiB: 81920,
				UtilizationPercent: 1.506347,
			},
			Gpu: &gpu.GpuInfo{
				UtilizationPercent: 9,
			},
			AllocationSummary: &gpu.AllocationSummary{
				Current: 0,
				Total: 7,
			},
			Profiles: &[]gpu.VgpuProfile{
				{
					Id: "475",
					Name: "A100-1-10C",
					VramMiB: 10240,
					AliasName: "A100-1-10C",
					Count: 2,
					Remaining: 2,
				},
				{
					Id: "476",
					Name: "A100-1-10CME",
					VramMiB: 10240,
					AliasName: "A100-1-10CME",
					Count: 0,
					Remaining: 0,
				},
				{
					Id: "477",
					Name: "A100-2-20C",
					VramMiB: 20480,
					AliasName: "A100-2-20C",
					Count: 1,
					Remaining: 1,
				},
				{
					Id: "478",
					Name: "A100-3-40C",
					VramMiB: 40960,
					AliasName: "A100-3-40C",
					Count: 1,
					Remaining: 1,
				},
			},
			AttachedInstances: &[]gpu.AttachedInstance{},
			Status: gpu.GpuStatusInfo{
				Current: gpu.GpuStatusIdle,
				IsProcessing: false,
			},
		},
	}
}

func buildInstanceLinks(vmId string) gpu.InstanceLinks {
	return gpu.InstanceLinks{
		Grafana: fmt.Sprintf(
			"https://%s/grafana/d/PVW6vU7Wz/instance?refresh=5m&kiosk=tv&orgId=1&var-%s",
			base.DataCenterVip,
			vmId,
		),
		Console: fmt.Sprintf(
			"https://%s:6080/console/mock/%s",
			base.DataCenterVip,
			vmId,
		),
	}
}