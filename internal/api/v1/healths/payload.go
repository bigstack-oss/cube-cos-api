package healths

import (
	"fmt"
	"time"

	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/bigstack-oss/cube-cos-api/internal/status"
	duration "github.com/xhit/go-str2duration"
)

// M1 TODO: this will be removed once the real data is available in the COS side
func (h *helper) genFakeHealthSummary() interface{} {
	return cubecos.Health{
		Overall: &cubecos.Overall{
			Status: status.Details{
				Current:     "ng",
				Description: "ceph has 2 ceph_osd down",
			},
		},
		Services: []definition.Service{
			{
				Name:     "clusterLink",
				Category: "core",
				Status:   status.NewOk(),
				Modules: []definition.Module{
					{
						Name:   "link",
						Status: status.NewOk(),
					},
					{
						Name:   "clock",
						Status: status.NewOk(),
					},
					{
						Name:   "dns",
						Status: status.NewOk(),
					},
				},
			},
			{
				Name:     "clusterSys",
				Category: "core",
				Status:   status.NewOk(),
				Modules: []definition.Module{
					{
						Name:   "bootstrap",
						Status: status.NewOk(),
					},
					{
						Name:   "license",
						Status: status.NewOk(),
					},
				},
			},
			{
				Name:     "clusterSettings",
				Category: "core",
				Status:   status.NewOk(),
				Modules: []definition.Module{
					{
						Name:   "etcd",
						Status: status.NewOk(),
					},
					{
						Name:   "nodelist",
						Status: status.NewOk(),
					},
				},
			},
			{
				Name:     "haCluster",
				Category: "core",
				Status:   status.NewOk(),
				Modules: []definition.Module{
					{
						Name:   "hacluster",
						Status: status.NewOk(),
					},
				},
			},
			{
				Name:     "msgQueue",
				Category: "core",
				Status:   status.NewOk(),
				Modules: []definition.Module{
					{
						Name:   "rabbitmq",
						Status: status.NewOk(),
					},
				},
			},
			{
				Name:     "iaasDb",
				Category: "core",
				Status:   status.NewOk(),
				Modules: []definition.Module{
					{
						Name:   "mysql",
						Status: status.NewOk(),
					},
					{
						Name:   "mongodb",
						Status: status.NewOk(),
					},
				},
			},
			{
				Name:     "virtualIp",
				Category: "core",
				Status:   status.NewOk(),
				Modules: []definition.Module{
					{
						Name:   "vip",
						Status: status.NewOk(),
					},
					{
						Name:   "haproxy_ha",
						Status: status.NewOk(),
					},
				},
			},
			{
				Name:     "singleSignOn",
				Category: "core",
				Status:   status.NewOk(),
				Modules: []definition.Module{
					{
						Name:   "k3s",
						Status: status.NewOk(),
					},
					{
						Name:   "keycloak",
						Status: status.NewOk(),
					},
				},
			},
			{
				Name:     "apiService",
				Category: "core",
				Status:   status.NewOk(),
				Modules: []definition.Module{
					{
						Name:   "haproxy",
						Status: status.NewOk(),
					},
					{
						Name:   "httpd",
						Status: status.NewOk(),
					},
					{
						Name:   "skyline",
						Status: status.NewOk(),
					},
					{
						Name:   "lmi",
						Status: status.NewOk(),
					},
					{
						Name:   "memcache",
						Status: status.NewOk(),
					},
				},
			},
			{
				Name:     "storage",
				Category: "storage",
				Status: &status.Details{
					Current:     "ng",
					Description: "ceph has 2 ceph_osd down",
				},
				Modules: []definition.Module{
					{
						Name:   "ceph",
						Status: status.NewOk(),
					},
					{
						Name:   "ceph_mon",
						Status: status.NewOk(),
					},
					{
						Name:   "ceph_mgr",
						Status: status.NewOk(),
					},
					{
						Name:   "ceph_mds",
						Status: status.NewOk(),
					},
					{
						Name: "ceph_osd",
						Status: &status.Details{
							Current:     "ng",
							Description: "2 osd down",
						},
					},
					{
						Name:   "ceph_rgw",
						Status: status.NewOk(),
					},
					{
						Name:   "rbd_target",
						Status: status.NewOk(),
					},
				},
			},
			{
				Name:     "compute",
				Category: "cloud computing",
				Status:   status.NewOk(),
				Modules: []definition.Module{
					{
						Name:   "nova",
						Status: status.NewOk(),
					},
					{
						Name:   "cyborg",
						Status: status.NewOk(),
					},
				},
			},
			{
				Name:     "network",
				Category: "cloud computing",
				Status:   status.NewOk(),
				Modules: []definition.Module{
					{
						Name:   "neutron",
						Status: status.NewOk(),
					},
				},
			},
			{
				Name:     "lbaas",
				Category: "cloud computing",
				Status:   status.NewOk(),
				Modules: []definition.Module{
					{
						Name:   "octavia",
						Status: status.NewOk(),
					},
				},
			},
			{
				Name:     "blockStorage",
				Category: "cloud computing",
				Status:   status.NewOk(),
				Modules: []definition.Module{
					{
						Name:   "cinder",
						Status: status.NewOk(),
					},
				},
			},
			{
				Name:     "fileStorage",
				Category: "cloud computing",
				Status:   status.NewOk(),
				Modules: []definition.Module{
					{
						Name:   "manila",
						Status: status.NewOk(),
					},
				},
			},
			{
				Name:     "objectStorage",
				Category: "cloud computing",
				Status:   status.NewOk(),
				Modules: []definition.Module{
					{
						Name:   "swift",
						Status: status.NewOk(),
					},
				},
			},
			{
				Name:     "bareMetal",
				Category: "cloud computing",
				Status:   status.NewOk(),
				Modules: []definition.Module{
					{
						Name:   "ironic",
						Status: status.NewOk(),
					},
				},
			},
			{
				Name:     "dnsaas",
				Category: "cloud computing",
				Status:   status.NewOk(),
				Modules: []definition.Module{
					{
						Name:   "designate",
						Status: status.NewOk(),
					},
				},
			},
			{
				Name:     "k8saas",
				Category: "cloud computing",
				Status:   status.NewOk(),
				Modules: []definition.Module{
					{
						Name:   "rancher",
						Status: status.NewOk(),
					},
				},
			},
			{
				Name:     "orchestration",
				Category: "cloud computing",
				Status:   status.NewOk(),
				Modules: []definition.Module{
					{
						Name:   "heat",
						Status: status.NewOk(),
					},
				},
			},
			{
				Name:     "instanceHa",
				Category: "cloud computing",
				Status:   status.NewOk(),
				Modules: []definition.Module{
					{
						Name:   "masakari",
						Status: status.NewOk(),
					},
				},
			},
			{
				Name:     "image",
				Category: "cloud computing",
				Status:   status.NewOk(),
				Modules: []definition.Module{
					{
						Name:   "glance",
						Status: status.NewOk(),
					},
				},
			},
			{
				Name:     "businessLogic",
				Category: "cloud computing",
				Status:   status.NewOk(),
				Modules: []definition.Module{
					{
						Name:   "senlin",
						Status: status.NewOk(),
					},
					{
						Name:   "watcher",
						Status: status.NewOk(),
					},
				},
			},
			{
				Name:     "dataPipe",
				Category: "infrascope",
				Status:   status.NewOk(),
				Modules: []definition.Module{
					{
						Name:   "kafka",
						Status: status.NewOk(),
					},
					{
						Name:   "zookeeper",
						Status: status.NewOk(),
					},
				},
			},
			{
				Name:     "notification",
				Category: "infrascope",
				Status:   status.NewOk(),
				Modules: []definition.Module{
					{
						Name:   "influxdb",
						Status: status.NewOk(),
					},
					{
						Name:   "kapacitor",
						Status: status.NewOk(),
					},
				},
			},
			{
				Name:     "metrics",
				Category: "infrascope",
				Status:   status.NewOk(),
				Modules: []definition.Module{
					{
						Name:   "monasca",
						Status: status.NewOk(),
					},
					{
						Name:   "telegraf",
						Status: status.NewOk(),
					},
					{
						Name:   "grafana",
						Status: status.NewOk(),
					},
				},
			},
			{
				Name:     "logAnalytics",
				Category: "infrascope",
				Status:   status.NewOk(),
				Modules: []definition.Module{
					{
						Name:   "filebeat",
						Status: status.NewOk(),
					},
					{
						Name:   "logstash",
						Status: status.NewOk(),
					},
					{
						Name:   "opensearch-dashboards",
						Status: status.NewOk(),
					},
					{
						Name:   "opensearch",
						Status: status.NewOk(),
					},
					{
						Name:   "auditbeat",
						Status: status.NewOk(),
					},
				},
			},
		},
	}
}

