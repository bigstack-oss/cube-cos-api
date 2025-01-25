package cubecos

import definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"

var (
	OrderSensitiveServices = []definition.Service{
		{
			Name: "clusterLink",
			Modules: []definition.Module{
				{Name: "link", IsAutoRepairable: false},
				{Name: "clock", IsAutoRepairable: true},
				{Name: "dns", IsAutoRepairable: false},
			},
		},
		{
			Name: "clusterSys",
			Modules: []definition.Module{
				{Name: "bootstrap", IsAutoRepairable: false},
				{Name: "license", IsAutoRepairable: false},
			},
		},
		{
			Name: "clusterSettings",
			Modules: []definition.Module{
				{Name: "etcd", IsAutoRepairable: true},
				{Name: "nodelist", IsAutoRepairable: false},
			},
		},
		{
			Name: "haCluster",
			Modules: []definition.Module{
				{Name: "hacluster", IsAutoRepairable: true},
			},
		},
		{
			Name: "msgQueue",
			Modules: []definition.Module{
				{Name: "rabbitmq", IsAutoRepairable: true},
			},
		},
		{
			Name: "iaasDb",
			Modules: []definition.Module{
				{Name: "mysql", IsAutoRepairable: true},
			},
		},
		{
			Name: "virtualIp",
			Modules: []definition.Module{
				{Name: "vip", IsAutoRepairable: true},
				{Name: "haproxy_ha", IsAutoRepairable: true},
			},
		},
		{
			Name: "storage",
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
			Name: "apiService",
			Modules: []definition.Module{
				{Name: "haproxy", IsAutoRepairable: true},
				{Name: "httpd", IsAutoRepairable: true},
				{Name: "skyline", IsAutoRepairable: true},
				{Name: "lmi", IsAutoRepairable: true},
				{Name: "memcache", IsAutoRepairable: true},
			},
		},
		{
			Name: "singleSignOn",
			Modules: []definition.Module{
				{Name: "k3s", IsAutoRepairable: true},
				{Name: "keycloak", IsAutoRepairable: true},
			},
		},
		{
			Name: "network",
			Modules: []definition.Module{
				{Name: "neutron", IsAutoRepairable: true},
			},
		},
		{
			Name: "compute",
			Modules: []definition.Module{
				{Name: "nova", IsAutoRepairable: true},
				{Name: "cyborg", IsAutoRepairable: true},
			},
		},
		{
			Name: "bareMetal",
			Modules: []definition.Module{
				{Name: "ironic", IsAutoRepairable: true},
			},
		},
		{
			Name: "image",
			Modules: []definition.Module{
				{Name: "glance", IsAutoRepairable: true},
			},
		},
		{
			Name: "blockStor",
			Modules: []definition.Module{
				{Name: "cinder", IsAutoRepairable: true},
			},
		},
		{
			Name: "fileStor",
			Modules: []definition.Module{
				{Name: "manila", IsAutoRepairable: true},
			},
		},
		{
			Name: "objectStor",
			Modules: []definition.Module{
				{Name: "swift", IsAutoRepairable: false},
			},
		},
		{
			Name: "orchestration",
			Modules: []definition.Module{
				{Name: "heat", IsAutoRepairable: true},
			},
		},
		{
			Name: "lbaas",
			Modules: []definition.Module{
				{Name: "octavia", IsAutoRepairable: true},
			},
		},
		{
			Name: "dnsaas",
			Modules: []definition.Module{
				{Name: "designate", IsAutoRepairable: true},
			},
		},
		{
			Name: "k8saas",
			Modules: []definition.Module{
				{Name: "rancher", IsAutoRepairable: false},
			},
		},
		{
			Name: "instanceHa",
			Modules: []definition.Module{
				{Name: "masakari", IsAutoRepairable: true},
			},
		},
		{
			Name: "businessLogic",
			Modules: []definition.Module{
				{Name: "senlin", IsAutoRepairable: true},
				{Name: "watcher", IsAutoRepairable: true},
			},
		},
		{
			Name: "dataPipe",
			Modules: []definition.Module{
				{Name: "zookeeper", IsAutoRepairable: true},
				{Name: "kafka", IsAutoRepairable: true},
			},
		},
		{
			Name: "metrics",
			Modules: []definition.Module{
				{Name: "monasca", IsAutoRepairable: true},
				{Name: "telegraf", IsAutoRepairable: true},
				{Name: "grafana", IsAutoRepairable: true},
			},
		},
		{
			Name: "logAnalytics",
			Modules: []definition.Module{
				{Name: "filebeat", IsAutoRepairable: true},
				{Name: "auditbeat", IsAutoRepairable: true},
				{Name: "logstash", IsAutoRepairable: true},
				{Name: "opensearch", IsAutoRepairable: true},
				{Name: "opensearch-dashboards", IsAutoRepairable: true},
			},
		},
		{
			Name: "notifications",
			Modules: []definition.Module{
				{Name: "influxdb", IsAutoRepairable: true},
				{Name: "kapacitor", IsAutoRepairable: true},
			},
		},
		{
			Name: "node",
			Modules: []definition.Module{
				{Name: "node", IsAutoRepairable: false},
			},
		},
	}
)
