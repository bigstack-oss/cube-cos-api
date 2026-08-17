package cubecos

import (
	"context"
	"os/exec"
	"strconv"
	"strings"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	log "go-micro.dev/v5/logger"
)

// NvidiaSmiDevice is one GPU's device-level stats as reported by `nvidia-smi
// -q`. MemoryUsedMiB is Reserved+Used: newer drivers split a fixed ECC/driver
// reservation out of Used that the previous NVML v1 GetMemoryInfo() call
// folded into a single "Total - Free" number, so callers that want the old
// allocatedMiB semantics need both fields summed back together.
// MemoryTotalMiB and MemoryUsedMiB are plain values because a device record is
// only returned when both were parsed (see parseNvidiaSmiDeviceBlock). The
// utilization fields are pointers because nvidia-smi reports them as `N/A`
// whenever the card cannot produce them at all -- most notably once MIG is
// enabled, where NVIDIA provides no device-level utilization -- and a nil there
// must stay distinguishable from a real 0%.
type NvidiaSmiDevice struct {
	UUID                     string
	MemoryTotalMiB           int
	MemoryUsedMiB            int
	GpuUtilizationPercent    *uint32
	MemoryUtilizationPercent *uint32
}

// NvidiaSmiVgpuInstance is one active vGPU instance as reported by `nvidia-smi
// vgpu -q`. VgpuTypeId is the vGPU Type ID, which is what hex reports as the
// profile id for both vGPU flavours, so it maps straight onto a hex profile
// with no further lookup.
//
// Unlike a device record, an instance is reported as soon as it has a VM UUID,
// so none of its stats are guaranteed present: a MIG-backed vGPU reports its
// framebuffer but `N/A` for every utilization, and a nil must not be reported
// as 0.
type NvidiaSmiVgpuInstance struct {
	VmUUID                string
	VgpuTypeId            uint32
	MemoryUsedMiB         *int
	MemoryTotalMiB        *int
	GpuUtilizationPercent *uint32
}

// GetNvidiaSmiDevices runs `nvidia-smi -q` once and returns every GPU
// currently visible to nvidia-smi, keyed by UUID. A GPU bound to vfio-pci for
// passthrough is invisible to nvidia-smi and simply absent from the map --
// hex's own gpu_device_list relies on the same behavior (see sdk_gpu.sh's
// "GPUs already bound to vfio-pci... no longer enumerable by nvidia-smi"), so
// that is not treated as an error here either. The returned error is non-nil
// only when nvidia-smi itself could not be run at all (driver not loaded,
// binary missing).
func GetNvidiaSmiDevices() (map[string]NvidiaSmiDevice, error) {
	output, err := runNvidiaSmi("-q")
	if err != nil {
		return nil, err
	}

	return parseNvidiaSmiDevices(output), nil
}

// GetNvidiaSmiVgpuInstances runs `nvidia-smi vgpu -q` once and returns every
// active vGPU instance grouped by the PCI bus id of the physical GPU hosting
// it. hex's PciAddress is itself sourced from nvidia-smi's own pci.bus_id
// (see sdk_gpu.sh's gpu_device_list), so the two use the same format and can
// be matched directly. A GPU with no active instances is simply absent from
// the map.
func GetNvidiaSmiVgpuInstances() (map[string][]NvidiaSmiVgpuInstance, error) {
	output, err := runNvidiaSmi("vgpu", "-q")
	if err != nil {
		return nil, err
	}

	return parseNvidiaSmiVgpuInstances(output), nil
}

func runNvidiaSmi(args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(30))
	defer cancel()

	out, err := exec.CommandContext(ctx, "nvidia-smi", args...).Output()
	if err != nil {
		log.Errorf("nvidiasmi: failed to run nvidia-smi %v: %v", args, err)
		return "", err
	}

	return string(out), nil
}

