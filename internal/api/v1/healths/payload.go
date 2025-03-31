package healths

import (
	"fmt"
	"time"

	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/bigstack-oss/cube-cos-api/internal/status"
	duration "github.com/xhit/go-str2duration"
)

// M1 TODO: this will be removed once the real data is available in the COS side
func (h *helper) genFakeHealthSummary() any {
	return cubecos.Health{
		Overall: &cubecos.Overall{
			Status: status.Health{
				Current:     "ng",
				Description: "ceph has 2 ceph_osd down",
			},
		},
		Services: []v1.Service{
			{
				Name:     "clusterLink",
				Category: "core",
				Status:   status.NewOk(),
				Modules: []v1.Module{
					{
						Name:         "link",
						Status:       status.NewOk(),
						IsRepairable: cubecos.IsRepairableModule("link"),
					},
					{
						Name:         "clock",
						Status:       status.NewOk(),
						IsRepairable: cubecos.IsRepairableModule("clock"),
					},
					{
						Name:         "dns",
						Status:       status.NewOk(),
						IsRepairable: cubecos.IsRepairableModule("dns"),
					},
				},
			},
			{
				Name:     "clusterSys",
				Category: "core",
				Status:   status.NewOk(),
				Modules: []v1.Module{
					{
						Name:         "bootstrap",
						Status:       status.NewOk(),
						IsRepairable: cubecos.IsRepairableModule("bootstrap"),
					},
					{
						Name:         "license",
						Status:       status.NewOk(),
						IsRepairable: cubecos.IsRepairableModule("license"),
					},
				},
			},
			{
				Name:     "clusterSettings",
				Category: "core",
				Status:   status.NewOk(),
				Modules: []v1.Module{
					{
						Name:         "etcd",
						Status:       status.NewOk(),
						IsRepairable: cubecos.IsRepairableModule("etcd"),
					},
					{
						Name:         "nodelist",
						Status:       status.NewOk(),
						IsRepairable: cubecos.IsRepairableModule("nodelist"),
					},
				},
			},
			{
				Name:     "haCluster",
				Category: "core",
				Status:   status.NewOk(),
				Modules: []v1.Module{
					{
						Name:         "hacluster",
						Status:       status.NewOk(),
						IsRepairable: cubecos.IsRepairableModule("hacluster"),
					},
				},
			},
			{
				Name:     "msgQueue",
				Category: "core",
				Status:   status.NewOk(),
				Modules: []v1.Module{
					{
						Name:         "rabbitmq",
						Status:       status.NewOk(),
						IsRepairable: cubecos.IsRepairableModule("rabbitmq"),
					},
				},
			},
			{
				Name:     "iaasDb",
				Category: "core",
				Status:   status.NewOk(),
				Modules: []v1.Module{
					{
						Name:         "mysql",
						Status:       status.NewOk(),
						IsRepairable: cubecos.IsRepairableModule("mysql"),
					},
					{
						Name:         "mongodb",
						Status:       status.NewOk(),
						IsRepairable: cubecos.IsRepairableModule("mongodb"),
					},
				},
			},
			{
				Name:     "virtualIp",
				Category: "core",
				Status:   status.NewOk(),
				Modules: []v1.Module{
					{
						Name:         "vip",
						Status:       status.NewOk(),
						IsRepairable: cubecos.IsRepairableModule("vip"),
					},
					{
						Name:         "haproxy_ha",
						Status:       status.NewOk(),
						IsRepairable: cubecos.IsRepairableModule("haproxy_ha"),
					},
				},
			},
			{
				Name:     "singleSignOn",
				Category: "core",
				Status:   status.NewOk(),
				Modules: []v1.Module{
					{
						Name:         "k3s",
						Status:       status.NewOk(),
						IsRepairable: cubecos.IsRepairableModule("k3s"),
					},
					{
						Name:         "keycloak",
						Status:       status.NewOk(),
						IsRepairable: cubecos.IsRepairableModule("keycloak"),
					},
				},
			},
			{
				Name:     "apiService",
				Category: "core",
				Status:   status.NewOk(),
				Modules: []v1.Module{
					{
						Name:         "haproxy",
						Status:       status.NewOk(),
						IsRepairable: cubecos.IsRepairableModule("haproxy"),
					},
					{
						Name:         "httpd",
						Status:       status.NewOk(),
						IsRepairable: cubecos.IsRepairableModule("httpd"),
					},
					{
						Name:         "skyline",
						Status:       status.NewOk(),
						IsRepairable: cubecos.IsRepairableModule("skyline"),
					},
					{
						Name:         "lmi",
						Status:       status.NewOk(),
						IsRepairable: cubecos.IsRepairableModule("lmi"),
					},
					{
						Name:         "memcache",
						Status:       status.NewOk(),
						IsRepairable: cubecos.IsRepairableModule("memcache"),
					},
					{
						Name:         "api",
						Status:       status.NewOk(),
						IsRepairable: cubecos.IsRepairableModule("api"),
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
				Modules: []v1.Module{
					{
						Name:         "ceph",
						Status:       status.NewOk(),
						IsRepairable: cubecos.IsRepairableModule("ceph"),
					},
					{
						Name:         "ceph_mon",
						Status:       status.NewOk(),
						IsRepairable: cubecos.IsRepairableModule("ceph_mon"),
					},
					{
						Name:         "ceph_mgr",
						Status:       status.NewOk(),
						IsRepairable: cubecos.IsRepairableModule("ceph_mgr"),
					},
					{
						Name:         "ceph_mds",
						Status:       status.NewOk(),
						IsRepairable: cubecos.IsRepairableModule("ceph_mds"),
					},
					{
						Name: "ceph_osd",
						Status: &status.Details{
							Current:     "ng",
							Description: "2 osd down",
						},
						IsRepairable: cubecos.IsRepairableModule("ceph_osd"),
					},
					{
						Name:         "ceph_rgw",
						Status:       status.NewOk(),
						IsRepairable: cubecos.IsRepairableModule("ceph_rgw"),
					},
					{
						Name:         "rbd_target",
						Status:       status.NewOk(),
						IsRepairable: cubecos.IsRepairableModule("rbd_target"),
					},
				},
			},
			{
				Name:     "compute",
				Category: "cloud computing",
				Status:   status.NewOk(),
				Modules: []v1.Module{
					{
						Name:         "nova",
						Status:       status.NewOk(),
						IsRepairable: cubecos.IsRepairableModule("nova"),
					},
					{
						Name:         "cyborg",
						Status:       status.NewOk(),
						IsRepairable: cubecos.IsRepairableModule("cyborg"),
					},
				},
			},
			{
				Name:     "network",
				Category: "cloud computing",
				Status:   status.NewOk(),
				Modules: []v1.Module{
					{
						Name:         "neutron",
						Status:       status.NewOk(),
						IsRepairable: cubecos.IsRepairableModule("neutron"),
					},
				},
			},
			{
				Name:     "lbaas",
				Category: "cloud computing",
				Status:   status.NewOk(),
				Modules: []v1.Module{
					{
						Name:         "octavia",
						Status:       status.NewOk(),
						IsRepairable: cubecos.IsRepairableModule("octavia"),
					},
				},
			},
			{
				Name:     "blockStorage",
				Category: "cloud computing",
				Status:   status.NewOk(),
				Modules: []v1.Module{
					{
						Name:         "cinder",
						Status:       status.NewOk(),
						IsRepairable: cubecos.IsRepairableModule("cinder"),
					},
				},
			},
			{
				Name:     "fileStorage",
				Category: "cloud computing",
				Status:   status.NewOk(),
				Modules: []v1.Module{
					{
						Name:         "manila",
						Status:       status.NewOk(),
						IsRepairable: cubecos.IsRepairableModule("manila"),
					},
				},
			},
			{
				Name:     "objectStorage",
				Category: "cloud computing",
				Status:   status.NewOk(),
				Modules: []v1.Module{
					{
						Name:         "swift",
						Status:       status.NewOk(),
						IsRepairable: cubecos.IsRepairableModule("swift"),
					},
				},
			},
			{
				Name:     "bareMetal",
				Category: "cloud computing",
				Status:   status.NewOk(),
				Modules: []v1.Module{
					{
						Name:         "ironic",
						Status:       status.NewOk(),
						IsRepairable: cubecos.IsRepairableModule("ironic"),
					},
				},
			},
			{
				Name:     "dnsaas",
				Category: "cloud computing",
				Status:   status.NewOk(),
				Modules: []v1.Module{
					{
						Name:         "designate",
						Status:       status.NewOk(),
						IsRepairable: cubecos.IsRepairableModule("designate"),
					},
				},
			},
			{
				Name:     "k8saas",
				Category: "cloud computing",
				Status:   status.NewOk(),
				Modules: []v1.Module{
					{
						Name:         "rancher",
						Status:       status.NewOk(),
						IsRepairable: cubecos.IsRepairableModule("rancher"),
					},
				},
			},
			{
				Name:     "orchestration",
				Category: "cloud computing",
				Status:   status.NewOk(),
				Modules: []v1.Module{
					{
						Name:         "heat",
						Status:       status.NewOk(),
						IsRepairable: cubecos.IsRepairableModule("heat"),
					},
				},
			},
			{
				Name:     "instanceHa",
				Category: "cloud computing",
				Status:   status.NewOk(),
				Modules: []v1.Module{
					{
						Name:         "masakari",
						Status:       status.NewOk(),
						IsRepairable: cubecos.IsRepairableModule("masakari"),
					},
				},
			},
			{
				Name:     "image",
				Category: "cloud computing",
				Status:   status.NewOk(),
				Modules: []v1.Module{
					{
						Name:         "glance",
						Status:       status.NewOk(),
						IsRepairable: cubecos.IsRepairableModule("glance"),
					},
				},
			},
			{
				Name:     "businessLogic",
				Category: "cloud computing",
				Status:   status.NewOk(),
				Modules: []v1.Module{
					{
						Name:         "senlin",
						Status:       status.NewOk(),
						IsRepairable: cubecos.IsRepairableModule("senlin"),
					},
					{
						Name:         "watcher",
						Status:       status.NewOk(),
						IsRepairable: cubecos.IsRepairableModule("watcher"),
					},
				},
			},
			{
				Name:     "dataPipe",
				Category: "infrascope",
				Status:   status.NewOk(),
				Modules: []v1.Module{
					{
						Name:         "kafka",
						Status:       status.NewOk(),
						IsRepairable: cubecos.IsRepairableModule("kafka"),
					},
					{
						Name:         "zookeeper",
						Status:       status.NewOk(),
						IsRepairable: cubecos.IsRepairableModule("zookeeper"),
					},
				},
			},
			{
				Name:     "notification",
				Category: "infrascope",
				Status:   status.NewOk(),
				Modules: []v1.Module{
					{
						Name:         "influxdb",
						Status:       status.NewOk(),
						IsRepairable: cubecos.IsRepairableModule("influxdb"),
					},
					{
						Name:         "kapacitor",
						Status:       status.NewOk(),
						IsRepairable: cubecos.IsRepairableModule("kapacitor"),
					},
				},
			},
			{
				Name:     "metrics",
				Category: "infrascope",
				Status:   status.NewOk(),
				Modules: []v1.Module{
					{
						Name:         "monasca",
						Status:       status.NewOk(),
						IsRepairable: cubecos.IsRepairableModule("monasca"),
					},
					{
						Name:         "telegraf",
						Status:       status.NewOk(),
						IsRepairable: cubecos.IsRepairableModule("telegraf"),
					},
					{
						Name:         "grafana",
						Status:       status.NewOk(),
						IsRepairable: cubecos.IsRepairableModule("grafana"),
					},
				},
			},
			{
				Name:     "logAnalytics",
				Category: "infrascope",
				Status:   status.NewOk(),
				Modules: []v1.Module{
					{
						Name:         "filebeat",
						Status:       status.NewOk(),
						IsRepairable: cubecos.IsRepairableModule("filebeat"),
					},
					{
						Name:         "logstash",
						Status:       status.NewOk(),
						IsRepairable: cubecos.IsRepairableModule("logstash"),
					},
					{
						Name:         "opensearch-dashboards",
						Status:       status.NewOk(),
						IsRepairable: cubecos.IsRepairableModule("opensearch-dashboards"),
					},
					{
						Name:         "opensearch",
						Status:       status.NewOk(),
						IsRepairable: cubecos.IsRepairableModule("opensearch"),
					},
					{
						Name:         "auditbeat",
						Status:       status.NewOk(),
						IsRepairable: cubecos.IsRepairableModule("auditbeat"),
					},
				},
			},
		},
	}
}

func (h *helper) getHealthSummary() any {
	return cubecos.ListModuleHealth(h.past)
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
