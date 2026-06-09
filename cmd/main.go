package main

import (
	"fmt"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
	_ "github.com/bigstack-oss/cube-cos-api/api"
	"github.com/bigstack-oss/cube-cos-api/internal/config"
	"github.com/bigstack-oss/cube-cos-api/internal/runtime"
	svc "github.com/bigstack-oss/cube-cos-api/internal/service"
	log "go-micro.dev/v5/logger"
)

func nvmlTest() {
	// 1. 初始化 NVML (這是最重要的起手式)
	ret := nvml.Init()
	if ret != nvml.SUCCESS {
		log.Fatalf("無法初始化 NVML: %v", nvml.ErrorString(ret))
	}
	// 記得在程式結束時關閉它
	defer nvml.Shutdown()

	// 2. 取得目前系統上的 GPU 數量
	count, ret := nvml.DeviceGetCount()
	if ret != nvml.SUCCESS {
		log.Fatalf("無法取得 GPU 數量: %v", nvml.ErrorString(ret))
	}
	fmt.Printf("🎉 成功抓到！這台機器上有 %d 張 GPU\n", count)

	// 3. 抓取第 0 張卡的名稱
	device, ret := nvml.DeviceGetHandleByIndex(0)
	if ret != nvml.SUCCESS {
		log.Fatalf("無法取得 GPU Handle: %v", nvml.ErrorString(ret))
	}

	name, _ := device.GetName()
	pci, _ := device.GetPciInfo()

	pciAddress := fmt.Sprintf("%04x:%02x:%02x.0", pci.Domain, pci.Bus, pci.Device)

	var b []byte
	for _, v := range pci.BusId {
		if v == 0 {
			break
		}
		b = append(b, byte(v))
	}
	pciAddress2 := string(b)

	// 抓取 VRAM 記憶體資訊 (對應 UI 的 VRAM Allocation)
	memory, _ := device.GetMemoryInfo()

	// 抓取使用率 (對應 UI 的 Utilization)
	utilization, _ := device.GetUtilizationRates()

	log.Errorf("👉 第 0 張卡的型號是: %s\n", name)
	log.Errorf("👉 第 0 張卡的 PCI 是: %s\n", pciAddress)
	log.Errorf("👉 第 0 張卡的 PCI 是: %s\n", pciAddress2)
	log.Errorf("👉 第 0 張卡的 VRAM 總量: %d MB\n", memory.Total/1024/1024)
	log.Errorf("👉 第 0 張卡的 VRAM 使用量: %d MB\n", memory.Used/1024/1024)
	log.Errorf("👉 第 0 張卡的 VRAM 空閒量: %d MB\n", memory.Free/1024/1024)
	log.Errorf("👉 第 0 張卡的 GPU 使用率: %d%%\n", utilization.Gpu)
	log.Errorf("👉 第 0 張卡的 Memory 使用率: %d%%\n", utilization.Memory)
}

func main() {
	err := config.SyncOptions()

	if err != nil {
		log.Errorf("failed to load config(%v)", err)
		return
	}

	srv, err := runtime.NewHttpServer()
	if err != nil {
		log.Errorf("failed to init runtime(%v)", err)
		return
	}

	err = svc.Micro(srv).Run()
	if err != nil {
		log.Errorf("failed to run service(%v)", err)
	}
}