// parseNvidiaSmiDevices splits `nvidia-smi -q` output into per-device blocks
// (each headed by a "GPU <bus id>" line at zero indent) and parses each one.
func parseNvidiaSmiDevices(output string) map[string]NvidiaSmiDevice {
	devices := map[string]NvidiaSmiDevice{}

	for _, block := range splitTopLevelBlocks(output, "GPU ") {
		device, ok := parseNvidiaSmiDeviceBlock(block[1:])
		if !ok {
			continue
		}
		devices[device.UUID] = device
	}

	return devices
}

// parseNvidiaSmiDeviceBlock parses one device's fields out of the lines under
// its "GPU <bus id>" header. Real nvidia-smi -q output reuses field names
// ("Total", "Used") across multiple sibling sections (FB Memory Usage, BAR1
// Memory Usage, Conf Compute Protected Memory Usage), so values are only
// captured while inside the specific section they belong to -- tracked via
// the nearest preceding 4-indent header line -- to avoid one section's value
// silently overwriting another's.
func parseNvidiaSmiDeviceBlock(lines []string) (NvidiaSmiDevice, bool) {
	device := NvidiaSmiDevice{}
	section := ""
	haveTotal, haveUsed, haveReserved := false, false, false
	reservedMiB := 0

	for _, line := range lines {
		indent := leadingSpaceCount(line)
		key, value, isKV := splitNvidiaSmiKeyValue(line)

		if indent == 4 {
			if !isKV {
				section = strings.TrimSpace(line)
				continue
			}
			section = ""
			if key == "GPU UUID" {
				device.UUID = value
			}
			continue
		}

		if indent != 8 || !isKV {
			continue
		}

		switch {
		case section == "FB Memory Usage" && key == "Total":
			if n, err := parseLeadingIntField(value); err == nil {
				device.MemoryTotalMiB = n
				haveTotal = true
			}
		case section == "FB Memory Usage" && key == "Reserved":
			if n, err := parseLeadingIntField(value); err == nil {
				reservedMiB = n
				haveReserved = true
			}
		case section == "FB Memory Usage" && key == "Used":
			if n, err := parseLeadingIntField(value); err == nil {
				device.MemoryUsedMiB = n
				haveUsed = true
			}
		case section == "Utilization" && key == "GPU":
			if n, err := parseLeadingUint32Field(value); err == nil {
				device.GpuUtilizationPercent = &n
			}
		case section == "Utilization" && key == "Memory":
			if n, err := parseLeadingUint32Field(value); err == nil {
				device.MemoryUtilizationPercent = &n
			}
		}
	}

	if device.UUID == "" || !haveTotal || !haveUsed {
		return NvidiaSmiDevice{}, false
	}

	if haveReserved {
		device.MemoryUsedMiB += reservedMiB
	}

	return device, true
}

// parseNvidiaSmiVgpuInstances splits `nvidia-smi vgpu -q` output into
// per-device blocks (same "GPU <bus id>" header convention as `-q`) and
// parses the active vGPU instances nested under each one.
func parseNvidiaSmiVgpuInstances(output string) map[string][]NvidiaSmiVgpuInstance {
	instances := map[string][]NvidiaSmiVgpuInstance{}

	for _, block := range splitTopLevelBlocks(output, "GPU ") {
		pciAddress := strings.TrimSpace(strings.TrimPrefix(block[0], "GPU"))

		deviceInstances := parseNvidiaSmiVgpuInstanceBlock(block[1:])
		if len(deviceInstances) > 0 {
			instances[pciAddress] = deviceInstances
		}
	}

	return instances
}

