package healths

import (
	"fmt"

	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/status"
	log "go-micro.dev/v5/logger"
)

var (
	healthRepair       = "health: repair service %s"
	healthRepairFailed = "health: failed to repair service: %s"
	healthVerify       = "health: verify service %s"
	healthVerifyFailed = "health: failed to verify service: %s"
)

func (o *Operator) operateReq(health cubecos.Health) error {
	switch health.Overall.Status.Desired {
	case status.Ok:
		go o.applyBackgroundRepairing(health)
		return nil
	}

	return fmt.Errorf(
		"unknown desired action(%s) for health",
		health.Overall.Status.Desired,
	)
}

func (o *Operator) applyBackgroundRepairing(health cubecos.Health) {
	var err error
	health.Error, err = cubecos.GetUnhealthyServices()
	if err != nil {
		health.Overall.Status.SetCurrentToError(err)
		o.reportToController(&health)
		log.Errorf("health: failed to get unhealthy services: %s", err.Error())
		return
	}

	o.RepairServices(&health)
}

func (o *Operator) RepairServices(health *cubecos.Health) {
	moveErrorSvcToFixing(health)
	o.reportToController(health)

	result := repairAndVerify(health)
	if result.HasUnhealthyService() {
		result.Overall.Status.SetCurrentToError(nil)
	} else {
		result.Overall.Status.SetCurrentToOk()
	}

	o.reportToController(result)
}

func moveErrorSvcToFixing(health *cubecos.Health) {
	health.Fixing = append(health.Fixing, health.Error...)
	health.Error = nil
	for svcIdx, svc := range health.Fixing {
		for modIdx := range svc.Modules {
			health.Fixing[svcIdx].Modules[modIdx].Status.SetCurrentToRepairing()
		}
	}
}

func repairAndVerify(health *cubecos.Health) *cubecos.Health {
	result := health.CopyEmptyServiceStruct()

	for _, svc := range health.Fixing {
		log.Infof(healthRepair, svc.Name)
		err := cubecos.RepairServiceHealth(svc)
		if err != nil {
			log.Errorf(healthRepairFailed, err.Error())
		}

		log.Infof(healthVerify, svc.Name)
		err = cubecos.CheckServiceHealth(svc)
		if err != nil {
			log.Errorf(healthVerifyFailed, err.Error())
			result.AddError(svc, err)
			continue
		}

		result.AddOk(svc)
	}

	return &result
}
