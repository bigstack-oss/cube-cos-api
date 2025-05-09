package cubecos

import (
	"context"
	"fmt"
	"os/exec"
	"slices"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/aws"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/influx"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/base"
	cuberr "github.com/bigstack-oss/cube-cos-api/internal/definition/v1/errors"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/event"
	"github.com/bigstack-oss/cube-cos-api/internal/status"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/api/query"
	json "github.com/json-iterator/go"
	"github.com/mohae/deepcopy"
	log "go-micro.dev/v5/logger"
)

const (
	healthMeasurement     = `fn: (r) => r._measurement == "health"`
	convertValueToField   = `rowKey: ["_time","component","node","code"], columnKey: ["_field"], valueColumn: "_value"`
	repairingCode         = 0
	defaultAggreateWindow = 10 * time.Minute
)

var (
	updateHealthSummary sync.Mutex
	healthSummary       Health
	healthCheck         = "health: check module %s"
)

type Health struct {
	*v1.DataCenter `json:"dataCenter,omitempty" bson:"dataCenter,omitempty"`
	*Overall       `json:"overall,omitempty" bson:"overall"`
	Services       []v1.Service `json:"services" bson:"services"`
}

type Overall struct {
	Status status.Health `json:"status" bson:"status"`
}

type ModuleHealth struct {
	Category     string           `json:"category"`
	Name         string           `json:"name"`
	Module       string           `json:"module"`
	IsRepairable bool             `json:"isRepairable"`
	History      []v1.HealthCheck `json:"history"`
	Status       status.Health    `json:"status"`
}

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

func (h *Health) SetRepairingStatus(service v1.Service) {
	h.Status.Current = status.Repairing
	h.Status.IsFixing = true
	h.Status.Description = service.Status.Description
}

func IsRepairing() bool {
	_, err := exec.Command("hex_sdk", "is_repairing").Output()
	return isRepairingCode(err)
}

func IsRepairable() bool {
	if !IsClusterSetReady() {
		log.Errorf("data center is not ready for repairing")
		return false
	}

	if base.CurrentRole == "" {
		log.Errorf("role is not set for repairing")
		return false
	}

	return true
}

func GetUnhealthyServices() ([]v1.Service, error) {
	unhealthy := map[string]v1.Service{}

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

func setUnhealthyModule(unhealthy map[string]v1.Service, service v1.Service, module v1.Module) {
	_, found := unhealthy[service.Name]
	if !found {
		unhealthy[service.Name] = service.CopyModuleEmptyStruct()
	}

	svc := unhealthy[service.Name]
	svc.Modules = append(svc.Modules, module)
	unhealthy[service.Name] = svc
}

func convertToList(unhealthyMap map[string]v1.Service) []v1.Service {
	unhealthySvcs := []v1.Service{}
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

	log.Errorf("healths: found unhealthy module %s(%s)", moduleName, string(out))
	return false
}

func RepairServiceHealth(service v1.Service) error {
	errs := []error{}

	for _, module := range service.Modules {
		err := RepairModule(module.Name)
		if err != nil {
			errs = append(errs, err)
		}
	}

	return cuberr.CombineErrors(errs)
}

func CheckServiceHealth(service v1.Service) error {
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
		"healths: failed to repair module %s: %s",
		moduleName,
		string(out),
	)
}

func SyncHealthHistory() {
	updateHealthSummary.Lock()
	defer updateHealthSummary.Unlock()

	services := GetServicesToCheckHealth()
	syncServiceHealth(&services, "24h")
	healthSummary = genHealthSummary(services)
}

func GetHealthSummary() Health {
	return healthSummary
}

func GetServiceHealthHistory(serviceName, duration string) []ModuleHealth {
	modules := ServiceToModules[serviceName]
	statuses := []ModuleHealth{}

	for _, module := range modules {
		history, err := GetModuleHealthHistory(module.Name, duration, v1.AscSort, false)
		if err != nil {
			continue
		}

		statuses = append(statuses, ModuleHealth{
			Category:     ServiceToCategory[serviceName],
			Name:         serviceName,
			Module:       module.Name,
			IsRepairable: IsRepairableModule(module.Name),
			History:      history,
		})
	}

	return statuses
}

