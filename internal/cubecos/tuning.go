package cubecos

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	cuberr "github.com/bigstack-oss/cube-cos-api/internal/error"
	"github.com/google/uuid"
	log "go-micro.dev/v5/logger"
	"gopkg.in/yaml.v2"
)

const (
	policyFile = "/etc/policies/tuning/tuning1_0.yml"

	// private tunings
	CubeSysHa            = "cubesys.ha"
	CubeSysController    = "cubesys.controller"
	CubeSysControllerVip = "cubesys.control.vip"

	// public tunings
	BarbicanDebugEnabled        = "barbican.debug.enabled"
	CephDebugEnabled            = "ceph.debug.enabled"
	CephMirrorMetaSync          = "ceph.mirror.meta.sync"
	CinderBackupAccount         = "cinder.backup.account"
	CinderBackupEndpoint        = "cinder.backup.endpoint"
	CinderBackupOverride        = "cinder.backup.override"
	CinderBackupPool            = "cinder.backup.pool"
	CinderBackupSecret          = "cinder.backup.secret"
	CinderBackupType            = "cinder.backup.type"
	CinderDebugEnabled          = "cinder.debug.enabled"
	CinderExternalAccount       = "cinder.external.%d.account"
	CinderExternalDriver        = "cinder.external.%d.driver"
	CinderExternalEndpoint      = "cinder.external.%d.endpoint"
	CinderExternalName          = "cinder.external.%d.name"
	CinderExternalPool          = "cinder.external.%d.pool"
	CinderExternalSecret        = "cinder.external.%d.secret"
	CubesysAlertLevel           = "cubesys.alert.level"
	CubesysAlertLevelS          = "cubesys.alert.level.%s"
	CubesysConntableMax         = "cubesys.conntable.max"
	CubesysLogDefaultRetention  = "cubesys.log.default.retention"
	CubesysProviderExtra        = "cubesys.provider.extra"
	CyborgDebugEnabled          = "cyborg.debug.enabled"
	DebugEnableCoreDumpS        = "debug.enable_core_dump.%s"
	DebugEnableKdump            = "debug.enable_kdump"
	DebugLevelS                 = "debug.level.%s"
	DebugMaxCoreDump            = "debug.max_core_dump"
	DesignateDebugEnabled       = "designate.debug.enabled"
	GlanceDebugEnabled          = "glance.debug.enabled"
	GlanceExportRp              = "glance.export.rp"
	HeatDebugEnabled            = "heat.debug.enabled"
	InfluxdbCuratorRp           = "influxdb.curator.rp"
	IronicDebugEnabled          = "ironic.debug.enabled"
	IronicDeployServer          = "ironic.deploy.server"
	KapacitorAlertCheckEnabled  = "kapacitor.alert.check.enabled"
	KapacitorAlertCheckEventId  = "kapacitor.alert.check.eventid"
	KapacitorAlertCheckInterval = "kapacitor.alert.check.interval"
	KapacitorAlertExtraPrefix   = "kapacitor.alert.extra.prefix"
	KapacitorAlertFlowBase      = "kapacitor.alert.flow.base"
	KapacitorAlertFlowThreshold = "kapacitor.alert.flow.threshold"
	KapacitorAlertFlowUnit      = "kapacitor.alert.flow.unit"
	KeystoneDebugEnabled        = "keystone.debug.enabled"
	ManilaDebugEnabled          = "manila.debug.enabled"
	ManilaVolumeType            = "manila.volume.type"
	MasakariHostEvacuateAll     = "masakari.host.evacuate_all"
	MasakariWaitPeriod          = "masakari.wait.period"
	MonascaDebugEnabled         = "monasca.debug.enabled"
	MysqlBackupCuratorRp        = "mysql.backup.curator.rp"
	NetIfMtuName                = "net.if.mtu.<name>"
	NetIpv4TcpSyncookies        = "net.ipv4.tcp_syncookies"
	NetLacpDefaultRate          = "net.lacp.default.rate"
	NetLacpDefaultXmit          = "net.lacp.default.xmit"
	NeutronDebugEnabled         = "neutron.debug.enabled"
	NovaControlHostMemory       = "nova.control.host.memory"
	NovaControlHostVcpu         = "nova.control.host.vcpu"
	NovaDebugEnabled            = "nova.debug.enabled"
	NovaGpuType                 = "nova.gpu.type"
	NovaOvercommitCpuRatio      = "nova.overcommit.cpu.ratio"
	NovaOvercommitDiskRatio     = "nova.overcommit.disk.ratio"
	NovaOvercommitRamRatio      = "nova.overcommit.ram.ratio"
	NtpDebugEnabled             = "ntp.debug.enabled"
	OctaviaDebugEnabled         = "octavia.debug.enabled"
	OctaviaHa                   = "octavia.ha"
	OpensearchCuratorRp         = "opensearch.curator.rp"
	OpensearchHeapSize          = "opensearch.heap.size"
	SenlinDebugEnabled          = "senlin.debug.enabled"
	SkylineDebugEnabled         = "skyline.debug.enabled"
	SnapshotApplyAction         = "snapshot.apply.action"
	SnapshotApplyPolicyIgnore   = "snapshot.apply.policy.ignore"
	SshdBindToAllInterfaces     = "sshd.bind_to_all_interfaces"
	SshdSessionInactivity       = "sshd.session.inactivity"
	TimeTimezone                = "time.timezone"
	UpdateSecurityAutoUpdate    = "update.security.autoupdate"
	WatcherDebugEnabled         = "watcher.debug.enabled"
)

