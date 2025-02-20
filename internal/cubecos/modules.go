package cubecos

import definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"

var (
	OrderSensitiveServices = []definition.Service{
		{
			Name:     "clusterLink",
			Category: "core",
			Modules: []definition.Module{
				{Name: "link", IsAutoRepairable: false},
				{Name: "clock", IsAutoRepairable: true},
				{Name: "dns", IsAutoRepairable: false},
			},
		},
		{
			Name:     "clusterSys",
			Category: "core",
			Modules: []definition.Module{
				{Name: "bootstrap", IsAutoRepairable: false},
				{Name: "license", IsAutoRepairable: false},
			},
		},
		{
			Name:     "clusterSettings",
			Category: "core",
			Modules: []definition.Module{
				{Name: "etcd", IsAutoRepairable: true},
				{Name: "nodelist", IsAutoRepairable: false},
			},
		},
		{
			Name:     "haCluster",
			Category: "core",
			Modules: []definition.Module{
				{Name: "hacluster", IsAutoRepairable: true},
			},
		},
		{
			Name:     "msgQueue",
			Category: "core",
			Modules: []definition.Module{
				{Name: "rabbitmq", IsAutoRepairable: true},
			},
		},
		{
			Name:     "iaasDb",
			Category: "core",
			Modules: []definition.Module{
				{Name: "mysql", IsAutoRepairable: true},
				{Name: "mongodb", IsAutoRepairable: true},
			},
		},
		{
			Name:     "virtualIp",
			Category: "core",
			Modules: []definition.Module{
				{Name: "vip", IsAutoRepairable: true},
				{Name: "haproxy_ha", IsAutoRepairable: true},
			},
		},
		{
			Name:     "storage",
			Category: "storage",
			Modules: []definition.Module{
				{Name: "ceph", IsAutoRepairable: false},
				{Name: "ceph_mon", IsAutoRepairable: true},
				{Name: "ceph_mgr", IsAutoRepairable: true},
				{Name: "ceph_mds", IsAutoRepairable: true},
				{Name: "ceph_osd", IsAutoRepairable: true},
				{Name: "ceph_rgw", IsAutoRepairable: true},
				{Name: "rbd_target", IsAutoRepairable: false},
			},
		},
		{
			Name:     "apiService",
			Category: "core",
			Modules: []definition.Module{
				{Name: "haproxy", IsAutoRepairable: true},
				{Name: "httpd", IsAutoRepairable: true},
				{Name: "skyline", IsAutoRepairable: true},
				{Name: "lmi", IsAutoRepairable: true},
				{Name: "memcache", IsAutoRepairable: true},
			},
		},
		{
			Name:     "singleSignOn",
			Category: "core",
			Modules: []definition.Module{
				{Name: "k3s", IsAutoRepairable: true},
				{Name: "keycloak", IsAutoRepairable: true},
			},
		},
		{
			Name:     "network",
			Category: "cloud computing",
			Modules: []definition.Module{
				{Name: "neutron", IsAutoRepairable: true},
			},
		},
		{
			Name:     "compute",
			Category: "cloud computing",
			Modules: []definition.Module{
				{Name: "nova", IsAutoRepairable: true},
				{Name: "cyborg", IsAutoRepairable: true},
			},
		},
		{
			Name:     "bareMetal",
			Category: "cloud computing",
			Modules: []definition.Module{
				{Name: "ironic", IsAutoRepairable: true},
			},
		},
		{
			Name:     "image",
			Category: "cloud computing",
			Modules: []definition.Module{
				{Name: "glance", IsAutoRepairable: true},
			},
		},
		{
			Name:     "blockStor",
			Category: "cloud computing",
			Modules: []definition.Module{
				{Name: "cinder", IsAutoRepairable: true},
			},
		},
		{
			Name:     "fileStor",
			Category: "cloud computing",
			Modules: []definition.Module{
				{Name: "manila", IsAutoRepairable: true},
			},
		},
		{
			Name:     "objectStor",
			Category: "cloud computing",
			Modules: []definition.Module{
				{Name: "swift", IsAutoRepairable: false},
			},
		},
		{
			Name:     "orchestration",
			Category: "cloud computing",
			Modules: []definition.Module{
				{Name: "heat", IsAutoRepairable: true},
			},
		},
		{
			Name:     "lbaas",
			Category: "cloud computing",
			Modules: []definition.Module{
				{Name: "octavia", IsAutoRepairable: true},
			},
		},
		{
			Name:     "dnsaas",
			Category: "cloud computing",
			Modules: []definition.Module{
				{Name: "designate", IsAutoRepairable: true},
			},
		},
		{
			Name:     "k8saas",
			Category: "cloud computing",
			Modules: []definition.Module{
				{Name: "rancher", IsAutoRepairable: false},
			},
		},
		{
			Name:     "instanceHa",
			Category: "cloud computing",
			Modules: []definition.Module{
				{Name: "masakari", IsAutoRepairable: true},
			},
		},
		{
			Name:     "businessLogic",
			Category: "cloud computing",
			Modules: []definition.Module{
				{Name: "senlin", IsAutoRepairable: true},
				{Name: "watcher", IsAutoRepairable: true},
			},
		},
		{
			Name:     "dataPipe",
			Category: "infrascope",
			Modules: []definition.Module{
				{Name: "zookeeper", IsAutoRepairable: true},
				{Name: "kafka", IsAutoRepairable: true},
			},
		},
		{
			Name:     "metrics",
			Category: "infrascope",
			Modules: []definition.Module{
				{Name: "monasca", IsAutoRepairable: true},
				{Name: "telegraf", IsAutoRepairable: true},
				{Name: "grafana", IsAutoRepairable: true},
			},
		},
		{
			Name:     "logAnalytics",
			Category: "infrascope",
			Modules: []definition.Module{
				{Name: "filebeat", IsAutoRepairable: true},
				{Name: "auditbeat", IsAutoRepairable: true},
				{Name: "logstash", IsAutoRepairable: true},
				{Name: "opensearch", IsAutoRepairable: true},
				{Name: "opensearch-dashboards", IsAutoRepairable: true},
			},
		},
		{
			Name:     "notifications",
			Category: "infrascope",
			Modules: []definition.Module{
				{Name: "influxdb", IsAutoRepairable: true},
				{Name: "kapacitor", IsAutoRepairable: true},
			},
		},
		{
			Name:               "node",
			IsInternalViewOnly: true,
			Modules: []definition.Module{
				{Name: "node", IsAutoRepairable: false},
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
