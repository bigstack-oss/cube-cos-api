package cubecos

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// nvidiaSmiQFixture is real `nvidia-smi -q` output captured on cn13 (RTX PRO
// 6000 Blackwell, GPU slot 3) during the nvidia-smi feasibility research,
// trimmed to the single-device case this parser needs to handle. Kept
// verbatim (including the sections this package does not parse) so the
// section-scoping logic is exercised against the real field-name collisions
// (Total/Used/Reserved repeated across FB/BAR1/Conf Compute sections).
const nvidiaSmiQFixture = `
==============NVSMI LOG==============

Timestamp                                 : Wed Jul 22 15:37:16 2026
Driver Version                            : 580.105.06
CUDA Version                              : Not Found
vGPU Driver Capability
        Heterogenous Multi-vGPU           : Supported

Attached GPUs                             : 4
GPU 00000001:C8:00.0
    Product Name                          : NVIDIA RTX PRO 6000 Blackwell Server Edition
    Product Brand                         : NVIDIA
    Display Mode                          : Requested functionality has been deprecated
    GPU UUID                              : GPU-0a5ba9ad-a575-72c7-d51d-3e726232cfd1
    Minor Number                          : 3
    PCI
        Bus                               : 0xC8
        Bus Id                            : 00000001:C8:00.0
    FB Memory Usage
        Total                             : 97887 MiB
        Reserved                          : 2288 MiB
        Used                              : 5952 MiB
        Free                              : 89648 MiB
    BAR1 Memory Usage
        Total                             : 131072 MiB
        Used                              : 9 MiB
        Free                              : 131063 MiB
    Conf Compute Protected Memory Usage
        Total                             : 0 MiB
        Used                              : 0 MiB
        Free                              : 0 MiB
    Utilization
        GPU                               : 0 %
        Memory                            : 0 %
        Encoder                           : 0 %
        Decoder                           : 0 %
        JPEG                              : 0 %
        OFA                               : 0 %
    Encoder Stats
        Active Sessions                   : 0
        Average FPS                       : 0
        Average Latency                   : 0
`

// nvidiaSmiVgpuQFixture is real `nvidia-smi vgpu -q` output captured on cn13
// (same device as above) with two active vGPU instances attached to two
// running VMs, taken from the nvidia-smi feasibility research's raw records.
const nvidiaSmiVgpuQFixture = `GPU 00000001:C8:00.0
    Active vGPUs                          : 2
    vGPU ID                               : 3251637554
        VM UUID                           : 083bc63f-0a3c-4647-9bb4-27f905b7d7bb
        VM Name                           : instance-00000004
        vGPU Name                         : NVIDIA RTX Pro 6000 Blackwell DC-3Q
        vGPU Type                         : 1519
        vGPU UUID                         : c47a3625-859f-11f1-84fe-7572c7d51d3e
        Guest Driver Version              : N/A
        License Status                    : N/A (Expiry: N/A)
        GPU Instance ID                   : N/A
        Placement ID                      : 0
        PCI
            Bus Id                        : 00000000:00:00.0
        FB Memory Usage
            Total                         : 3072 MiB
            Used                          : 128 MiB
            Free                          : 2944 MiB
        Utilization
            GPU                           : 0 %
            Memory                        : 0 %
            Encoder                       : 0 %
            Decoder                       : 0 %
    vGPU ID                               : 3251637566
        VM UUID                           : f0adb76f-7612-4c68-84b4-375a4e6c5476
        VM Name                           : instance-00000005
        vGPU Name                         : NVIDIA RTX Pro 6000 Blackwell DC-3Q
        vGPU Type                         : 1519
        vGPU UUID                         : 094593d6-85a0-11f1-802e-72c7d51d3e72
        Guest Driver Version              : 580.105.08
        License Status                    : Unlicensed (Unrestricted)
        GPU Instance ID                   : N/A
        Placement ID                      : 3
        PCI
            Bus Id                        : 00000000:00:05.0
        FB Memory Usage
            Total                         : 3072 MiB
            Used                          : 128 MiB
            Free                          : 2944 MiB
        Utilization
            GPU                           : 12 %
            Memory                        : 7 %
            Encoder                       : 0 %
            Decoder                       : 0 %
`

