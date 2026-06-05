# `listLocalGpuCards` Function Explanation

## Overview

`listLocalGpuCards` is a method on the `helper` struct that queries the **local** node's GPU hardware using NVIDIA's NVML library and returns a detailed inventory of all GPU cards, including VRAM usage, GPU utilization, vGPU profiles, and attached VM instances.

It is the local-node path of `listNodeGpuCards()`, which decides between local and remote execution based on whether the target node is the current machine.

## Architecture Context

```mermaid
flowchart TD
    A[HTTP GET /v1/nodes/:nodeName/gpuCards] --> B[listNodeGpuCards handler]
    B --> C{Is node local?}
    C -->|Yes| D["listLocalGpuCards()"]
    C -->|No| E["listRemoteGpuCards()"]
    D --> F[hex_sdk: list_gpus]
    D --> G[NVML: Device enumeration]
    D --> H[listVgpuProfiles]
    D --> I[listAttachedInstances]
    I --> J{GPU resource type?}
    J -->|pgpu| K[listPgpuAttachedInstances]
    J -->|sriovVgpu / migBackedVgpu| L[listVgpuAttachedInstances]
    K --> M[OpenStack: list servers]
    L --> N[NVML: active vGPUs]
    L --> O[OpenStack: get server]
    E --> P[HTTP call to remote node]
```

## Data Flow

```mermaid
sequenceDiagram
    participant Client
    participant Handler as Gin Handler
    participant Helper as helper struct
    participant HexSDK as hex_sdk CLI
    participant NVML as NVIDIA NVML
    participant OpenStack

    Client->>Handler: GET /v1/nodes/:nodeName/gpuCards
    Handler->>Helper: listNodeGpuCards()
    Helper->>Helper: IsLocal(node)?
    Helper->>HexSDK: GetNodeGpusMap(node)
    HexSDK-->>Helper: map[pciAddress] → GpuFromHex

    Helper->>NVML: DeviceGetCount()
    loop For each GPU device (0..count-1)
        Helper->>NVML: DeviceGetHandleByIndex(i)
        Helper->>NVML: GetUUID()
        Helper->>NVML: GetPciInfo()
        Helper->>NVML: GetMemoryInfo()
        Helper->>NVML: GetUtilizationRates()
        Helper->>Helper: extractPciAddress(pciInfo)
        Helper->>Helper: listVgpuProfiles(device, hexGpu)
        Helper->>Helper: listAttachedInstances(opts)
        alt pGPU
            Helper->>OpenStack: ListServers(host=node)
        else vGPU (SR-IOV / MIG)
            Helper->>NVML: GetActiveVgpus()
            Helper->>OpenStack: GetServer(vmId)
        end
        Helper->>Helper: updateVgpuProfilesRemaining()
    end
    Helper-->>Handler: []gpu.GpuCard
    Handler-->>Client: 200 OK + JSON payload
```

## Step-by-Step Walkthrough

| Step | Action                                                          | Source                        |
| ---- | --------------------------------------------------------------- | ----------------------------- |
| 1    | Fetch GPU metadata from `hex_sdk list_gpus` (JSON)              | `cubecos.GetNodeGpusMap`      |
| 2    | Get total NVIDIA device count via NVML                          | `nvml.DeviceGetCount()`       |
| 3    | **Per device:** get handle, UUID, PCI info, memory, utilization | NVML calls                    |
| 4    | Normalize PCI address (strip 8-char domain → 4-char)            | `extractPciAddress`           |
| 5    | Look up hex metadata by PCI address                             | `hexGpusMap[pciAddress]`      |
| 6    | List vGPU profiles (MIG / SR-IOV only)                          | `listVgpuProfiles`            |
| 7    | List attached VM instances                                      | `listAttachedInstances`       |
| 8    | Calculate remaining profile slots                               | `updateVgpuProfilesRemaining` |
| 9    | Assemble `gpu.GpuCard` struct and append                        | —                             |

## Key Data Structures

```mermaid
classDiagram
    class GpuCard {
        +string Id
        +string Name
        +string PciAddress
        +ResourceType ResourceType
        +VramInfo Vram
        +GpuInfo Gpu
        +AllocationSummary AllocationSummary
        +[]VgpuProfile Profiles
        +[]AttachedInstance AttachedInstances
        +GpuStatusInfo Status
    }
    class VramInfo {
        +int AllocatedMiB
        +int TotalMiB
        +float64 UtilizationPercent
    }
    class GpuInfo {
        +float64 UtilizationPercent
    }
    class AllocationSummary {
        +int Current
        +int Total
    }
    class VgpuProfile {
        +string Id
        +string Name
        +uint64 VramMiB
        +string AliasName
        +int Count
        +int Remaining
    }
    class AttachedInstance {
        +string Id
        +string Name
        +string ProfileAlias
        +uint32 UtilizationPercent
        +InstanceMemoryUsage MemoryUsage
        +InstanceLinks Links
    }
    GpuCard --> VramInfo
    GpuCard --> GpuInfo
    GpuCard --> AllocationSummary
    GpuCard --> VgpuProfile : 0..*
    GpuCard --> AttachedInstance : 0..*
```

## External Dependencies

| Dependency                  | Purpose                                                                                        |
| --------------------------- | ---------------------------------------------------------------------------------------------- |
| `github.com/NVIDIA/go-nvml` | Direct access to NVIDIA GPU hardware (device enumeration, memory, utilization, vGPU instances) |
| `hex_sdk` CLI               | Provides platform-level GPU metadata (ID, name, type, status, allocation, vGPU profile counts) |
| OpenStack Compute API       | Resolves VM instance names and creates VNC console links                                       |
| Grafana                     | Generates per-instance monitoring dashboard URLs                                               |

## Error Handling Strategy

- **Fatal errors** (stop processing): failure to get device count, failure to call `hex_sdk`.
- **Non-fatal errors** (skip device, continue loop): failure to get device handle, UUID, or PCI info for a single device.
- **Non-fatal warnings** (use zero values): failure to get memory info or utilization rates — the card is still reported with zeroed metrics.

## GPU Resource Type Branching

```mermaid
flowchart LR
    A[ResourceType] -->|unset| B[No attached instances]
    A -->|pgpu| C[Single active server on node]
    A -->|sriovVgpu| D[Multiple vGPU instances via NVML]
    A -->|migBackedVgpu| D
```

- **pGPU (passthrough):** The entire physical GPU is passed through to one VM. The function finds the active OpenStack server on that node.
- **SR-IOV vGPU / MIG-backed vGPU:** The GPU is partitioned. NVML's `GetActiveVgpus()` returns individual vGPU instances, each mapped to a VM via its VM ID.

## File Location

- **Source:** `internal/apis/v1/handlers/nodes/gpu.go`
- **Handler registration:** `internal/apis/v1/handlers/nodes/handlers.go` (route: `GET /v1/nodes/:nodeName/gpuCards`)
- **Type definitions:** `internal/definition/v1/gpu/gpu.go`
- **hex_sdk integration:** `internal/cubecos/nodes.go`
