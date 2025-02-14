package cubecos

import (
	"fmt"
	"os/exec"

	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	cuberr "github.com/bigstack-oss/cube-cos-api/internal/errors"
	"github.com/bigstack-oss/cube-cos-api/internal/status"
	log "go-micro.dev/v5/logger"
)

type Health struct {
	*definition.DataCenter `json:"dataCenter,omitempty" bson:"dataCenter,omitempty"`
	*Overall               `json:"overall,omitempty" bson:"overall"`
	Services               []definition.Service `json:"services" bson:"services"`
}

type HealthStatus struct {
	Category string        `json:"category"`
	Service  string        `json:"service"`
	Module   string        `json:"module"`
	History  []HealthCheck `json:"history"`
}

type HealthCheck struct {
	Time   string `json:"time"`
	Status string `json:"status"`
	*Error `json:"error,omitempty"`
}

type Error struct {
	Type        string   `json:"type"`
	Reason      string   `json:"reason"`
	Nodes       []string `json:"nodes"`
	Description string   `json:"description"`
	Details     string   `json:"details"`
	Log         string   `json:"log"`
}

type Overall struct {
	Status status.Details `json:"status,omitempty" bson:"status"`
}

var (
	healthCheck = "health: check module %s"
)

func (h *Health) HasUnhealthyService() bool {
	for _, svc := range h.Services {
		if svc.Status.Current != status.Ok {
			return true
		}
	}

	return false
}

func (h *Health) CopyEmptyServiceStruct() Health {
	return Health{
		DataCenter: h.DataCenter,
		Overall:    h.Overall,
	}
}

// M1 TODO:
// Waiting for COS developer to implement the /var/run/{markerfile} to check if the data center is repairing.
func IsRepairing() bool {
	return false
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
