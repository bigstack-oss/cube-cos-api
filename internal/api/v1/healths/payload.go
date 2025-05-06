package healths

import (
	"github.com/bigstack-oss/cube-cos-api/internal/api"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	log "go-micro.dev/v5/logger"
)

func (h *helper) genServiceHealthHistory() []cubecos.HealthStatus {
	return cubecos.GetServiceHealthHistory(h.serviceType, h.past)
}

func (h *helper) genModuleHealthHistory() cubecos.HealthStatus {
	service := cubecos.ModuleToService[h.moduleType]
	history, err := cubecos.GetModuleHealthHistory(h.moduleType, h.past, v1.AscSort, false)
	if err != nil {
		log.Errorf("healths(%s): %v", api.GetReqId(h.c), err)
	}

	return cubecos.HealthStatus{
		Category:     cubecos.ServiceToCategory[service],
		Name:         service,
		Module:       h.moduleType,
		IsRepairable: cubecos.IsRepairableModule(h.moduleType),
		History:      history,
	}
}

func genCheckRepairReq() *cubecos.Health {
	h := &cubecos.Health{}
	h.Overall = &cubecos.Overall{}
	h.Overall.Status.SetDesiredToCheckingAndRepairing()
	return h
}

func genForceRepairReq(module v1.Module) *cubecos.Health {
	h := &cubecos.Health{}
	h.Overall = &cubecos.Overall{}
	h.Overall.Status.SetDesiredToRepairing()
	svc := cubecos.ModuleToService[module.Name]
	h.Services = []v1.Service{
		{
			Name:     svc,
			Category: cubecos.ServiceToCategory[svc],
			Modules:  []v1.Module{module},
		},
	}

	return h
}
