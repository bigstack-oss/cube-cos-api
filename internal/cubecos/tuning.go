package cubecos

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"slices"
	"strconv"
	"strings"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/http"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/bodies"
	conf "github.com/bigstack-oss/cube-cos-api/internal/config"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/base"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/errors"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/tunings"
	"github.com/google/uuid"
	json "github.com/json-iterator/go"
	log "go-micro.dev/v5/logger"
	"gopkg.in/yaml.v3"
)

const (
	TuningPolicyFile = "/etc/policies/tuning/tuning1_0.yml"

	// private tunings
	CubeSysHa                = "cubesys.ha"
	CubeSysController        = "cubesys.controller"
	CubeSysControllerVip     = "cubesys.control.vip"
	CubeSysControllerIp      = "cubesys.controller.ip"
	CubeSysManagementNetwork = "cubesys.management"
	CubeSysStorageNetwork    = "cubesys.storage"
	CubeNetIfAddrPrefix      = "net.if.addr."
	NetIfAddrEth0            = "net.if.addr.eth0"

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

	// setting sys
	SysProductDescription = "sys.product.description"
	SysProductVersion     = "sys.product.version"
)

var (
	tuningToRoles     = map[string][]*nodes.Role{}
	tuningToSelectors = map[string]nodes.Selector{}
)

func init() {
	setTuningToRoles()
	setTuningToSelectors()
	setTuningSpecs()
}

func setTuningToRoles() {
	tuningToRoles[BarbicanDebugEnabled] = nodes.AllGeneralRoles
	tuningToRoles[CephDebugEnabled] = nodes.AllGeneralRoles
	tuningToRoles[CephMirrorMetaSync] = nodes.ControlRoles
	tuningToRoles[CinderBackupAccount] = nodes.AllGeneralRoles
	tuningToRoles[CinderBackupEndpoint] = nodes.AllRoles
	tuningToRoles[CinderBackupOverride] = nodes.AllRoles
	tuningToRoles[CinderBackupPool] = nodes.AllRoles
	tuningToRoles[CinderBackupSecret] = nodes.AllRoles
	tuningToRoles[CinderBackupType] = nodes.AllRoles
	tuningToRoles[CinderDebugEnabled] = nodes.AllRoles
	tuningToRoles[CinderExternalAccount] = nodes.AllRoles
	tuningToRoles[CinderExternalDriver] = nodes.AllRoles
	tuningToRoles[CinderExternalEndpoint] = nodes.AllRoles
	tuningToRoles[CinderExternalName] = nodes.AllRoles
	tuningToRoles[CinderExternalPool] = nodes.AllRoles
	tuningToRoles[CinderExternalSecret] = nodes.AllRoles
	tuningToRoles[CubesysAlertLevel] = nodes.AllRoles
	tuningToRoles[CubesysAlertLevelS] = nodes.AllRoles
	tuningToRoles[CubesysConntableMax] = nodes.AllRoles
	tuningToRoles[CubesysLogDefaultRetention] = nodes.AllRoles
	tuningToRoles[CubesysProviderExtra] = nodes.AllRoles
	tuningToRoles[CyborgDebugEnabled] = nodes.AllRoles
	tuningToRoles[DebugEnableCoreDumpS] = nodes.AllRoles
	tuningToRoles[DebugEnableKdump] = nodes.AllRoles
	tuningToRoles[DebugLevelS] = nodes.AllRoles
	tuningToRoles[DebugMaxCoreDump] = nodes.AllRoles
	tuningToRoles[DesignateDebugEnabled] = nodes.AllRoles
	tuningToRoles[GlanceDebugEnabled] = nodes.AllRoles
	tuningToRoles[GlanceExportRp] = nodes.AllRoles
	tuningToRoles[HeatDebugEnabled] = nodes.AllRoles
	tuningToRoles[InfluxdbCuratorRp] = nodes.AllRoles
	tuningToRoles[IronicDebugEnabled] = nodes.AllGeneralRoles
	tuningToRoles[IronicDeployServer] = nodes.AllGeneralRoles
	tuningToRoles[KapacitorAlertCheckEnabled] = nodes.ControlRoles
	tuningToRoles[KapacitorAlertCheckEventId] = nodes.ControlRoles
	tuningToRoles[KapacitorAlertCheckInterval] = nodes.ControlRoles
	tuningToRoles[KapacitorAlertExtraPrefix] = nodes.ControlRoles
	tuningToRoles[KapacitorAlertFlowBase] = nodes.ControlRoles
	tuningToRoles[KapacitorAlertFlowThreshold] = nodes.ControlRoles
	tuningToRoles[KapacitorAlertFlowUnit] = nodes.ControlRoles
	tuningToRoles[KeystoneDebugEnabled] = nodes.AllRoles
	tuningToRoles[ManilaDebugEnabled] = nodes.AllRoles
	tuningToRoles[ManilaVolumeType] = nodes.AllRoles
	tuningToRoles[MasakariHostEvacuateAll] = nodes.AllRoles
	tuningToRoles[MasakariWaitPeriod] = nodes.ControlRoles
	tuningToRoles[MonascaDebugEnabled] = nodes.AllRoles
	tuningToRoles[MysqlBackupCuratorRp] = nodes.AllRoles
	tuningToRoles[NetIfMtuName] = nodes.AllRoles
	tuningToRoles[NetIpv4TcpSyncookies] = nodes.AllRoles
	tuningToRoles[NetLacpDefaultRate] = nodes.AllRoles
	tuningToRoles[NetLacpDefaultXmit] = nodes.AllRoles
	tuningToRoles[NeutronDebugEnabled] = nodes.AllRoles
	tuningToRoles[NovaControlHostMemory] = nodes.ComputeRoles
	tuningToRoles[NovaControlHostVcpu] = []*nodes.Role{nodes.GetControlConvergeRole(), nodes.GetEdgeCoreRole()}
	tuningToRoles[NovaDebugEnabled] = nodes.AllRoles
	tuningToRoles[NovaGpuType] = nodes.ComputeRoles
	tuningToRoles[NovaOvercommitCpuRatio] = nodes.AllRoles
	tuningToRoles[NovaOvercommitDiskRatio] = nodes.AllRoles
	tuningToRoles[NovaOvercommitRamRatio] = nodes.AllRoles
	tuningToRoles[NtpDebugEnabled] = nodes.AllRoles
	tuningToRoles[OctaviaDebugEnabled] = nodes.AllRoles
	tuningToRoles[OctaviaHa] = nodes.AllRoles
	tuningToRoles[OpensearchCuratorRp] = nodes.AllRoles
	tuningToRoles[OpensearchHeapSize] = nodes.AllRoles
	tuningToRoles[SenlinDebugEnabled] = nodes.AllRoles
	tuningToRoles[SkylineDebugEnabled] = nodes.AllRoles
	tuningToRoles[SnapshotApplyAction] = nodes.AllRoles
	tuningToRoles[SnapshotApplyPolicyIgnore] = nodes.AllRoles
	tuningToRoles[SshdBindToAllInterfaces] = nodes.AllRoles
	tuningToRoles[SshdSessionInactivity] = nodes.AllRoles
	tuningToRoles[TimeTimezone] = nodes.AllRoles
	tuningToRoles[UpdateSecurityAutoUpdate] = nodes.AllRoles
	tuningToRoles[WatcherDebugEnabled] = nodes.AllRoles
}

