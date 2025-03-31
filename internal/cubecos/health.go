package cubecos

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/influx"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	cuberr "github.com/bigstack-oss/cube-cos-api/internal/errors"
	"github.com/bigstack-oss/cube-cos-api/internal/status"
	"github.com/influxdata/influxdb-client-go/v2/api"
	log "go-micro.dev/v5/logger"
)

const (
	healthMeasurement   = `fn: (r) => r._measurement == "health"`
	convertValueToField = `rowKey: ["_time","component","node","code"], columnKey: ["_field"], valueColumn: "_value"`
	descByTime          = `columns: ["_time"], desc: true`
)

type Health struct {
	*definition.DataCenter `json:"dataCenter,omitempty" bson:"dataCenter,omitempty"`
	*Overall               `json:"overall,omitempty" bson:"overall"`
	Services               []definition.Service `json:"services" bson:"services"`
}

type HealthStatus struct {
	Category     string        `json:"category"`
	Name         string        `json:"name"`
	Module       string        `json:"module"`
	IsRepairable bool          `json:"isRepairable"`
	History      []HealthCheck `json:"history"`
}

type HealthCheck struct {
	Time   string `json:"time"`
	Status string `json:"status"`
	*Error `json:"error,omitempty"`
}

type HealthPoint struct {
	Time        string `json:"time"`
	Code        int    `json:"code"`
	Component   string `json:"component"`
	Description string `json:"description"`
	Details     string `json:"details"`
	Log         string `json:"log"`
	Node        string `json:"node"`
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
	Status status.Health `json:"status" bson:"status"`
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
			if !module.IsRepairable {
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

func ListModuleHealth(duration string) ([]HealthStatus, error) {
	health := []HealthStatus{}

	// for _, service := range OrderSensitiveServices {
	// 	for _, module := range service.Modules {
	// 		history, err := GetModuleHealthHistory(module.Name, duration)
	// 		if err != nil {
	// 			continue
	// 		}
	// 	}
	// }

	return health, nil
}

func GetModuleHealthHistory(moduleName, duration string) ([]HealthPoint, error) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(60))
	defer cancel()
	stmt := GenModuleHealthHistoryQuery(moduleName, duration)

	h := influx.GetGlobalHelper()
	c, err := h.QueryApiClient.Query(ctx, stmt)
	if err != nil {
		log.Errorf("healths: failed to get query cursor: %v", err)
		return nil, err
	}

	defer c.Close()
	points := []HealthPoint{}
	err = parseHealthPoints(c, &points)
	if err != nil {
		log.Errorf("healths: failed to parse events from cursor: %v", err)
		return nil, err
	}

	return points, nil

}

func GenModuleHealthHistoryQuery(moduleName, past string) string {
	query := influx.Query{}
	return query.Bucket("event").
		Range(genTimeDuration(past)).
		Filter(healthMeasurement).
		Filter(genModuleFilter(moduleName)).
		Pivot(convertValueToField).
		Group("").
		Sort(descByTime).
		String()
}

func genModuleFilter(modulName string) string {
	return fmt.Sprintf(`fn: (r) => r.component == %q`, modulName)
}

func genTimeDuration(past string) string {
	return fmt.Sprintf("start: -%s", past)
}

func parseHealthPoints(c *api.QueryTableResult, points *[]HealthPoint) error {
	for c.Next() {
		record := c.Record()
		*points = append(
			*points,
			HealthPoint{
				Time:        record.Time().String(),
				Code:        int(record.ValueByKey("code").(int64)),
				Component:   record.ValueByKey("component").(string),
				Description: record.ValueByKey("description").(string),
				Details:     record.ValueByKey("detail").(string),
				Log:         record.ValueByKey("log").(string),
				Node:        record.ValueByKey("node").(string),
			},
		)
	}
	if c.Err() != nil {
		return c.Err()
	}

	return nil
}