func GetModuleHealthHistory(moduleName, duration, order string, onlyLast bool) ([]v1.HealthCheck, error) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(60))
	defer cancel()
	stmt := GenModuleHealthHistoryQuery(moduleName, duration, order, onlyLast)
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

	checks = aggregateHealthsByTime(checks, defaultAggreateWindow)
	SetUnhealthLogUrl(&checks)
	return checks, nil
}

func SetUnhealthLogUrl(history *[]v1.HealthCheck) {
	for i, check := range *history {
		if !check.IsNg() {
			continue
		}

		if check.Error == nil {
			continue
		}

		if check.Error.Log == "" {
			continue
		}

		setPresignedUrl(&(*history)[i])
	}
}

func setPresignedUrl(check *v1.HealthCheck) {
	h := aws.GetGlobalHelper()
	url, err := h.GenPresignedUrl("log", genHealthLogKey(check.Error.Log), v1.Day*7)
	if err != nil {
		log.Errorf("healths: failed to generate presigned url: %v", err)
		return
	}

	check.Error.Log = url
}

func genHealthLogKey(log string) string {
	return strings.TrimPrefix(log, "s3://log/")
}

func aggregateHealthsByTime(checks []v1.HealthCheck, duration time.Duration) []v1.HealthCheck {
	if len(checks) == 0 {
		return []v1.HealthCheck{}
	}

	grouped := groupHealthsByTime(checks, duration)
	keys := getSortedTimeKeys(grouped)
	aggregated := []v1.HealthCheck{}
	for i, key := range keys {
		if i == 0 {
			firstCheck := backfillFirstCheck(key)
			aggregated = append(aggregated, firstCheck)
		}

		picked := pickHealthOrUnhealthy(grouped[key])
		aggregated = append(aggregated, picked)
	}

	return aggregated
}

func backfillFirstCheck(t time.Time) v1.HealthCheck {
	date, err := time.Parse(event.TimeLayout, t.Add(-defaultAggreateWindow).Local().String())
	if err != nil {
		return v1.HealthCheck{}
	}

	return v1.HealthCheck{
		Time:   v1.TimeRFC3339Z(date),
		Status: status.Ok,
	}
}

func groupHealthsByTime(checks []v1.HealthCheck, duration time.Duration) map[time.Time][]v1.HealthCheck {
	grouped := make(map[time.Time][]v1.HealthCheck)

	for _, check := range checks {
		t, err := time.Parse(time.RFC3339, check.Time)
		if err != nil {
			continue
		}

		key := t.Truncate(duration)
		grouped[key] = append(grouped[key], check)
	}

	return grouped
}

func getSortedTimeKeys(grouped map[time.Time][]v1.HealthCheck) []time.Time {
	keys := make([]time.Time, 0, len(grouped))
	for key := range grouped {
		keys = append(keys, key)
	}

	sort.Slice(keys, func(i, j int) bool {
		return keys[i].Before(keys[j])
	})

	return keys
}

func pickHealthOrUnhealthy(group []v1.HealthCheck) v1.HealthCheck {
	for _, check := range group {
		if check.IsNg() {
			return check
		}
	}

	return group[0]
}