func TestParseNvidiaSmiDevices(t *testing.T) {
	devices := parseNvidiaSmiDevices(nvidiaSmiQFixture)

	require.Len(t, devices, 1)

	device, ok := devices["GPU-0a5ba9ad-a575-72c7-d51d-3e726232cfd1"]
	require.True(t, ok)
	require.Equal(t, 97887, device.MemoryTotalMiB)
	// Reserved(2288) + Used(5952), matching the legacy NVML v1 GetMemoryInfo
	// "Total - Free" semantics the existing API contract keeps.
	require.Equal(t, 8240, device.MemoryUsedMiB)
	require.Equal(t, uint32(0), device.GpuUtilizationPercent)
	require.Equal(t, uint32(0), device.MemoryUtilizationPercent)
}

// A device block missing a required field (no GPU UUID line, e.g. a stripped
// or truncated block) must not be reported at all -- an absent device is the
// same "not visible" signal a vfio-passthrough pgpu produces, and a caller
// must not mistake a zero-valued partial parse for a real, healthy card.
func TestParseNvidiaSmiDevicesSkipsIncompleteBlock(t *testing.T) {
	incomplete := `GPU 00000001:C9:00.0
    Product Name                          : NVIDIA RTX PRO 6000 Blackwell Server Edition
`
	devices := parseNvidiaSmiDevices(incomplete)

	require.Empty(t, devices)
}

func TestParseNvidiaSmiDevicesNoDevices(t *testing.T) {
	require.Empty(t, parseNvidiaSmiDevices("no devices were found"))
}

func TestParseNvidiaSmiVgpuInstances(t *testing.T) {
	instancesByPci := parseNvidiaSmiVgpuInstances(nvidiaSmiVgpuQFixture)

	require.Len(t, instancesByPci, 1)
	instances := instancesByPci["00000001:C8:00.0"]
	require.Len(t, instances, 2)

	require.Equal(t, NvidiaSmiVgpuInstance{
		VmUUID:                "083bc63f-0a3c-4647-9bb4-27f905b7d7bb",
		VgpuTypeId:            1519,
		MemoryUsedMiB:         128,
		MemoryTotalMiB:        3072,
		GpuUtilizationPercent: 0,
	}, instances[0])

	require.Equal(t, NvidiaSmiVgpuInstance{
		VmUUID:                "f0adb76f-7612-4c68-84b4-375a4e6c5476",
		VgpuTypeId:            1519,
		MemoryUsedMiB:         128,
		MemoryTotalMiB:        3072,
		GpuUtilizationPercent: 12,
	}, instances[1])
}

func TestParseNvidiaSmiVgpuInstancesNoActiveVgpus(t *testing.T) {
	fixture := `GPU 00000001:C8:00.0
    Active vGPUs                          : 0
`
	instancesByPci := parseNvidiaSmiVgpuInstances(fixture)

	// A device with zero active instances is absent from the map entirely,
	// same as a device nvidia-smi never mentions.
	require.Empty(t, instancesByPci)
}

func TestParseLeadingIntField(t *testing.T) {
	n, err := parseLeadingIntField("97887 MiB")
	require.NoError(t, err)
	require.Equal(t, 97887, n)

	n, err = parseLeadingIntField("0 %")
	require.NoError(t, err)
	require.Equal(t, 0, n)

	_, err = parseLeadingIntField("N/A")
	require.Error(t, err)

	_, err = parseLeadingIntField("")
	require.Error(t, err)
}

func TestSplitNvidiaSmiKeyValue(t *testing.T) {
	key, value, ok := splitNvidiaSmiKeyValue("    Bus Id                            : 00000001:C8:00.0")
	require.True(t, ok)
	require.Equal(t, "Bus Id", key)
	// The value itself contains colons; only the " : " separator splits.
	require.Equal(t, "00000001:C8:00.0", value)

	_, _, ok = splitNvidiaSmiKeyValue("    FB Memory Usage")
	require.False(t, ok)
}
