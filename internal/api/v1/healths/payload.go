package healths

import (
	"github.com/bigstack-oss/cube-cos-api/internal/api"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	log "go-micro.dev/v5/logger"
)

func (h *helper) getHealthSummary() any {
	return cubecos.GetHealthSummary(h.past)
}

func (h *helper) genServiceHealthHistory() []cubecos.HealthStatus {
	return cubecos.GetServiceHealthHistory(h.service, h.past)
}

func (h *helper) genModuleHealthHistory() cubecos.HealthStatus {
	service := cubecos.ModuleToService[h.module]
	history, err := cubecos.GetModuleHealthHistory(h.module, h.past)
	if err != nil {
		log.Errorf("request(%s): %v", api.GetReqId(h.c), err)
	}

	return cubecos.HealthStatus{
		Category:     cubecos.ServiceToCategory[service],
		Name:         service,
		Module:       h.module,
		IsRepairable: cubecos.IsRepairableModule(h.module),
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