func GetTuningValue(name string) (string, error) {
	out, err := exec.Command("hex_tuning_helper", conf.Opts.Spec.Identity.Policy, "", name).Output()
	if err != nil {
		log.Errorf("tunings: failed to read hex tuning value: %v", err)
		return "", err
	}

	keyValue := strings.Split(string(out), "'")
	if len(keyValue) < 2 {
		return "", errors.ErrTuningNotFound
	}

	return keyValue[1], nil
}

func GetSourceTuning(name string) (*tunings.Tuning, error) {
	policy, err := GetTuningPolicy(TuningPolicyFile)
	if err != nil {
		return nil, err
	}

	for _, tuning := range policy.Tunings {
		if tuning.Name == name {
			return &tuning, nil
		}
	}

	return nil, errors.ErrTuningNotFound
}

func ApplyTuning(isolatedDir string) error {
	out, err := exec.Command("hex_config", "apply", isolatedDir).CombinedOutput()
	if err != nil {
		log.Errorf("tunings: failed to apply hex tuning value: %s", string(out))
		return err
	}

	return nil
}

func IsTuningApplied(tuning tunings.Tuning) error {
	maxTries := 10
	for range maxTries {
		if isValueApplied(tuning) {
			return nil
		}

		wait.Seconds(2)
	}

	return fmt.Errorf(
		"tuning: %s's value(%s) is not applied",
		tuning.Name,
		tuning.StrValue(),
	)
}

func ApplyTunings(tunings []tunings.Tuning) error {
	newTunings, err := genTuningsAsYaml(tunings)
	if err != nil {
		return err
	}

	tmpTuningDir := genTmpTuningDir()
	err = writeTuningToFile(tmpTuningDir, newTunings)
	if err != nil {
		return err
	}

	err = ApplyTuning(tmpTuningDir)
	if err != nil {
		return err
	}

	return nil
}

func GetTuningPolicy(filePath string) (*tunings.Policy, error) {
	b, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	policy := &tunings.Policy{}
	err = yaml.Unmarshal(b, policy)
	if err != nil {
		return nil, err
	}

	return policy, nil
}