func GenModuleHealthHistoryQuery(moduleName, past, order string, onlyLast bool) string {
	query := influx.Query{}
	query.Bucket("events").
		Range(genTimeDuration(past)).
		Filter(healthMeasurement).
		Filter(genModuleFilter(moduleName)).
		Pivot(convertValueToField).
		Group("").
		Sort(order)

	if onlyLast {
		query.Limit("n: 1")
	}

	return query.String()
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
	desc := parseHealthResult(record)
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

func parseHealthResult(record *query.FluxRecord) string {
	code := record.ValueByKey("code")
	code, ok := code.(string)
	if !ok {
		return "ng"
	}

	desc := record.ValueByKey("description")
	desc, ok = desc.(string)
	if !ok {
		return "ng"
	}

	isOkOrFixingDesc := desc == "ok" || desc == "fixing"
	if code != "0" && !isOkOrFixingDesc {
		return status.Ng
	}

	return status.Ok
}

func GetServicesToCheckHealth() []v1.Service {
	services := deepcopy.Copy(OrderSensitiveServices).([]v1.Service)
	for i := range services {
		if services[i].IsInternalViewOnly {
			services = slices.Delete(services, i, i+1)
			continue
		}

		services[i].Status = status.NewHealthOk()
	}

	return services
}

func GetRepairingInfo() (*v1.ReairingInfo, error) {
	b, err := exec.Command("hex_sdk", "-v", "is_repairing").Output()
	if err != nil {
		return nil, err
	}

	info := v1.ReairingInfo{}
	err = json.Unmarshal(b, &info)
	if err != nil {
		log.Errorf("healths: failed to unmarshal repairing info: %s", err.Error())
		return nil, err
	}

	return &info, nil
}

func syncServiceHealth(services *[]v1.Service, duration string) {
	for s, service := range *services {
		service.InitOkStatus()

		for m, module := range service.Modules {
			history, err := GetModuleHealthHistory(module.Name, duration, v1.DescSort, true)
			if err != nil {
				continue
			}

			module.InitOkStatus()
			if isLastCheckUnhealthy(history) {
				module.SetUnhealthyStatus()
				service.ConvergeUnhealthyStatus(module.Name)
			}

			service.Modules[m] = module
		}

		(*services)[s] = service
	}
}

func genHealthSummary(services []v1.Service) Health {
	health := Health{Services: services}
	health.Overall = &Overall{Status: status.Health{Current: status.Ok}}
	syncUnhealthStatus(&health, services)
	syncRepairingStatus(&health, &services)
	return health
}

func syncUnhealthStatus(health *Health, services []v1.Service) {
	for _, service := range services {
		if !service.IsStatusOk() {
			health.Status.Current = status.Ng
			if health.Status.Description == "" {
				health.Status.Description = fmt.Sprintf("%s %s", service.Name, service.Status.Description)
			} else {
				health.Status.Description += fmt.Sprintf(", %s %s", service.Name, service.Status.Description)
			}
		}
	}
}

func syncRepairingStatus(health *Health, services *[]v1.Service) {
	if !IsRepairing() {
		return
	}

	info, err := GetRepairingInfo()
	if err != nil {
		return
	}

	for s, service := range *services {
		for m, module := range service.Modules {
			if module.Name != info.Service {
				continue
			}

			module.SetRepairingStatus()
			service.Modules[m] = module
			service.SetRepairingStatus(*info)
			health.SetRepairingStatus(service)
		}

		(*services)[s] = service
	}
}

func isLastCheckUnhealthy(history []v1.HealthCheck) bool {
	if len(history) == 0 {
		return false
	}

	lastCheck := history[len(history)-1]
	return lastCheck.IsNg()
}

func parseDetails(record *query.FluxRecord) string {
	details := record.ValueByKey("detail")
	val, ok := details.(string)
	if !ok {
		val = ""
	}

	return strings.ReplaceAll(val, `;`, "\n")
}

func parseTime(record *query.FluxRecord) string {
	date, err := time.Parse(event.TimeLayout, record.Time().Local().String())
	if err != nil {
		log.Debugf("failed to parse date from record: %v", record)
	}

	return v1.TimeRFC3339Z(date)
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

// note:
// in COS's current design, the return code of under repairing is 0
// and in the exec.command, the err will be nil when return code is 0
// so that's why we use err == nil to identify if cos is repairing
func isRepairingCode(err error) bool {
	if err == nil {
		return true
	}

	result, ok := err.(*exec.ExitError)
	if !ok {
		return false
	}

	return result.ExitCode() == repairingCode
}
