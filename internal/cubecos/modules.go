package cubecos

import definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"

var (
	OrderSensitiveServices = []definition.Service{
		{
			Name:     "clusterLink",
			Category: "core",
			Modules: []definition.Module{
				{Name: "link", IsRepairable: false},
				{Name: "clock", IsRepairable: true},
				{Name: "dns", IsRepairable: false},
			},
		},
		{
			Name:     "clusterSys",
			Category: "core",
			Modules: []definition.Module{
				{Name: "bootstrap", IsRepairable: false},
				{Name: "license", IsRepairable: false},
			},
		},
		{
			Name:     "clusterSettings",
			Category: "core",
			Modules: []definition.Module{
				{Name: "etcd", IsRepairable: true},
				{Name: "nodelist", IsRepairable: false},
			},
		},
		{
			Name:     "haCluster",
			Category: "core",
			Modules: []definition.Module{
				{Name: "hacluster", IsRepairable: true},
			},
		},
		{
			Name:     "msgQueue",
			Category: "core",
			Modules: []definition.Module{
				{Name: "rabbitmq", IsRepairable: true},
			},
		},
		{
			Name:     "iaasDb",
			Category: "core",
			Modules: []definition.Module{
				{Name: "mysql", IsRepairable: true},
				{Name: "mongodb", IsRepairable: true},
			},
		},
		{
			Name:     "virtualIp",
			Category: "core",
			Modules: []definition.Module{
				{Name: "vip", IsRepairable: true},
				{Name: "haproxy_ha", IsRepairable: true},
			},
		},
		{
			Name:     "storage",
			Category: "storage",
			Modules: []definition.Module{
				{Name: "ceph", IsRepairable: false},
				{Name: "ceph_mon", IsRepairable: true},
				{Name: "ceph_mgr", IsRepairable: true},
				{Name: "ceph_mds", IsRepairable: true},
				{Name: "ceph_osd", IsRepairable: true},
				{Name: "ceph_rgw", IsRepairable: true},
				{Name: "rbd_target", IsRepairable: false},
			},
		},
		{
			Name:     "apiService",
			Category: "core",
			Modules: []definition.Module{
				{Name: "haproxy", IsRepairable: true},
				{Name: "httpd", IsRepairable: true},
				{Name: "skyline", IsRepairable: true},
				{Name: "lmi", IsRepairable: true},
				{Name: "memcache", IsRepairable: true},
				{Name: "api", IsRepairable: true},
			},
		},
		{
			Name:     "singleSignOn",
			Category: "core",
			Modules: []definition.Module{
				{Name: "k3s", IsRepairable: true},
				{Name: "keycloak", IsRepairable: true},
			},
		},
		{
			Name:     "network",
			Category: "cloud computing",
			Modules: []definition.Module{
				{Name: "neutron", IsRepairable: true},
			},
		},
		{
			Name:     "compute",
			Category: "cloud computing",
			Modules: []definition.Module{
				{Name: "nova", IsRepairable: true},
				{Name: "cyborg", IsRepairable: true},
			},
		},
		{
			Name:     "bareMetal",
			Category: "cloud computing",
			Modules: []definition.Module{
				{Name: "ironic", IsRepairable: true},
			},
		},
		{
			Name:     "image",
			Category: "cloud computing",
			Modules: []definition.Module{
				{Name: "glance", IsRepairable: true},
			},
		},
		{
			Name:     "blockStor",
			Category: "cloud computing",
			Modules: []definition.Module{
				{Name: "cinder", IsRepairable: true},
			},
		},
		{
			Name:     "fileStor",
			Category: "cloud computing",
			Modules: []definition.Module{
				{Name: "manila", IsRepairable: true},
			},
		},
		{
			Name:     "objectStor",
			Category: "cloud computing",
			Modules: []definition.Module{
				{Name: "swift", IsRepairable: false},
			},
		},
		{
			Name:     "orchestration",
			Category: "cloud computing",
			Modules: []definition.Module{
				{Name: "heat", IsRepairable: true},
			},
		},
		{
			Name:     "lbaas",
			Category: "cloud computing",
			Modules: []definition.Module{
				{Name: "octavia", IsRepairable: true},
			},
		},
		{
			Name:     "dnsaas",
			Category: "cloud computing",
			Modules: []definition.Module{
				{Name: "designate", IsRepairable: true},
			},
		},
		{
			Name:     "k8saas",
			Category: "cloud computing",
			Modules: []definition.Module{
				{Name: "rancher", IsRepairable: false},
			},
		},
		{
			Name:     "instanceHa",
			Category: "cloud computing",
			Modules: []definition.Module{
				{Name: "masakari", IsRepairable: true},
			},
		},
		{
			Name:     "businessLogic",
			Category: "cloud computing",
			Modules: []definition.Module{
				{Name: "senlin", IsRepairable: true},
				{Name: "watcher", IsRepairable: true},
			},
		},
		{
			Name:     "dataPipe",
			Category: "infrascope",
			Modules: []definition.Module{
				{Name: "zookeeper", IsRepairable: true},
				{Name: "kafka", IsRepairable: true},
			},
		},
		{
			Name:     "metrics",
			Category: "infrascope",
			Modules: []definition.Module{
				{Name: "monasca", IsRepairable: true},
				{Name: "telegraf", IsRepairable: true},
				{Name: "grafana", IsRepairable: true},
			},
		},
		{
			Name:     "logAnalytics",
			Category: "infrascope",
			Modules: []definition.Module{
				{Name: "filebeat", IsRepairable: true},
				{Name: "auditbeat", IsRepairable: true},
				{Name: "logstash", IsRepairable: true},
				{Name: "opensearch", IsRepairable: true},
				{Name: "opensearch-dashboards", IsRepairable: true},
			},
		},
		{
			Name:     "notifications",
			Category: "infrascope",
			Modules: []definition.Module{
				{Name: "influxdb", IsRepairable: true},
				{Name: "kapacitor", IsRepairable: true},
			},
		},
		{
			Name:               "node",
			IsInternalViewOnly: true,
			Modules: []definition.Module{
				{Name: "node", IsRepairable: false},
			},
		},
	}

	Modules           = map[string]definition.Module{}
	ModuleToService   = map[string]string{}
	ServiceToCategory = map[string]string{}
	ServiceToModules  = map[string][]definition.Module{}
)

func init() {
	initModuleMap()
	initModuleToServiceMap()
	initServiceToCategoryMap()
	initServiceToModulesMap()
}

func initModuleMap() {
	for _, service := range OrderSensitiveServices {
		for _, module := range service.Modules {
			Modules[module.Name] = module
		}
	}
}

func initModuleToServiceMap() {
	for _, service := range OrderSensitiveServices {
		for _, module := range service.Modules {
			ModuleToService[module.Name] = service.Name
		}
	}
}

func initServiceToCategoryMap() {
	for _, service := range OrderSensitiveServices {
		ServiceToCategory[service.Name] = service.Category
	}
}

func initServiceToModulesMap() {
	for _, service := range OrderSensitiveServices {
		ServiceToModules[service.Name] = service.Modules
	}
}

func IsValidService(service string) bool {
	_, ok := ServiceToModules[service]
	return ok
}

func IsValidServiceAndModule(service, module string) bool {
	modules, ok := ServiceToModules[service]
	if !ok {
		return false
	}

	for _, m := range modules {
		if m.Name == module {
			return true
		}
	}

	return false
}

func IsRepairableModule(module string) bool {
	m, ok := Modules[module]
	if !ok {
		return false
	}

	return m.IsRepairable
}