func IsTuningDeleted(tuning tunings.Tuning) bool {
	_, valueErr := GetTuningValue(tuning.Name)
	policy, policyErr := GetTuningPolicy(TuningPolicyFile)
	if policyErr != nil {
		return false
	}

	return !policy.HasMatchedTuning(tuning) &&
		noValueInSettings(valueErr)
}

func ListTunings(opts tunings.ListOptions) ([]tunings.Tuning, error) {
	localTunings := tunings.ListLocal()
	if !opts.AllNodes {
		return localTunings, nil
	}

	allTunings, err := ListTuningsFromOtherNodes()
	if err != nil {
		return nil, err
	}

	allTunings[base.Hostname] = localTunings
	return aggregateTunings(allTunings), nil
}

func ListTuningsFromOtherNodes() (map[string][]tunings.Tuning, error) {
	nodeTunings := map[string][]tunings.Tuning{}
	for _, node := range nodes.List() {
		if node.IsLocal() {
			continue
		}

		if node.IsDown() {
			continue
		}

		tunings, err := getNodeTunings(node)
		if err != nil {
			log.Errorf("tunings: failed to get tunings from node %s: %v", node.Hostname, err)
			continue
		}

		nodeTunings[node.Hostname] = tunings
	}

	return nodeTunings, nil
}

func SyncTunings() {
	for _, spec := range tunings.ListSpecs() {
		srcTuning, err := GetSourceTuning(spec.Name)
		if err == nil {
			srcTuning.IsModified = true
			srcTuning.Description = spec.Description
			srcTuning.Limitation = spec.Limitation
			srcTuning.Hosts = []nodes.Host{{Name: base.Hostname, Ip: base.AdvertiseIp}}
			checkAndUpdateTuning(spec.Name, *srcTuning)
		}

		if errors.Is(err, errors.ErrTuningNotFound) {
			setDefaultTuning(spec)
		}
	}
}

func setTuningToSelectors() {
	tuningToSelectors[NovaGpuType] = nodes.Selector{
		Enabled: true,
		Labels:  map[string]string{"isGpuEnabled": "true"},
	}
}

func setTuningSpecs() {
	out, err := exec.Command("hex_sdk", "-f", "json", "tuning_dump").Output()
	if err != nil {
		log.Errorf("tunings: failed to get tuning specs: %v", err)
		return
	}

	rawSpecs := []tunings.RawSpec{}
	err = json.Unmarshal(out, &rawSpecs)
	if err != nil {
		log.Errorf("tunings: failed to unmarshal tuning specs: %v", err)
		return
	}

	for _, rawtuningSpec := range rawSpecs {
		tunings.SetSpec(rawtuningSpec.Name, convertToTuningSpec(rawtuningSpec))
	}
}

func convertToTuningSpec(rawSpec tunings.RawSpec) *tunings.Spec {
	spec := &tunings.Spec{
		Name:        rawSpec.Name,
		Description: rawSpec.Description,
		Limitation:  convertLimit(rawSpec),
		Roles:       getTuningRoles(rawSpec.Name),
		Selector:    getTuningSelectors(rawSpec.Name),
	}

	return spec
}

func convertLimit(raw tunings.RawSpec) tunings.Limitation {
	switch raw.Limitation.Type {
	case "int", "uint":
		return convertIntLimit(raw.Limitation)
	case "bool", "boolean":
		return convertBoolLimit(raw.Limitation)
	case "str":
		return convertStringLimit(raw.Limitation)
	}

	return tunings.Limitation{
		Type:    fmt.Sprintf("invalid tuning spec from cos(%s)", raw.Limitation.Type),
		Default: raw.Limitation.Default,
	}
}

func getTuningRoles(tuningName string) []*nodes.Role {
	roles, found := tuningToRoles[tuningName]
	if found {
		return roles
	}

	return nodes.AllRoles
}

func getTuningSelectors(tuningName string) nodes.Selector {
	selectors, found := tuningToSelectors[tuningName]
	if found {
		return selectors
	}

	return nodes.Selector{}
}

func convertIntLimit(raw tunings.RawLimitation) tunings.Limitation {
	defaultVal, err := strconv.Atoi(raw.Default)
	if err != nil {
		log.Errorf("tunings: failed to convert default value %s to int: %v", raw.Default, err)
	}

	min, err := strconv.Atoi(raw.Min)
	if err != nil {
		log.Errorf("tunings: failed to convert min value %s to int: %v", raw.Min, err)
		min = 0
	}

	max, err := strconv.Atoi(raw.Max)
	if err != nil {
		log.Errorf("tunings: failed to convert max value %s to int: %v", raw.Max, err)
		max = 0
	}

	return tunings.Limitation{
		Type:    raw.Type,
		Default: defaultVal,
		Min:     &min,
		Max:     &max,
		Regex:   raw.Regex,
	}
}