func (h *helper) genFakeHealthHistoryOfService() []cubecos.HealthStatus {
	modules := cubecos.ServiceToModules[h.service]
	statuses := []cubecos.HealthStatus{}

	pastTime, err := duration.Str2Duration(h.past)
	if err != nil {
		pastTime = 1 * time.Hour
	}
	h.period.stop = definition.TimeLocalISO8601(time.Now())
	h.period.start = definition.TimeLocalISO8601(time.Now().Add(-pastTime))

	for _, module := range modules {
		interval := 5 * time.Minute
		history := []cubecos.HealthCheck{}
		count := 0

		for start := h.StartTime(); !start.After(h.StopTime()); start = start.Add(interval) {
			timestamp := h.StartTime().Add(time.Duration(count) * interval).Format(time.RFC3339)
			status := "ok"
			checkResult := cubecos.HealthCheck{Time: timestamp, Status: status}
			if count%5 == 0 {
				h.setFakeError(&checkResult)
			}

			history = append(history, checkResult)
			count++
		}

		statuses = append(
			statuses,
			cubecos.HealthStatus{
				Category: cubecos.ServiceToCategory[h.service],
				Name:     h.service,
				Module:   module.Name,
				History:  history,
			},
		)
	}

	return statuses
}

// M1 TODO: this will be removed once the real data is available in the COS side
func (h *helper) genFakeHealthHistoryOfModule() cubecos.HealthStatus {
	interval := 5 * time.Minute
	history := []cubecos.HealthCheck{}
	count := 0

	pastTime, err := duration.Str2Duration(h.past)
	if err != nil {
		pastTime = 1 * time.Hour
	}
	h.period.stop = definition.TimeLocalISO8601(time.Now())
	h.period.start = definition.TimeLocalISO8601(time.Now().Add(-pastTime))

	for start := h.StartTime(); !start.After(h.StopTime()); start = start.Add(interval) {
		timestamp := h.StartTime().Add(time.Duration(count) * interval).Format(time.RFC3339)
		status := "ok"
		checkResult := cubecos.HealthCheck{Time: timestamp, Status: status}
		if count%5 == 0 {
			h.setFakeError(&checkResult)
		}

		history = append(history, checkResult)
		count++
	}

	return cubecos.HealthStatus{
		Category: cubecos.ServiceToCategory[h.service],
		Name:     h.service,
		Module:   h.module,
		History:  history,
	}
}

