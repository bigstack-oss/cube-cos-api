package cubecos

import (
	"context"
	"fmt"
	"os/exec"
	"slices"
	"sort"
	"strings"
	"sync/atomic"
	ostime "time"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/aws"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/influx"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/base"
	cuberr "github.com/bigstack-oss/cube-cos-api/internal/definition/v1/errors"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/events"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/health"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/services"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/time"
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
	defaultAggreateWindow = 10 * ostime.Minute
)

var (
	healthSummary = atomic.Pointer[Health]{}
	healthCheck   = "health: check module %s"
)

type Health struct {
	*base.DataCenter `json:"dataCenter,omitempty" bson:"dataCenter,omitempty"`
	*Overall         `json:"overall,omitempty" bson:"overall"`
	Services         []services.Service `json:"services" bson:"services"`
}

type Overall struct {
	Status status.Health `json:"status" bson:"status"`
}

type ModuleHealth struct {
	Category     string         `json:"category"`
	Name         string         `json:"name"`
	Module       string         `json:"module"`
	IsRepairable bool           `json:"isRepairable"`
	History      []health.Check `json:"history"`
	Status       status.Health  `json:"status"`
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

func (h *Health) SetRepairingStatus(service services.Service) {
	h.Status.Current = status.Repairing
	h.Status.IsFixing = true
	h.Status.Description = service.Status.Description
}

func IsRepairing() bool {
	_, err := exec.Command("hex_sdk", "is_repairing").Output()
	return isRepairingCode(err)
}

func IsRepairable() bool {
	if !IsDataCenterReady() {
		log.Errorf("healths: data center is not ready for repairing")
		return false
	}

	if base.CurrentRole == "" {
		log.Errorf("healths: role is not set for repairing")
		return false
	}

	return true
}

func GetUnhealthyServices() ([]services.Service, error) {
	unhealthy := map[string]services.Service{}

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

func setUnhealthyModule(unhealthy map[string]services.Service, service services.Service, module services.Module) {
	_, found := unhealthy[service.Name]
	if !found {
		unhealthy[service.Name] = service.CopyModuleEmptyStruct()
	}

	svc := unhealthy[service.Name]
	svc.Modules = append(svc.Modules, module)
	unhealthy[service.Name] = svc
}

func convertToList(unhealthyMap map[string]services.Service) []services.Service {
	unhealthySvcs := []services.Service{}
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

func RepairServiceHealth(service services.Service) error {
	errs := []error{}

	for _, module := range service.Modules {
		err := RepairModule(module.Name)
		if err != nil {
			errs = append(errs, err)
		}
	}

	return cuberr.CombineErrors(errs)
}

func CheckServiceHealth(service services.Service) error {
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
	services := GetServicesToCheckHealth()
	syncServiceHealth(&services, "24h")
	summary := genHealthSummary(services)
	healthSummary.Swap(&summary)
}

func GetHealthSummary() Health {
	summary := healthSummary.Load()
	if summary == nil {
		return Health{}
	}

	return *summary
}

func GetServiceHealthHistory(serviceName, duration string) []ModuleHealth {
	modules := ServiceToModules[serviceName]
	statuses := []ModuleHealth{}

	for _, module := range modules {
		history, err := GetModuleHealthHistoryWithParams(module.Name, duration, health.AscSort, false)
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

func GetModuleHealthHistory(stmt string, needsAggregate bool) ([]health.Check, error) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(60))
	defer cancel()
	h := influx.GetGlobalHelper()
	c, err := h.QueryApiClient.Query(ctx, stmt)
	if err != nil {
		log.Errorf("healths: failed to get query cursor(%v)", err)
		return nil, err
	}

	defer c.Close()
	checks := []health.Check{}
	err = parseHealthCheck(c, &checks)
	if err != nil {
		log.Errorf("healths: failed to parse events from cursor(%v)", err)
		return nil, err
	}

	if needsAggregate {
		checks = aggregateHealthsByTime(checks, defaultAggreateWindow)
	}

	setUnhealthLogUrl(&checks)
	return checks, nil
}

func GetModuleHealthHistoryWithParams(module, duration, order string, onlyLast bool) ([]health.Check, error) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(60))
	defer cancel()
	stmt := GenModuleHealthHistoryQuery(module, duration, order, onlyLast)
	h := influx.GetGlobalHelper()
	c, err := h.QueryApiClient.Query(ctx, stmt)
	if err != nil {
		log.Errorf("healths: failed to get query cursor(%v)", err)
		return nil, err
	}

	defer c.Close()
	checks := []health.Check{}
	err = parseHealthCheck(c, &checks)
	if err != nil {
		log.Errorf("healths: failed to parse events from cursor(%v)", err)
		return nil, err
	}

	checks = aggregateHealthsByTime(checks, defaultAggreateWindow)
	setUnhealthLogUrl(&checks)
	return checks, nil
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

func GetServicesToCheckHealth() []services.Service {
	services := deepcopy.Copy(OrderSensitiveServices).([]services.Service)
	for i := range services {
		if services[i].IsInternalViewOnly {
			services = slices.Delete(services, i, i+1)
			continue
		}

		services[i].Status = status.NewHealthOk()
	}

	return services
}

func GetRepairingInfo() (*services.ReairingInfo, error) {
	b, err := exec.Command("hex_sdk", "-v", "is_repairing").Output()
	if err != nil {
		return nil, err
	}

	info := services.ReairingInfo{}
	err = json.Unmarshal(b, &info)
	if err != nil {
		log.Errorf("healths: failed to unmarshal repairing info(%v)", err)
		return nil, err
	}

	return &info, nil
}

func setUnhealthLogUrl(history *[]health.Check) {
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

func setPresignedUrl(check *health.Check) {
	h := aws.GetGlobalHelper()
	url, err := h.GenPresignedUrl("log", genHealthLogKey(check.Error.Log), time.Day*7)
	if err != nil {
		log.Errorf("healths: failed to generate presigned url(%v)", err)
		return
	}

	check.Error.Log = url
}

func genHealthLogKey(log string) string {
	return strings.TrimPrefix(log, "s3://log/")
}

func aggregateHealthsByTime(checks []health.Check, duration ostime.Duration) []health.Check {
	if len(checks) == 0 {
		return []health.Check{}
	}

	grouped := groupHealthsByTime(checks, duration)
	keys := getSortedTimeKeys(grouped)
	aggregated := []health.Check{}
	for i, key := range keys {
		if i == 0 {
			firstCheck := backfillFirstCheck(key)
			aggregated = append(aggregated, firstCheck)
		}

		picked := pickFixingOrOkOrNg(grouped[key])
		aggregated = append(aggregated, picked)
	}

	return aggregated
}

func backfillFirstCheck(t ostime.Time) health.Check {
	date, err := ostime.Parse(events.TimeLayout, t.Add(-defaultAggreateWindow).Local().String())
	if err != nil {
		return health.Check{}
	}

	return health.Check{
		Time:   time.RFC3339Z(date),
		Status: status.Ok,
	}
}

func groupHealthsByTime(checks []health.Check, duration ostime.Duration) map[ostime.Time][]health.Check {
	grouped := make(map[ostime.Time][]health.Check)

	for _, check := range checks {
		t, err := ostime.Parse(time.FormatRFC3339, check.Time)
		if err != nil {
			continue
		}

		key := t.Truncate(duration)
		grouped[key] = append(grouped[key], check)
	}

	return grouped
}

func getSortedTimeKeys(grouped map[ostime.Time][]health.Check) []ostime.Time {
	keys := make([]ostime.Time, 0, len(grouped))
	for key := range grouped {
		keys = append(keys, key)
	}

	sort.Slice(keys, func(i, j int) bool {
		return keys[i].Before(keys[j])
	})

	return keys
}

func pickFixingOrOkOrNg(group []health.Check) health.Check {
	for _, check := range group {
		if check.IsFix() {
			check.Status = status.Fixing
			return check
		}

		if check.IsNg() {
			check.Status = status.Ng
			return check
		}
	}

	check := group[0]
	check.Status = status.Ok
	return check
}

func genModuleFilter(modulName string) string {
	return fmt.Sprintf(`fn: (r) => r.component == "%s"`, modulName)
}

func genTimeDuration(past string) string {
	return fmt.Sprintf("start: -%s", past)
}

func parseHealthCheck(c *api.QueryTableResult, checks *[]health.Check) error {
	for c.Next() {
		*checks = append(
			*checks,
			genHealthCheckByRecord(c.Record()),
		)
	}
	if c.Err() != nil {
		return c.Err()
	}

	return nil
}

func genHealthCheckByRecord(record *query.FluxRecord) health.Check {
	healthCheck := health.Check{
		Time:     parseTime(record),
		Hostname: parseNode(record),
		Status:   parseHealthResult(record),
	}

	syncStatusDetails(record, &healthCheck)
	return healthCheck
}

func syncStatusDetails(record *query.FluxRecord, check *health.Check) {
	if !check.IsNg() {
		return
	}

	check.Error = &health.Error{
		Type:        fmt.Sprintf("%s failure", record.ValueByKey("component").(string)),
		Reason:      record.ValueByKey("description").(string),
		Description: fmt.Sprintf("there's a failure was detected from node %s, please see the detail or log to know more", record.ValueByKey("node").(string)),
		Details:     parseDetails(record),
		Nodes:       parseNodes(record),
		Log:         parseLog(record),
	}
}

func parseHealthResult(record *query.FluxRecord) string {
	val := record.ValueByKey("description")
	desc, ok := val.(string)
	if !ok {
		return status.Ng
	}

	return desc
}

func syncServiceHealth(services *[]services.Service, duration string) {
	for s, service := range *services {
		service.SetOk()

		for m, module := range service.Modules {
			history, err := GetModuleHealthHistoryWithParams(module.Name, duration, health.DescSort, true)
			if err != nil {
				continue
			}

			module.SetOk()
			if isLastCheckUnhealthy(history) {
				module.SetUnhealthyStatus()
				service.ConvergeUnhealthyStatus(module.Name)
			}

			service.Modules[m] = module
		}

		(*services)[s] = service
	}
}

func genHealthSummary(services []services.Service) Health {
	health := Health{Services: services}
	health.Overall = &Overall{Status: status.Health{Current: status.Ok}}
	syncUnhealthStatus(&health, services)
	syncRepairingStatus(&health, &services)
	return health
}

func syncUnhealthStatus(health *Health, services []services.Service) {
	for _, service := range services {
		if service.IsStatusOk() {
			continue
		}

		health.Status.Current = status.Ng
		if health.Status.Description == "" {
			health.Status.Description = fmt.Sprintf("%s %s", service.Name, service.Status.Description)
		} else {
			health.Status.Description += fmt.Sprintf(", %s %s", service.Name, service.Status.Description)
		}
	}
}

func syncRepairingStatus(health *Health, services *[]services.Service) {
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

func isLastCheckUnhealthy(history []health.Check) bool {
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
		return ""
	}

	return strings.ReplaceAll(val, `;`, "\n")
}

func parseTime(record *query.FluxRecord) string {
	date, err := ostime.Parse(events.TimeLayout, record.Time().Local().String())
	if err != nil {
		log.Debugf("healths: failed to parse date from record(%v)", err)
	}

	return time.RFC3339Z(date)
}

func parseNodes(record *query.FluxRecord) []string {
	node := record.ValueByKey("node")
	val, ok := node.(string)
	if !ok {
		return []string{}
	}

	return []string{val}
}

func parseNode(record *query.FluxRecord) string {
	node := record.ValueByKey("node")
	val, ok := node.(string)
	if !ok {
		return "unknown"
	}

	return val
}

func parseLog(record *query.FluxRecord) string {
	log := record.ValueByKey("log")
	val, ok := log.(string)
	if !ok {
		return ""
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