var (
	BarbicanDebugEnabledSpec = &definition.TuningSpec{
		Name:        BarbicanDebugEnabled,
		Description: "Set to true to enable barbican verbose log.",
		ExampleValue: definition.ExampleValue{
			Type:    "bool",
			Default: false,
		},
		Roles: definition.AllGeneralRoles,
	}
	CephDebugEnabledSpec = &definition.TuningSpec{
		Name:        CephDebugEnabled,
		Description: "Set to true to enable ceph debug logs.",
		ExampleValue: definition.ExampleValue{
			Type:    "bool",
			Default: false,
		},
		Roles: definition.AllGeneralRoles,
	}
	CephMirrorMetaSyncSpec = &definition.TuningSpec{
		Name:        CephMirrorMetaSync,
		Description: "Set to true to enable automatically volume metadata sync.",
		ExampleValue: definition.ExampleValue{
			Type:    "bool",
			Default: true,
		},
		Roles: definition.ControlRoles,
	}
	CinderBackupAccountSpec = &definition.TuningSpec{
		Name:        CinderBackupAccount,
		Description: "Set cinder backup storage account.",
		ExampleValue: definition.ExampleValue{
			Type:    "string",
			Default: "",
		},
		Roles: definition.AllGeneralRoles,
	}
	CinderBackupEndpointSpec = &definition.TuningSpec{
		Name:        CinderBackupEndpoint,
		Description: "Set cinder backup storage endpoint.",
		ExampleValue: definition.ExampleValue{
			Type:    "string",
			Default: "",
		},
		Roles: definition.AllRoles,
	}
	CinderBackupOverrideSpec = &definition.TuningSpec{
		Name:        CinderBackupOverride,
		Description: "Enable override cinder backup configurations.",
		ExampleValue: definition.ExampleValue{
			Type:    "bool",
			Default: false,
		},
		Roles: definition.AllRoles,
	}
	CinderBackupPoolSpec = &definition.TuningSpec{
		Name:        CinderBackupPool,
		Description: "Set cinder backup storage pool.",
		ExampleValue: definition.ExampleValue{
			Type:    "string",
			Default: "",
		},
		Roles: definition.AllRoles,
	}
	CinderBackupSecretSpec = &definition.TuningSpec{
		Name:        CinderBackupSecret,
		Description: "Set cinder backup storage account secret.",
		ExampleValue: definition.ExampleValue{
			Type:    "string",
			Default: "",
		},
		Roles: definition.AllRoles,
	}
	CinderBackupTypeSpec = &definition.TuningSpec{
		Name:        CinderBackupType,
		Description: "Set cinder backup storage type <cube-storage|cube-swift>.",
		ExampleValue: definition.ExampleValue{
			Type:    "string",
			Default: "",
		},
		Roles: definition.AllRoles,
	}
	CinderDebugEnabledSpec = &definition.TuningSpec{
		Name:        CinderDebugEnabled,
		Description: "Set to true to enable cinder verbose log.",
		ExampleValue: definition.ExampleValue{
			Type:    "bool",
			Default: false,
		},
		Roles: definition.AllRoles,
	}
	CinderExternalAccountSpec = &definition.TuningSpec{
		Name:        CinderExternalAccount,
		Description: "Set cinder external storage account.",
		ExampleValue: definition.ExampleValue{
			Type:    "string",
			Default: "",
		},
		Roles: definition.AllRoles,
	}
	CinderExternalDriverSpec = &definition.TuningSpec{
		Name:        CinderExternalDriver,
		Description: "Set cinder external storage type name <cube|purestorage>.",
		ExampleValue: definition.ExampleValue{
			Type:    "string",
			Default: "",
		},
		Roles: definition.AllRoles,
	}
	CinderExternalEndpointSpec = &definition.TuningSpec{
		Name:        CinderExternalEndpoint,
		Description: "Set cinder external storage endpoint.",
		ExampleValue: definition.ExampleValue{
			Type:    "string",
			Default: "",
		},
		Roles: definition.AllRoles,
	}
	CinderExternalNameSpec = &definition.TuningSpec{
		Name:        CinderExternalName,
		Description: "Set cinder external storage rule name.",
		ExampleValue: definition.ExampleValue{
			Type:    "string",
			Default: "",
		},
		Roles: definition.AllRoles,
	}
	CinderExternalPoolSpec = &definition.TuningSpec{
		Name:        CinderExternalPool,
		Description: "Set cinder external storage pool.",
		ExampleValue: definition.ExampleValue{
			Type:    "string",
			Default: "",
		},
		Roles: definition.AllRoles,
	}
	CinderExternalSecretSpec = &definition.TuningSpec{
		Name:        CinderExternalSecret,
		Description: "Set cinder external storage account secret.",
		ExampleValue: definition.ExampleValue{
			Type:    "string",
			Default: "",
		},
		Roles: definition.AllRoles,
	}
	CubesysAlertLevelSpec = &definition.TuningSpec{
		Name:        CubesysAlertLevel,
		Description: "Set health alert sensible level. (0: default, 1: highly sensitive)",
		ExampleValue: definition.ExampleValue{
			Type:    "int",
			Default: 0,
			Min:     0,
			Max:     2147483647,
		},
		Roles: definition.AllRoles,
	}
	CubesysAlertLevelSSpec = &definition.TuningSpec{
		Name:        CubesysAlertLevelS,
		Description: "Set health alert sensible level for service %s. (0: default, 1: highly sensitive)",
		ExampleValue: definition.ExampleValue{
			Type:    "int",
			Default: 0,
			Min:     0,
			Max:     2147483647,
		},
		Roles: definition.AllRoles,
	}
	CubesysConntableMaxSpec = &definition.TuningSpec{
		Name:        CubesysConntableMax,
		Description: "Set max connection table size.",
		ExampleValue: definition.ExampleValue{
			Type:    "int",
			Default: 262144,
			Min:     0,
			Max:     2147483647,
		},
		Roles: definition.AllRoles,
	}
	CubesysLogDefaultRetentionSpec = &definition.TuningSpec{
		Name:        CubesysLogDefaultRetention,
		Description: "Set log file retention policy in days.",
		ExampleValue: definition.ExampleValue{
			Type:    "int",
			Default: 14,
			Min:     0,
			Max:     365,
		},
		Roles: definition.AllRoles,
	}
	CubesysProviderExtraSpec = &definition.TuningSpec{
		Name:        CubesysProviderExtra,
		Description: "Set extra provider interfaces ('pvd-' prefix and <= 15 chars) [IF.2:pvd-xxx,eth2:pvd-yyy,...].",
		ExampleValue: definition.ExampleValue{
			Type:    "string",
			Default: "",
		},
		Roles: definition.AllRoles,
	}
	CyborgDebugEnabledSpec = &definition.TuningSpec{
		Name:        CyborgDebugEnabled,
		Description: "Set to true to enable cyborg verbose log.",
		ExampleValue: definition.ExampleValue{
			Type:    "bool",
			Default: false,
		},
		Roles: definition.AllRoles,
	}
	DebugEnableCoreDumpSSpec = &definition.TuningSpec{
		Name:        DebugEnableCoreDumpS,
		Description: "Enable core dump for process %s",
		ExampleValue: definition.ExampleValue{
			Type:    "bool",
			Default: false,
		},
		Roles: definition.AllRoles,
	}
	DebugEnableKdumpSpec = &definition.TuningSpec{
		Name:        DebugEnableKdump,
		Description: "Enable kdump to collect dump from kernel panic",
		ExampleValue: definition.ExampleValue{
			Type:    "bool",
			Default: false,
		},
		Roles: definition.AllRoles,
	}
	DebugLevelSSpec = &definition.TuningSpec{
		Name:        DebugLevelS,
		Description: "Set debug level for process %s",
		ExampleValue: definition.ExampleValue{
			Type:    "int",
			Default: 0,
			Min:     0,
			Max:     9,
		},
		Roles: definition.AllRoles,
	}
	DebugMaxCoreDumpSpec = &definition.TuningSpec{
		Name:        DebugMaxCoreDump,
		Description: "Set the total number of core files before oldest are removed",
		ExampleValue: definition.ExampleValue{
			Type:    "int",
			Default: 0,
			Min:     0,
			Max:     999,
		},
		Roles: definition.AllRoles,
	}
	DesignateDebugEnabledSpec = &definition.TuningSpec{
		Name:        DesignateDebugEnabled,
		Description: "Set to true to enable designate verbose log.",
		ExampleValue: definition.ExampleValue{
			Type:    "bool",
			Default: false,
		},
		Roles: definition.ControlRoles,
	}
	GlanceDebugEnabledSpec = &definition.TuningSpec{
		Name:        GlanceDebugEnabled,
		Description: "Set to true to enable glance verbose log.",
		ExampleValue: definition.ExampleValue{
			Type:    "bool",
			Default: false,
		},
		Roles: definition.AllRoles,
	}
	GlanceExportRpSpec = &definition.TuningSpec{
		Name:        GlanceExportRp,
		Description: "glance export retention policy in copies.",
		ExampleValue: definition.ExampleValue{
			Type:    "int",
			Default: 3,
			Min:     0,
			Max:     255,
		},
		Roles: definition.AllRoles,
	}
	HeatDebugEnabledSpec = &definition.TuningSpec{
		Name:        HeatDebugEnabled,
		Description: "Set to true to enable heat verbose log.",
		ExampleValue: definition.ExampleValue{
			Type:    "bool",
			Default: false,
		},
		Roles: definition.AllRoles,
	}
	InfluxdbCuratorRpSpec = &definition.TuningSpec{
		Name:        InfluxdbCuratorRp,
		Description: "influxdb curator retention policy in days.",
		ExampleValue: definition.ExampleValue{
			Type:    "int",
			Default: 7,
			Min:     0,
			Max:     365,
		},
		Roles: definition.AllRoles,
	}
	IronicDebugEnabledSpec = &definition.TuningSpec{
		Name:        IronicDebugEnabled,
		Description: "Set to true to enable ironic verbose log.",
		ExampleValue: definition.ExampleValue{
			Type:    "bool",
			Default: false,
		},
		Roles: definition.AllGeneralRoles,
	}
	IronicDeployServerSpec = &definition.TuningSpec{
		Name:        IronicDeployServer,
		Description: "Set to true to enable ironic deploy server (dhcp/tftp/pxe/http).",
		ExampleValue: definition.ExampleValue{
			Type:    "bool",
			Default: false,
		},
		Roles: definition.AllGeneralRoles,
	}
	KapacitorAlertCheckEnabledSpec = &definition.TuningSpec{
		Name:        KapacitorAlertCheckEnabled,
		Description: "Set true to enable kapacitor alert check.",
		ExampleValue: definition.ExampleValue{
			Type:    "bool",
			Default: false,
		},
		Roles: definition.ControlRoles,
	}
	KapacitorAlertCheckEventIdSpec = &definition.TuningSpec{
		Name:        KapacitorAlertCheckEventId,
		Description: "Set kapacitor alert check eventid.",
		ExampleValue: definition.ExampleValue{
			Type:    "string",
			Default: "SYS00002W",
		},
		Roles: definition.ControlRoles,
	}
	KapacitorAlertCheckIntervalSpec = &definition.TuningSpec{
		Name:        KapacitorAlertCheckInterval,
		Description: "Set kapacitor alert check interval (default to 60m).",
		ExampleValue: definition.ExampleValue{
			Type:    "string",
			Default: "60m",
		},
		Roles: definition.ControlRoles,
	}
	KapacitorAlertExtraPrefixSpec = &definition.TuningSpec{
		Name:        KapacitorAlertExtraPrefix,
		Description: "Set kapacitor alert message prefix.",
		ExampleValue: definition.ExampleValue{
			Type:    "string",
			Default: "Cube",
		},
		Roles: definition.ControlRoles,
	}
	KapacitorAlertFlowBaseSpec = &definition.TuningSpec{
		Name:        KapacitorAlertFlowBase,
		Description: "Set kapacitor alert base for abnormal flow.",
		ExampleValue: definition.ExampleValue{
			Type:    "string",
			Default: "7d",
		},
		Roles: definition.ControlRoles,
	}
	KapacitorAlertFlowThresholdSpec = &definition.TuningSpec{
		Name:        KapacitorAlertFlowThreshold,
		Description: "Set kapacitor alert threshold for abnormal flow.",
		ExampleValue: definition.ExampleValue{
			Type:    "int",
			Default: 30,
			Min:     0,
			Max:     65535,
		},
		Roles: definition.ControlRoles,
	}
	KapacitorAlertFlowUnitSpec = &definition.TuningSpec{
		Name:        KapacitorAlertFlowUnit,
		Description: "Set kapacitor alert unit for abnormal flow.",
		ExampleValue: definition.ExampleValue{
			Type:    "string",
			Default: "5m",
		},
		Roles: definition.ControlRoles,
	}
	KeystoneDebugEnabledSpec = &definition.TuningSpec{
		Name:        KeystoneDebugEnabled,
		Description: "Set to true to enable keystone verbose log.",
		ExampleValue: definition.ExampleValue{
			Type:    "bool",
			Default: false,
		},
		Roles: definition.AllRoles,
	}
	ManilaDebugEnabledSpec = &definition.TuningSpec{
		Name:        ManilaDebugEnabled,
		Description: "Set to true to enable manila verbose log.",
		ExampleValue: definition.ExampleValue{
			Type:    "bool",
			Default: false,
		},
		Roles: definition.AllRoles,
	}
	ManilaVolumeTypeSpec = &definition.TuningSpec{
		Name:        ManilaVolumeType,
		Description: "Set manila backend volume type.",
		ExampleValue: definition.ExampleValue{
			Type:    "string",
			Default: "CubeStorage",
		},
		Roles: definition.AllRoles,
	}
	MasakariHostEvacuateAllSpec = &definition.TuningSpec{
		Name:        MasakariHostEvacuateAll,
		Description: "Set to true to enable evacuate all instances when host goes down.",
		ExampleValue: definition.ExampleValue{
			Type:    "bool",
			Default: true,
		},
		Roles: definition.AllRoles,
	}
	MasakariWaitPeriodSpec = &definition.TuningSpec{
		Name:        MasakariWaitPeriod,
		Description: "Set wait period after service update",
		ExampleValue: definition.ExampleValue{
			Type:    "int",
			Default: 0,
			Min:     0,
			Max:     99999,
		},
		Roles: definition.ControlRoles,
	}
	MonascaDebugEnabledSpec = &definition.TuningSpec{
		Name:        MonascaDebugEnabled,
		Description: "Set to true to enable monasca verbose log.",
		ExampleValue: definition.ExampleValue{
			Type:    "bool",
			Default: false,
		},
		Roles: definition.AllRoles,
	}
	MysqlBackupCuratorRpSpec = &definition.TuningSpec{
		Name:        MysqlBackupCuratorRp,
		Description: "mysql backup retention policy in weeks.",
		ExampleValue: definition.ExampleValue{
			Type:    "int",
			Default: 14,
			Min:     0,
			Max:     52,
		},
		Roles: definition.AllRoles,
	}
	NetIfMtuNameSpec = &definition.TuningSpec{
		Name:        NetIfMtuName,
		Description: "Set interface MTU (MTU of parent interface must be greater than its VLAN interface).",
		ExampleValue: definition.ExampleValue{
			Type:    "int",
			Default: 1500,
			Min:     68,
			Max:     65536,
		},
		Roles: definition.AllRoles,
	}
	NetIpv4TcpSyncookiesSpec = &definition.TuningSpec{
		Name:        NetIpv4TcpSyncookies,
		Description: "Turn on the Linux SYN cookies implementation.",
		ExampleValue: definition.ExampleValue{
			Type:    "bool",
			Default: true,
		},
		Roles: definition.AllRoles,
	}
	NetLacpDefaultRateSpec = &definition.TuningSpec{
		Name:        NetLacpDefaultRate,
		Description: "Set default LACP rate (fast/slow).",
		ExampleValue: definition.ExampleValue{
			Type:    "string",
			Default: "fast",
		},
		Roles: definition.AllRoles,
	}
	NetLacpDefaultXmitSpec = &definition.TuningSpec{
		Name:        NetLacpDefaultXmit,
		Description: "Set default LACP transmit hash policy (layer2/layer2+3/layer3+4).",
		ExampleValue: definition.ExampleValue{
			Type:    "string",
			Default: "layer3+4",
		},
		Roles: definition.AllRoles,
	}
	NeutronDebugEnabledSpec = &definition.TuningSpec{
		Name:        NeutronDebugEnabled,
		Description: "Set to true to enable neutron verbose log.",
		ExampleValue: definition.ExampleValue{
			Type:    "bool",
			Default: false,
		},
		Roles: definition.AllRoles,
	}
	NovaControlHostMemorySpec = &definition.TuningSpec{
		Name:        NovaControlHostMemory,
		Description: "Amount of memory in MB to reserve for the control host.",
		ExampleValue: definition.ExampleValue{
			Type:    "int",
			Default: 0,
			Min:     0,
			Max:     524288,
		},
		Roles: definition.ComputeRoles,
	}
	NovaControlHostVcpuSpec = &definition.TuningSpec{
		Name:        NovaControlHostVcpu,
		Description: "Amount of vcpu to reserve for the control host.",
		ExampleValue: definition.ExampleValue{
			Type:    "int",
			Default: 0,
			Min:     0,
			Max:     128,
		},
		Roles: []*definition.Role{definition.GetControlConvergeRole(), definition.GetEdgeCoreRole()},
	}
	NovaDebugEnabledSpec = &definition.TuningSpec{
		Name:        NovaDebugEnabled,
		Description: "Set to true to enable nova verbose log.",
		ExampleValue: definition.ExampleValue{
			Type:    "bool",
			Default: false,
		},
		Roles: definition.AllRoles,
	}
	NovaGpuTypeSpec = &definition.TuningSpec{
		Name:        NovaGpuType,
		Description: "Specify a supported gpu type instances would get.",
		ExampleValue: definition.ExampleValue{
			Type:    "string",
			Default: "",
		},
		Roles: definition.ComputeRoles,
		Selector: definition.Selector{
			Enabled: true,
			Labels:  map[string]string{"isGpuEnabled": "true"},
		},
	}
	NovaOvercommitCpuRatioSpec = &definition.TuningSpec{
		Name:        NovaOvercommitCpuRatio,
		Description: "Specify an allowed CPU overcommitted ratio.",
		ExampleValue: definition.ExampleValue{
			Type:    "float",
			Default: 16.0,
		},
		Roles: definition.AllRoles,
	}
	NovaOvercommitDiskRatioSpec = &definition.TuningSpec{
		Name:        NovaOvercommitDiskRatio,
		Description: "Specify an allowed disk overcommitted ratio.",
		ExampleValue: definition.ExampleValue{
			Type:    "float",
			Default: 1.0,
		},
		Roles: definition.AllRoles,
	}
	NovaOvercommitRamRatioSpec = &definition.TuningSpec{
		Name:        NovaOvercommitRamRatio,
		Description: "Specify an allowed RAM overcommitted ratio.",
		ExampleValue: definition.ExampleValue{
			Type:    "float",
			Default: 1.5,
		},
		Roles: definition.AllRoles,
	}
	NtpDebugEnabledSpec = &definition.TuningSpec{
		Name:        NtpDebugEnabled,
		Description: "Set to true to enable ntp verbose log.",
		ExampleValue: definition.ExampleValue{
			Type:    "bool",
			Default: false,
		},
		Roles: definition.AllRoles,
	}
	OctaviaDebugEnabledSpec = &definition.TuningSpec{
		Name:        OctaviaDebugEnabled,
		Description: "Set to true to enable octavia verbose log.",
		ExampleValue: definition.ExampleValue{
			Type:    "bool",
			Default: false,
		},
		Roles: definition.AllRoles,
	}
	OctaviaHaSpec = &definition.TuningSpec{
		Name:        OctaviaHa,
		Description: "Set to true to enable octavia HA mode.",
		ExampleValue: definition.ExampleValue{
			Type:    "bool",
			Default: false,
		},
		Roles: definition.AllRoles,
	}
	OpensearchCuratorRpSpec = &definition.TuningSpec{
		Name:        OpensearchCuratorRp,
		Description: "opensearch curator retention policy in days.",
		ExampleValue: definition.ExampleValue{
			Type:    "int",
			Default: 7,
			Min:     0,
			Max:     365,
		},
		Roles: definition.AllRoles,
	}
	OpensearchHeapSizeSpec = &definition.TuningSpec{
		Name:        OpensearchHeapSize,
		Description: "Set opensearch heap size in MB.",
		ExampleValue: definition.ExampleValue{
			Type:    "int",
			Default: 1024,
			Min:     256,
			Max:     65536,
		},
		Roles: definition.AllRoles,
	}
	SenlinDebugEnabledSpec = &definition.TuningSpec{
		Name:        SenlinDebugEnabled,
		Description: "Set to true to enable senlin verbose log.",
		ExampleValue: definition.ExampleValue{
			Type:    "bool",
			Default: false,
		},
		Roles: definition.AllRoles,
	}
	SkylineDebugEnabledSpec = &definition.TuningSpec{
		Name:        SkylineDebugEnabled,
		Description: "Set to true to enable skyline verbose log.",
		ExampleValue: definition.ExampleValue{
			Type:    "bool",
			Default: false,
		},
		Roles: definition.AllRoles,
	}
	SnapshotApplyActionSpec = &definition.TuningSpec{
		Name:        SnapshotApplyAction,
		Description: "Set snapshot apply action <apply|revert>.",
		ExampleValue: definition.ExampleValue{
			Type:    "string",
			Default: "apply",
		},
		Roles: definition.AllRoles,
	}
	SnapshotApplyPolicyIgnoreSpec = &definition.TuningSpec{
		Name:        SnapshotApplyPolicyIgnore,
		Description: "Set snapshot apply policy ignore <true|false>.",
		ExampleValue: definition.ExampleValue{
			Type:    "bool",
			Default: false,
		},
		Roles: definition.AllRoles,
	}
	SshdBindToAllInterfacesSpec = &definition.TuningSpec{
		Name:        SshdBindToAllInterfaces,
		Description: "Set to true to bind sshd to all interfaces.",
		ExampleValue: definition.ExampleValue{
			Type:    "bool",
			Default: false,
		},
		Roles: definition.AllRoles,
	}
	SshdSessionInactivitySpec = &definition.TuningSpec{
		Name:        SshdSessionInactivity,
		Description: "Set sshd session inactivity timeout in seconds.",
		ExampleValue: definition.ExampleValue{
			Type:    "int",
			Default: 0,
			Min:     0,
			Max:     86400,
		},
		Roles: definition.AllRoles,
	}
	TimeTimezoneSpec = &definition.TuningSpec{
		Name:        TimeTimezone,
		Description: "Set system timezone.",
		ExampleValue: definition.ExampleValue{
			Type:    "string",
			Default: "UTC",
		},
		Roles: definition.AllRoles,
	}
	UpdateSecurityAutoUpdateSpec = &definition.TuningSpec{
		Name:        UpdateSecurityAutoUpdate,
		Description: "Set to true to enable security autoupdate.",
		ExampleValue: definition.ExampleValue{
			Type:    "bool",
			Default: false,
		},
		Roles: definition.AllRoles,
	}
	WatcherDebugEnabledSpec = &definition.TuningSpec{
		Name:        WatcherDebugEnabled,
		Description: "Set to true to enable watcher verbose log.",
		ExampleValue: definition.ExampleValue{
			Type:    "bool",
			Default: false,
		},
		Roles: definition.AllRoles,
	}
)