func convertBoolLimit(raw tunings.RawLimitation) tunings.Limitation {
	defaultVal, err := strconv.ParseBool(strings.ToLower(raw.Default))
	if err != nil {
		log.Errorf("tunings: failed to convert default value %s to bool: %v", raw.Default, err)
		defaultVal = false
	}

	return tunings.Limitation{
		Type:    raw.Type,
		Default: defaultVal,
		Regex:   raw.Regex,
	}
}

func convertStringLimit(raw tunings.RawLimitation) tunings.Limitation {
	return tunings.Limitation{
		Type:    raw.Type,
		Default: raw.Default,
		Regex:   raw.Regex,
	}
}

func isValueApplied(tuning tunings.Tuning) bool {
	value, err := GetTuningValue(tuning.Name)
	if tuning.Enabled {
		return tuning.StrValue() == value
	}

	if err == nil {
		return false
	}

	if !noValueInSettings(err) {
		return false
	}

	policy, err := GetTuningPolicy(TuningPolicyFile)
	if err != nil {
		return false
	}

	return policy.HasMatchedTuning(tuning)
}

func noValueInSettings(err error) bool {
	return errors.Is(err, errors.ErrTuningNotFound)
}

func genTuningsAsYaml(list []tunings.Tuning) ([]byte, error) {
	tuningTemplate := tunings.Policy{
		Name:    "tuning",
		Version: "1.0",
		Enabled: true,
		Tunings: list,
	}

	yml, err := yaml.Marshal(&tuningTemplate)
	if err != nil {
		log.Errorf("tunings: failed to marshal batch tuning info: %v", err)
		return nil, err
	}

	return yml, nil
}

func writeTuningToFile(tmpDir string, yml []byte) error {
	fullDir := fmt.Sprintf("%s/tuning", tmpDir)
	err := os.MkdirAll(fullDir, 0755)
	if err != nil {
		log.Errorf("tunings: failed to create isolated tuning directory: %v", err)
		return err
	}

	file, err := os.Create(fmt.Sprintf("%s/tuning1_0.yml", fullDir))
	if err != nil {
		log.Errorf("tunings: failed to create isolated tuning file: %v", err)
		return err
	}

	defer file.Close()
	_, err = io.Writer.Write(file, yml)
	if err != nil {
		log.Errorf("tunings: failed to write tuning info to isolated file: %v", err)
		return err
	}

	return nil
}

func genTmpTuningDir() string {
	hash := uuid.New().String()[:8]
	return fmt.Sprintf("/tmp/tuning-%s", hash)
}

func getNodeTunings(node nodes.Node) ([]tunings.Tuning, error) {
	h := http.GetGlobalHelper()
	resp, err := h.R().
		SetResult(&bodies.TuningList{}).
		SetHeaders(nodes.GetSecretHeaders()).
		Get(node.GetTuningUrl())
	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, fmt.Errorf(
			"tunings: failed to get tunings from %s: %s",
			node.Hostname,
			string(resp.Body()),
		)
	}

	list := resp.Result().(*bodies.TuningList)
	return list.Data.Tunings, nil
}

func aggregateTunings(nodeToTuning map[string][]tunings.Tuning) []tunings.Tuning {
	mergedMap := make(map[string]tunings.Tuning)
	for _, tunings := range nodeToTuning {
		setTunings(mergedMap, tunings)
	}

	tunings := []tunings.Tuning{}
	for _, item := range mergedMap {
		tunings = append(tunings, item)
	}

	return tunings
}

func setTunings(mergedMap map[string]tunings.Tuning, tunings []tunings.Tuning) {
	for _, tuning := range tunings {
		key := tuning.SearchKey()
		existing, found := mergedMap[key]
		if found {
			existing.Hosts = slices.Concat(existing.Hosts, tuning.Hosts)
			mergedMap[key] = existing
		} else {
			mergedMap[key] = tuning
		}
	}
}

func checkAndUpdateTuning(key string, sourceTuning tunings.Tuning) {
	tuning := tunings.Get(key)
	if !isTuningChanged(tuning, sourceTuning) {
		return
	}

	tunings.SetLocal(sourceTuning)
}

func isTuningChanged(tuning, fileTuning tunings.Tuning) bool {
	if tuning.Value != fileTuning.Value {
		return true
	}

	if tuning.Enabled != fileTuning.Enabled {
		return true
	}

	return false
}

func setDefaultTuning(tuning tunings.Spec) {
	tunings.SetLocal(tunings.Tuning{
		Enabled:     true,
		Name:        tuning.Name,
		Value:       tuning.Limitation.Default,
		Hosts:       []nodes.Host{{Name: base.Hostname, Ip: base.AdvertiseIp}},
		Description: tuning.Description,
		Limitation:  tuning.Limitation,
		IsModified:  false,
	})
}
