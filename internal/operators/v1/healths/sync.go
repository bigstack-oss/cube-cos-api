package healths

import (
	"fmt"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/http"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
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
	case status.Repairing:
		go repairServices(health)
		return nil
	case status.CheckingAndRepairing:
		go checkAndRepairServices(health)
		return nil
	}

	return fmt.Errorf(
		"unknown desired action(%s) for health",
		health.Overall.Status.Desired,
	)
}

func checkAndRepairServices(health cubecos.Health) {
	var err error
	health.Services, err = cubecos.GetUnhealthyServices()
	if err != nil {
		log.Errorf("health: failed to get unhealthy services: %s", err.Error())
		return
	}

	repairServices(health)
	reportToController()
}

func repairServices(health cubecos.Health) {
	for _, svc := range health.Services {
		log.Infof(healthRepair, svc.Name)
		err := cubecos.RepairServiceHealth(svc)
		if err != nil {
			log.Errorf(healthRepairFailed, err.Error())
		}

		wait.Seconds(3)

		log.Infof(healthVerify, svc.Name)
		err = cubecos.CheckServiceHealth(svc)
		if err != nil {
			log.Warnf(healthVerifyFailed, err.Error())
		}
	}
}

func reportToController() {
	node, err := v1.GetOneOfControllerNode()
	if err != nil {
		log.Errorf("healths: failed to get controller nodes: %s", err.Error())
		return
	}

	h := http.GetGlobalHelper()
	resp, err := h.R().
		SetHeaders(v1.GenNodeAuthHeaders()).
		Delete(node.DeleteRepairingTaskUrl())
	if err != nil {
		log.Errorf("healths: failed to send repairing task to controller: %s", err.Error())
		return
	}

	if resp.IsError() {
		log.Errorf("healths: has error response from controller: %s", resp.String())
		return
	}
}