func init() {
	definition.SetSpecToTuning(BarbicanDebugEnabled, BarbicanDebugEnabledSpec)
	definition.SetSpecToTuning(CephDebugEnabled, CephDebugEnabledSpec)
	definition.SetSpecToTuning(CephMirrorMetaSync, CephMirrorMetaSyncSpec)
	definition.SetSpecToTuning(CinderBackupAccount, CinderBackupAccountSpec)
	definition.SetSpecToTuning(CinderBackupEndpoint, CinderBackupEndpointSpec)
	definition.SetSpecToTuning(CinderBackupOverride, CinderBackupOverrideSpec)
	definition.SetSpecToTuning(CinderBackupPool, CinderBackupPoolSpec)
	definition.SetSpecToTuning(CinderBackupSecret, CinderBackupSecretSpec)
	definition.SetSpecToTuning(CinderBackupType, CinderBackupTypeSpec)
	definition.SetSpecToTuning(CinderDebugEnabled, CinderDebugEnabledSpec)
	definition.SetSpecToTuning(CinderExternalAccount, CinderExternalAccountSpec)
	definition.SetSpecToTuning(CinderExternalDriver, CinderExternalDriverSpec)
	definition.SetSpecToTuning(CinderExternalEndpoint, CinderExternalEndpointSpec)
	definition.SetSpecToTuning(CinderExternalName, CinderExternalNameSpec)
	definition.SetSpecToTuning(CinderExternalPool, CinderExternalPoolSpec)
	definition.SetSpecToTuning(CinderExternalSecret, CinderExternalSecretSpec)
	definition.SetSpecToTuning(CubesysAlertLevel, CubesysAlertLevelSpec)
	definition.SetSpecToTuning(CubesysAlertLevelS, CubesysAlertLevelSSpec)
	definition.SetSpecToTuning(CubesysConntableMax, CubesysConntableMaxSpec)
	definition.SetSpecToTuning(CubesysLogDefaultRetention, CubesysLogDefaultRetentionSpec)
	definition.SetSpecToTuning(CubesysProviderExtra, CubesysProviderExtraSpec) // no setup logic in cubecos
	definition.SetSpecToTuning(CyborgDebugEnabled, CyborgDebugEnabledSpec)
	definition.SetSpecToTuning(DebugEnableCoreDumpS, DebugEnableCoreDumpSSpec)
	definition.SetSpecToTuning(DebugEnableKdump, DebugEnableKdumpSpec)
	definition.SetSpecToTuning(DebugLevelS, DebugLevelSSpec)
	definition.SetSpecToTuning(DebugMaxCoreDump, DebugMaxCoreDumpSpec)
	definition.SetSpecToTuning(DesignateDebugEnabled, DesignateDebugEnabledSpec) // need to check edge part
	definition.SetSpecToTuning(GlanceDebugEnabled, GlanceDebugEnabledSpec)       // no setup logic in cubecos
	definition.SetSpecToTuning(GlanceExportRp, GlanceExportRpSpec)
	definition.SetSpecToTuning(HeatDebugEnabled, HeatDebugEnabledSpec)
	definition.SetSpecToTuning(InfluxdbCuratorRp, InfluxdbCuratorRpSpec)
	definition.SetSpecToTuning(IronicDebugEnabled, IronicDebugEnabledSpec)
	definition.SetSpecToTuning(IronicDeployServer, IronicDeployServerSpec)
	definition.SetSpecToTuning(KapacitorAlertCheckEnabled, KapacitorAlertCheckEnabledSpec)
	definition.SetSpecToTuning(KapacitorAlertCheckEventId, KapacitorAlertCheckEventIdSpec)
	definition.SetSpecToTuning(KapacitorAlertCheckInterval, KapacitorAlertCheckIntervalSpec)
	definition.SetSpecToTuning(KapacitorAlertExtraPrefix, KapacitorAlertExtraPrefixSpec)
	definition.SetSpecToTuning(KapacitorAlertFlowBase, KapacitorAlertFlowBaseSpec)
	definition.SetSpecToTuning(KapacitorAlertFlowThreshold, KapacitorAlertFlowThresholdSpec)
	definition.SetSpecToTuning(KapacitorAlertFlowUnit, KapacitorAlertFlowUnitSpec)
	definition.SetSpecToTuning(KeystoneDebugEnabled, KeystoneDebugEnabledSpec) // no setup logic in cubecos
	definition.SetSpecToTuning(ManilaDebugEnabled, ManilaDebugEnabledSpec)
	definition.SetSpecToTuning(ManilaVolumeType, ManilaVolumeTypeSpec)
	definition.SetSpecToTuning(MasakariHostEvacuateAll, MasakariHostEvacuateAllSpec)
	definition.SetSpecToTuning(MasakariWaitPeriod, MasakariWaitPeriodSpec) // need to check IsEdge(s_eCubeRole) means what
	definition.SetSpecToTuning(MonascaDebugEnabled, MonascaDebugEnabledSpec)
	definition.SetSpecToTuning(MysqlBackupCuratorRp, MysqlBackupCuratorRpSpec)
	definition.SetSpecToTuning(NetIfMtuName, NetIfMtuNameSpec)
	definition.SetSpecToTuning(NetIpv4TcpSyncookies, NetIpv4TcpSyncookiesSpec)
	definition.SetSpecToTuning(NetLacpDefaultRate, NetLacpDefaultRateSpec)
	definition.SetSpecToTuning(NetLacpDefaultXmit, NetLacpDefaultXmitSpec)
	definition.SetSpecToTuning(NeutronDebugEnabled, NeutronDebugEnabledSpec)
	definition.SetSpecToTuning(NovaControlHostMemory, NovaControlHostMemorySpec)
	definition.SetSpecToTuning(NovaControlHostVcpu, NovaControlHostVcpuSpec) // why no compute role
	definition.SetSpecToTuning(NovaDebugEnabled, NovaDebugEnabledSpec)
	definition.SetSpecToTuning(NovaGpuType, NovaGpuTypeSpec)
	definition.SetSpecToTuning(NovaOvercommitCpuRatio, NovaOvercommitCpuRatioSpec)
	definition.SetSpecToTuning(NovaOvercommitDiskRatio, NovaOvercommitDiskRatioSpec)
	definition.SetSpecToTuning(NovaOvercommitRamRatio, NovaOvercommitRamRatioSpec)
	definition.SetSpecToTuning(NtpDebugEnabled, NtpDebugEnabledSpec) // no setup logic in cubecos
	definition.SetSpecToTuning(OctaviaDebugEnabled, OctaviaDebugEnabledSpec)
	definition.SetSpecToTuning(OctaviaHa, OctaviaHaSpec)
	definition.SetSpecToTuning(OpensearchCuratorRp, OpensearchCuratorRpSpec)
	definition.SetSpecToTuning(OpensearchHeapSize, OpensearchHeapSizeSpec)
	definition.SetSpecToTuning(SenlinDebugEnabled, SenlinDebugEnabledSpec)   // EOL in OpenStack, consider to remove
	definition.SetSpecToTuning(SkylineDebugEnabled, SkylineDebugEnabledSpec) // why only check IsControl() but no IsConverged()
	definition.SetSpecToTuning(SnapshotApplyAction, SnapshotApplyActionSpec) // no setup logic in cubecos
	definition.SetSpecToTuning(SnapshotApplyPolicyIgnore, SnapshotApplyPolicyIgnoreSpec)
	definition.SetSpecToTuning(SshdBindToAllInterfaces, SshdBindToAllInterfacesSpec)
	definition.SetSpecToTuning(SshdSessionInactivity, SshdSessionInactivitySpec)
	definition.SetSpecToTuning(TimeTimezone, TimeTimezoneSpec)
	definition.SetSpecToTuning(UpdateSecurityAutoUpdate, UpdateSecurityAutoUpdateSpec)
	definition.SetSpecToTuning(WatcherDebugEnabled, WatcherDebugEnabledSpec)
}