func (h *helper) setFakeError(checkResult *cubecos.HealthCheck) {
	checkResult.Status = "ng"
	checkResult.Error = &cubecos.Error{
		Type:        "service down",
		Nodes:       []string{definition.DataCenterName},
		Reason:      "1 node down",
		Description: "nova has 1 node down due to the memory exhausted, and the abnormal memory competition from PID(24887) is detected",
		Details:     "{ ... the best efforts of error summary / direction ...} ",
		Log: fmt.Sprintf(
			"http://{dataCenter}:8888/log/nova/%s-20250205113459-b3gc.log",
			definition.DataCenterName,
		),
	}
}

func genCheckRepairReq() *cubecos.Health {
	h := &cubecos.Health{}
	h.Overall = &cubecos.Overall{}
	h.Overall.Status.SetDesiredToCheckingAndRepairing()
	return h
}

func genForceRepairReq(module definition.Module) *cubecos.Health {
	h := &cubecos.Health{}
	h.Overall = &cubecos.Overall{}
	h.Overall.Status.SetDesiredToRepairing()
	svc := cubecos.ModuleToService[module.Name]
	h.Services = []definition.Service{
		{
			Name:     svc,
			Category: cubecos.ServiceToCategory[svc],
			Modules:  []definition.Module{module},
		},
	}
	return h
}