// parseNvidiaSmiVgpuInstanceBlock parses the vGPU instances nested under one
// device's block. Each instance starts at a 4-indent "vGPU ID" line; its
// scalar fields (VM UUID, vGPU Type) sit at 8-indent, and its FB Memory
// Usage/Utilization sub-sections nest their fields one level deeper again at
// 12-indent -- tracked the same section-scoped way as the device-level parser.
func parseNvidiaSmiVgpuInstanceBlock(lines []string) []NvidiaSmiVgpuInstance {
	var result []NvidiaSmiVgpuInstance
	var current *NvidiaSmiVgpuInstance
	section := ""

	flush := func() {
		if current != nil {
			result = append(result, *current)
		}
	}

	for _, line := range lines {
		indent := leadingSpaceCount(line)
		key, value, isKV := splitNvidiaSmiKeyValue(line)

		switch {
		case indent == 4 && isKV && key == "vGPU ID":
			flush()
			current = &NvidiaSmiVgpuInstance{}
			section = ""
		case current == nil:
			continue
		case indent == 8 && !isKV:
			section = strings.TrimSpace(line)
		case indent == 8 && isKV:
			section = ""
			switch key {
			case "VM UUID":
				current.VmUUID = value
			case "vGPU Type":
				if n, err := parseLeadingUint32Field(value); err == nil {
					current.VgpuTypeId = n
				}
			}
		case indent == 12 && isKV:
			switch {
			case section == "FB Memory Usage" && key == "Total":
				if n, err := parseLeadingIntField(value); err == nil {
					current.MemoryTotalMiB = &n
				}
			case section == "FB Memory Usage" && key == "Used":
				if n, err := parseLeadingIntField(value); err == nil {
					current.MemoryUsedMiB = &n
				}
			case section == "Utilization" && key == "GPU":
				if n, err := parseLeadingUint32Field(value); err == nil {
					current.GpuUtilizationPercent = &n
				}
			}
		}
	}
	flush()

	return result
}

// splitTopLevelBlocks splits text into blocks, each starting at a zero-indent
// line beginning with headerPrefix and running until the next such line (or
// EOF). Text before the first matching line is discarded.
func splitTopLevelBlocks(output, headerPrefix string) [][]string {
	var blocks [][]string
	var current []string

	for _, line := range strings.Split(output, "\n") {
		if leadingSpaceCount(line) == 0 && strings.HasPrefix(line, headerPrefix) {
			if current != nil {
				blocks = append(blocks, current)
			}
			current = []string{line}
			continue
		}
		if current != nil {
			current = append(current, line)
		}
	}
	if current != nil {
		blocks = append(blocks, current)
	}

	return blocks
}

func leadingSpaceCount(line string) int {
	return len(line) - len(strings.TrimLeft(line, " "))
}

// splitNvidiaSmiKeyValue splits an nvidia-smi "Key   : Value" line on the
// first " : " separator. Some values (e.g. a PCI bus id) contain colons of
// their own without surrounding spaces, so splitting on the spaced separator
// rather than the first bare ":" avoids truncating them.
func splitNvidiaSmiKeyValue(line string) (key, value string, ok bool) {
	idx := strings.Index(line, " : ")
	if idx < 0 {
		return "", "", false
	}

	return strings.TrimSpace(line[:idx]), strings.TrimSpace(line[idx+3:]), true
}

// parseLeadingIntField parses the leading integer out of values like "97887
// MiB" or "0 %".
func parseLeadingIntField(value string) (int, error) {
	fields := strings.Fields(value)
	if len(fields) == 0 {
		return 0, strconv.ErrSyntax
	}

	return strconv.Atoi(fields[0])
}

// parseLeadingUint32Field is parseLeadingIntField for the fields that land in
// a uint32 (utilization percentages, vGPU type ids, GPU instance profile ids).
// Parsing into an int and converting would wrap silently on a value past
// uint32 - the same "quietly produce a plausible wrong number" failure this
// package exists to avoid - so bound it at parse time instead.
func parseLeadingUint32Field(value string) (uint32, error) {
	fields := strings.Fields(value)
	if len(fields) == 0 {
		return 0, strconv.ErrSyntax
	}

	n, err := strconv.ParseUint(fields[0], 10, 32)
	if err != nil {
		return 0, err
	}

	return uint32(n), nil
}
