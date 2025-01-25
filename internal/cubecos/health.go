package cubecos

import (
	"fmt"
	"os/exec"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/mongo"
	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	cuberr "github.com/bigstack-oss/cube-cos-api/internal/errors"
	"github.com/bigstack-oss/cube-cos-api/internal/status"
	log "go-micro.dev/v5/logger"
	"go.mongodb.org/mongo-driver/bson"
)

type Health struct {
	*definition.DataCenter `json:"dataCenter,omitempty" bson:"dataCenter,omitempty"`
	*Overall               `json:"overall,omitempty" bson:"overall"`
	InUse                  []definition.Service `json:"inUse" bson:"inUse"`
	Error                  []definition.Service `json:"error" bson:"error"`
	Fixing                 []definition.Service `json:"fixing" bson:"fixing"`
}

type Overall struct {
	Status status.Details `json:"status,omitempty" bson:"status"`
}

var (
	healthCheck = "health: check module %s"
)

func (h *Health) HasUnhealthyService() bool {
	return len(h.Error) > 0
}

func (h *Health) CopyEmptyServiceStruct() Health {
	return Health{
		DataCenter: h.DataCenter,
		Overall:    h.Overall,
	}
}

func (h *Health) AddOk(svc definition.Service) {
	svc.Status.SetCurrentToOk()
	h.InUse = append(h.InUse, svc)
}

func (h *Health) AddError(svc definition.Service, err error) {
	svc.Status.SetCurrentToError(err)
	h.Error = append(h.Error, svc)
}

func IsRepairing() bool {
	h := mongo.GetGlobalHelper()
	count, err := h.GetCount(
		definition.HealthDB(),
		definition.RepairCollection(),
		bson.M{
			"dataCenter.name": definition.DataCenterName,
			"overall.current": "repairing",
		},
	)
	if err != nil {
		log.Errorf("failed to get repairing count: %s", err.Error())
		return false
	}

	return count > 0
}

func IsRepairable() bool {
	if !IsClusterSetReady() {
		log.Errorf("data center is not ready for repairing")
		return false
	}

	if definition.CurrentRole == "" {
		log.Errorf("role is not set for repairing")
		return false
	}

	return true
}

func GetUnhealthyServices() ([]definition.Service, error) {
	unhealthy := map[string]definition.Service{}

	for _, service := range OrderSensitiveServices {
		for _, module := range service.Modules {
			log.Infof(healthCheck, module.Name)
			if !module.IsAutoRepairable {
				continue
			}

			if IsModuleHealthy(module.Name) {
				continue
			}

			setUnhealthyModule(unhealthy, service, module)
		}
	}

	return convertToList(unhealthy), nil
}

func setUnhealthyModule(unhealthy map[string]definition.Service, service definition.Service, module definition.Module) {
	_, found := unhealthy[service.Name]
	if !found {
		unhealthy[service.Name] = service.CopyModuleEmptyStruct()
	}

	svc := unhealthy[service.Name]
	svc.Modules = append(svc.Modules, module)
	unhealthy[service.Name] = svc
}

func convertToList(unhealthyMap map[string]definition.Service) []definition.Service {
	unhealthySvcs := []definition.Service{}
	for _, svc := range unhealthyMap {
		unhealthySvcs = append(unhealthySvcs, svc)
	}

	return unhealthySvcs
}

func IsModuleHealthy(moduleName string) bool {
	checkModule := fmt.Sprintf("health_%s_check", moduleName)
	out, err := exec.Command("hex_sdk", checkModule).CombinedOutput()
	if err == nil {
		return true
	}

	if IsExpectedEmptyStdOut(err) {
		return true
	}

	log.Errorf("found unhealthy module %s: %s", moduleName, string(out))
	return false
}

func RepairServiceHealth(service definition.Service) error {
	errs := []error{}

	for _, module := range service.Modules {
		err := RepairModule(module.Name)
		if err != nil {
			errs = append(errs, err)
		}
	}

	return cuberr.CombineErrors(errs)
}

func CheckServiceHealth(service definition.Service) error {
	errs := []error{}
	for _, module := range service.Modules {
		if !IsModuleHealthy(module.Name) {
			errs = append(errs, fmt.Errorf("%s is still unhealthy", module.Name))
		}
	}

	return cuberr.CombineErrors(errs)
}

func RepairModule(moduleName string) error {
	repairModule := fmt.Sprintf("health_%s_repair", moduleName)
	out, err := exec.Command("hex_sdk", repairModule).CombinedOutput()
	if err == nil {
		return nil
	}

	if IsExpectedEmptyStdOut(err) {
		return nil
	}

	return fmt.Errorf(
		"failed to repair module %s: %s",
		moduleName,
		string(out),
	)
}
