package cubecos

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"

	"slices"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/influx"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	cuberr "github.com/bigstack-oss/cube-cos-api/internal/errors"
	"github.com/bigstack-oss/cube-cos-api/internal/status"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/api/query"
	"github.com/mohae/deepcopy"
	log "go-micro.dev/v5/logger"
)

const (
	healthMeasurement   = `fn: (r) => r._measurement == "health"`
	convertValueToField = `rowKey: ["_time","component","node","code"], columnKey: ["_field"], valueColumn: "_value"`
	descByTime          = `columns: ["_time"], desc: true`
	repairingCode       = 1
)

type Health struct {
	*definition.DataCenter `json:"dataCenter,omitempty" bson:"dataCenter,omitempty"`
	*Overall               `json:"overall,omitempty" bson:"overall"`
	Services               []definition.Service `json:"services" bson:"services"`
}

type HealthStatus struct {
	Category     string           `json:"category"`
	Name         string           `json:"name"`
	Module       string           `json:"module"`
	IsRepairable bool             `json:"isRepairable"`
	History      []v1.HealthCheck `json:"history"`
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

func IsRepairing() bool {
	_, err := exec.Command("hex_sdk", "is_repairing").Output()
	log.Infof("repairing: %v", isRepairingCode(err))
	return isRepairingCode(err)
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

func GetHealthSummary(duration string) Health {
	services := GetServicesToCheckHealth()
	syncServiceHealth(&services, duration)
	return genHealthSummary(services)
}

func GetServiceHealthHistory(serviceName, duration string) []HealthStatus {
	modules := ServiceToModules[serviceName]
	statuses := []HealthStatus{}

	for _, module := range modules {
		history, err := GetModuleHealthHistory(module.Name, duration)
		if err != nil {
			continue
		}

		statuses = append(statuses, HealthStatus{
			Category:     ServiceToCategory[serviceName],
			Name:         serviceName,
			Module:       module.Name,
			IsRepairable: IsRepairableModule(module.Name),
			History:      history,
		})
	}

	return statuses
}

func GetModuleHealthHistory(moduleName, duration string) ([]v1.HealthCheck, error) {
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
	checks := []v1.HealthCheck{}
	err = parseHealthCheck(c, &checks)
	if err != nil {
		log.Errorf("healths: failed to parse events from cursor: %v", err)
		return nil, err
	}

	return checks, nil
}

func GenModuleHealthHistoryQuery(moduleName, past string) string {
	query := influx.Query{}
	return query.Bucket("events").
		Range(genTimeDuration(past)).
		Filter(healthMeasurement).
		Filter(genModuleFilter(moduleName)).
		Pivot(convertValueToField).
		Group("").
		Sort(descByTime).
		String()
}

func genModuleFilter(modulName string) string {
	return fmt.Sprintf(`fn: (r) => r.component == "%s"`, modulName)
}

func genTimeDuration(past string) string {
	return fmt.Sprintf("start: -%s", past)
}

func parseHealthCheck(c *api.QueryTableResult, checks *[]v1.HealthCheck) error {
	for c.Next() {
		*checks = append(*checks, genHealthCheckByRecord(c.Record()))
		break
	}
	if c.Err() != nil {
		return c.Err()
	}

	return nil
}

func genHealthCheckByRecord(record *query.FluxRecord) v1.HealthCheck {
	healthCheck := v1.HealthCheck{Time: parseTime(record)}
	syncStatusDetails(record, &healthCheck)
	return healthCheck
}

func syncStatusDetails(record *query.FluxRecord, check *v1.HealthCheck) {
	desc := parseDescription(record)
	if desc == status.Ok {
		check.Status = status.Ok
		return
	}

	check.Status = status.Ng
	check.Error = &v1.Error{
		Type:        fmt.Sprintf("%s failure", record.ValueByKey("component").(string)),
		Reason:      record.ValueByKey("description").(string),
		Description: fmt.Sprintf("there's a failure was detected from node %s, please see the detail or log to know more", record.ValueByKey("node").(string)),
		Details:     parseDetails(record),
		Nodes:       parseNodes(record),
		Log:         parseLog(record),
	}
}

func GetServicesToCheckHealth() []definition.Service {
	services := deepcopy.Copy(OrderSensitiveServices).([]definition.Service)
	for i := range services {
		if services[i].IsInternalViewOnly {
			services = slices.Delete(services, i, i+1)
			continue
		}

		services[i].Status = status.NewOk()
	}

	return services
}

func syncServiceHealth(services *[]definition.Service, duration string) {
	for s, service := range *services {
		for m, module := range service.Modules {
			history, err := GetModuleHealthHistory(module.Name, duration)
			if err != nil {
				continue
			}

			module.InitOkStatus()
			record := captureUnhealthyRecord(history)
			if record != nil {
				module.SetUnhealthyStatus(record)
				service.ConvergeUnhealthyStatus(module.Name, record)
			}

			service.Modules[m] = module
		}

		(*services)[s] = service
	}
}

func genHealthSummary(services []definition.Service) Health {
	health := Health{Services: services}
	health.Overall = &Overall{Status: status.Health{Current: status.Ok}}
	syncUnhealthStatus(&health, services)
	syncRepairingStatus(&health, &services)
	return health
}

func syncUnhealthStatus(health *Health, services []definition.Service) {
	unhealthDesc := "failure services detected: "
	unhealthFound := false
	for _, service := range services {
		if !service.IsStatusOk() {
			unhealthFound = true
			unhealthDesc += fmt.Sprintf("%s(%s) ", service.Name, service.Status.Description)
		}
	}

	if unhealthFound {
		health.Status.Current = status.Ng
		health.Status.Description = unhealthDesc
	}
}

func syncRepairingStatus(health *Health, services *[]definition.Service) {
	if !IsRepairing() {
		return
	}

	b, err := exec.Command("hex_sdk", "-v", "is_repairing").Output()
	if err != nil {
		log.Errorf("healths: failed to get repairing info: %s", err.Error())
		return
	}

	repairingInfo := v1.ReairingInfo{}
	err = json.Unmarshal(b, &repairingInfo)
	if err != nil {
		log.Errorf("healths: failed to unmarshal repairing info: %s", err.Error())
		return
	}

	for s, service := range *services {
		for m, module := range service.Modules {
			if module.Name == repairingInfo.Service {
				module.SetRepairingStatus()
				service.SetRepairingStatus(repairingInfo)
				service.Modules[m] = module
				health.Status.Current = status.Repairing
				health.Status.IsFixing = true
				health.Status.Description = service.Status.Description
			}
		}

		(*services)[s] = service
	}
}

func captureUnhealthyRecord(history []v1.HealthCheck) *v1.HealthCheck {
	if len(history) == 0 {
		return nil
	}

	for _, check := range history {
		if check.Status != status.Ok {
			return &check
		}
	}

	return nil
}

func parseDetails(record *query.FluxRecord) string {
	details := record.ValueByKey("detail")
	val, ok := details.(string)
	if !ok {
		val = ""
	}

	return val
}

func parseTime(record *query.FluxRecord) string {
	date, err := time.Parse(eventTimeLayout, record.Time().Local().String())
	if err != nil {
		log.Debugf("failed to parse date from record: %v", record)
	}

	return definition.TimeRFC3339Z(date)
}

func parseDescription(record *query.FluxRecord) string {
	desc := record.ValueByKey("description")
	val, ok := desc.(string)
	if !ok {
		val = ""
	}

	return val
}

func parseNodes(record *query.FluxRecord) []string {
	node := record.ValueByKey("node")
	val, ok := node.(string)
	if !ok {
		return []string{}
	}

	return []string{val}
}

func parseLog(record *query.FluxRecord) string {
	log := record.ValueByKey("log")
	val, ok := log.(string)
	if !ok {
		val = ""
	}

	return val
}

func isRepairingCode(err error) bool {
	if err == nil {
		log.Infof("repairing: no repairing code found")
		return false
	}

	result, ok := err.(*exec.ExitError)
	if !ok {
		return false
	}

	return result.ExitCode() == repairingCode
}
