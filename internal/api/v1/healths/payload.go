package healths

import (
	"fmt"
	"time"

	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	duration "github.com/xhit/go-str2duration"
)

func (h *helper) getHealthSummary() any {
	return cubecos.GetHealthSummary(h.past)
}

func (h *helper) genFakeHealthHistoryOfService() []cubecos.HealthStatus {
	modules := cubecos.ServiceToModules[h.service]
	statuses := []cubecos.HealthStatus{}

	pastTime := 1 * time.Hour
	if h.isPastRequired() {
		pastTime, _ = duration.Str2Duration(h.past)
	}
	h.period.stop = v1.TimeRFC3339Z(time.Now())
	h.period.start = v1.TimeRFC3339Z(time.Now().Add(-pastTime))

	for _, module := range modules {
		interval := 5 * time.Minute
		history := []v1.HealthCheck{}
		count := 0

		for start := h.StartTime(); !start.After(h.StopTime()); start = start.Add(interval) {
			timestamp := h.StartTime().Add(time.Duration(count) * interval).Format(v1.RFC3339)
			status := "ok"
			checkResult := v1.HealthCheck{Time: timestamp, Status: status}
			if count%5 == 0 {
				h.setFakeError(&checkResult)
			}

			history = append(history, checkResult)
			count++
		}

		statuses = append(
			statuses,
			cubecos.HealthStatus{
				Category:     cubecos.ServiceToCategory[h.service],
				Name:         h.service,
				Module:       module.Name,
				IsRepairable: cubecos.IsRepairableModule(module.Name),
				History:      history,
			},
		)
	}

	return statuses
}

func (h *helper) genServiceHealthHistory() []cubecos.HealthStatus {
	return cubecos.GetServiceHealthHistory(h.service, h.past)
}

// M1 TODO: this will be removed once the real data is available in the COS side
func (h *helper) genFakeHealthHistoryOfModule() cubecos.HealthStatus {
	interval := 5 * time.Minute
	history := []v1.HealthCheck{}
	count := 0

	pastTime := 1 * time.Hour
	if h.isPastRequired() {
		pastTime, _ = duration.Str2Duration(h.past)
	}
	h.period.stop = v1.TimeRFC3339Z(time.Now())
	h.period.start = v1.TimeRFC3339Z(time.Now().Add(-pastTime))

	for start := h.StartTime(); !start.After(h.StopTime()); start = start.Add(interval) {
		timestamp := h.StartTime().Add(time.Duration(count) * interval).Format(v1.RFC3339)
		status := "ok"
		checkResult := v1.HealthCheck{Time: timestamp, Status: status}
		if count%5 == 0 {
			h.setFakeError(&checkResult)
		}

		history = append(history, checkResult)
		count++
	}

	return cubecos.HealthStatus{
		Category:     cubecos.ServiceToCategory[h.service],
		Name:         h.service,
		Module:       h.module,
		IsRepairable: cubecos.IsRepairableModule(h.module),
		History:      history,
	}
}

func (h *helper) setFakeError(checkResult *v1.HealthCheck) {
	checkResult.Status = "ng"
	checkResult.Error = &v1.Error{
		Type:        "service down",
		Nodes:       []string{v1.DataCenterName},
		Reason:      "1 node down",
		Description: "nova has 1 node down due to the memory exhausted, and the abnormal memory competition from PID(24887) is detected",
		Details:     "{ ... the best efforts of error summary / direction ...} ",
		Log: fmt.Sprintf(
			"http://{dataCenter}:8888/log/nova/%s-20250205113459-b3gc.log",
			v1.DataCenterName,
		),
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