func ReadHexTuning(parameterName string) (string, error) {
	b, err := exec.Command("hex_tuning_helper", "/etc/settings.txt", "", parameterName).Output()
	if err != nil {
		log.Errorf("failed to read hex tunning value: %s", err.Error())
		return "", err
	}

	keyValue := strings.Split(string(b), "'")
	if len(keyValue) < 2 {
		return "", cuberr.TuningParamNotFound
	}

	return keyValue[1], nil
}

func ApplyHexTuning(isolatedDir string) error {
	_, err := exec.Command("hex_config", "apply", isolatedDir).Output()
	if err != nil {
		log.Errorf("failed to apply hex tunning value: %s", err.Error())
		return err
	}

	return nil
}

func IsHexTuningApplied(tuning definition.Tuning) error {
	value, err := ReadHexTuning(tuning.Name)
	if err != nil {
		return err
	}

	if tuning.Value != value {
		return fmt.Errorf("tuning value is not applied: %s", tuning.Name)
	}

	return nil
}

func ApplyHexTunings(tunings []definition.Tuning) error {
	newTunings, err := genTuningsAsYaml(tunings)
	if err != nil {
		return err
	}

	tmpTuningDir := genTmpTuningDir()
	err = writeTuningToFile(tmpTuningDir, newTunings)
	if err != nil {
		return err
	}

	err = ApplyHexTuning(tmpTuningDir)
	if err != nil {
		return err
	}

	return nil
}

