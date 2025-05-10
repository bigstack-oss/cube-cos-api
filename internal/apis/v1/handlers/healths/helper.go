package healths

import (
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/mongo"
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/queries"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"
	"github.com/gin-gonic/gin"
	log "go-micro.dev/v5/logger"
	"go.mongodb.org/mongo-driver/bson"
)

type helper struct {
	c       *gin.Context
	handler string

	serviceType string
	moduleType  string
	module      *v1.Module

	period *v1.Period
	past   string

	watch bool
}

func initHelper(c *gin.Context, handler string) (*helper, error) {
	h := &helper{c: c, handler: handler}
	return h, h.parseParamsByHandler()
}

func (h *helper) getHealthSummary() (cubecos.Health, error) {
	summary := cubecos.GetHealthSummary()
	h.syncRepairingStatus(&summary)
	return summary, nil
}

func (h *helper) genServiceHealthHistory() []cubecos.ModuleHealth {
	return cubecos.GetServiceHealthHistory(h.serviceType, h.past)
}

func (h *helper) genModuleHealthHistory() cubecos.ModuleHealth {
	service := cubecos.ModuleToService[h.moduleType]
	history, err := cubecos.GetModuleHealthHistory(h.moduleType, h.past, v1.AscSort, false)
	if err != nil {
		log.Errorf("healths(%s): %v", queries.GetReqId(h.c), err)
	}

	return cubecos.ModuleHealth{
		Category:     cubecos.ServiceToCategory[service],
		Name:         service,
		Module:       h.moduleType,
		IsRepairable: cubecos.IsRepairableModule(h.moduleType),
		History:      history,
		Status:       h.getModuleStatus(),
	}
}

func (h *helper) genCheckRepairReq() cubecos.Health {
	health := &cubecos.Health{}
	health.Overall = &cubecos.Overall{}
	health.Overall.Status.SetDesiredToCheckingAndRepairing()
	return cubecos.Health{
		Overall: &cubecos.Overall{
			Status: status.Health{
				Desired: status.CheckingAndRepairing,
			},
		},
	}
}

func (h *helper) genForceRepairReq() cubecos.Health {
	svc := cubecos.ModuleToService[h.module.Name]
	return cubecos.Health{
		Overall: &cubecos.Overall{
			Status: status.Health{
				Desired: status.Repairing,
			},
		},
		Services: []v1.Service{
			{
				Name:     svc,
				Category: cubecos.ServiceToCategory[svc],
				Modules:  []v1.Module{*h.module},
			},
		},
	}
}

func (h *helper) requestForceRepair() {
	req := h.genForceRepairReq()
	reqQueue.Add(&req)
	err := h.setMoudleRepairingRecord()
	if err != nil {
		log.Errorf("healths(%s): failed to set module repairing record: %v", queries.GetReqId(h.c), err)
	}
}

func (h *helper) requestCheckRepair() {
	req := h.genCheckRepairReq()
	reqQueue.Add(&req)
	err := h.setRepairingRecord()
	if err != nil {
		log.Errorf("healths(%s): failed to set repairing record: %v", queries.GetReqId(h.c), err)
	}
}

func (h *helper) deleteCheckRepairTask() error {
	mongo := mongo.GetGlobalHelper()
	return mongo.DeleteAll(
		v1.Healths,
		v1.HealthRepairingCollection,
		bson.M{"isRepairing": true},
	)
}

func (h *helper) deleteModuleCheckRepairTask() error {
	mongo := mongo.GetGlobalHelper()
	return mongo.DeleteAll(
		v1.Healths,
		v1.HealthRepairingCollection,
		bson.M{"type": "forceRepair", "module": h.moduleType, "isRepairing": true},
	)
}