func genTuningsAsYaml(tunings []definition.Tuning) ([]byte, error) {
	tuningTemplate := definition.Policy{
		Name:    "tuning",
		Version: "1.0",
		Enabled: true,
		Tunings: tunings,
	}

	yml, err := yaml.Marshal(&tuningTemplate)
	if err != nil {
		log.Errorf("failed to marshal batch tuning info: %s", err.Error())
		return nil, err
	}

	return yml, nil
}

func writeTuningToFile(tmpDir string, yml []byte) error {
	fullDir := fmt.Sprintf("%s/tuning", tmpDir)
	err := os.MkdirAll(fullDir, 0755)
	if err != nil {
		log.Errorf("failed to create isolated tuning directory: %s", err.Error())
		return err
	}

	file, err := os.Create(fmt.Sprintf("%s/tuning1_0.yml", fullDir))
	if err != nil {
		log.Errorf("failed to create isolated tuning file: %s", err.Error())
		return err
	}

	defer file.Close()

	_, err = io.Writer.Write(file, yml)
	if err != nil {
		log.Errorf("failed to write tuning info to isolated file: %s", err.Error())
		return err
	}

	return nil
}

func genTmpTuningDir() string {
	hash := uuid.New().String()[:8]
	return fmt.Sprintf("/tmp/tuning-%s", hash)
}

func AcquireTuningLock() error {
	return nil
}

func ReleaseTuningLock() error {
	return nil
}

func GetPolicy() (*definition.Policy, error) {
	b, err := os.ReadFile(policyFile)
	if err != nil {
		return nil, err
	}

	policy := &definition.Policy{}
	err = yaml.Unmarshal(b, policy)
	if err != nil {
		return nil, err
	}

	return policy, nil
}

func IsHexTuningDeleted(tuning definition.Tuning) error {
	_, err := ReadHexTuning(tuning.Name)
	if err == nil {
		return fmt.Errorf("tuning value is not deleted: %s", tuning.Name)
	}

	if !errors.Is(err, cuberr.TuningParamNotFound) {
		return fmt.Errorf("failed to check if tuning is deleted: %s", err.Error())
	}

	return nil
}
