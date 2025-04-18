package cubecos

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"slices"
	"strconv"
	"strings"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/http"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	"github.com/bigstack-oss/cube-cos-api/internal/api"
	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	cuberr "github.com/bigstack-oss/cube-cos-api/internal/errors"
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
	tuningToRoles     = map[string][]*v1.Role{}
	tuningToSelectors = map[string]v1.Selector{}
)

func init() {
	setTuningToRoles()
	setTuningToSelectors()
}

func setTuningToRoles() {
	tuningToRoles[BarbicanDebugEnabled] = v1.AllGeneralRoles
	tuningToRoles[CephDebugEnabled] = v1.AllGeneralRoles
	tuningToRoles[CephMirrorMetaSync] = v1.ControlRoles
	tuningToRoles[CinderBackupAccount] = v1.AllGeneralRoles
	tuningToRoles[CinderBackupEndpoint] = v1.AllRoles
	tuningToRoles[CinderBackupOverride] = v1.AllRoles
	tuningToRoles[CinderBackupPool] = v1.AllRoles
	tuningToRoles[CinderBackupSecret] = v1.AllRoles
	tuningToRoles[CinderBackupType] = v1.AllRoles
	tuningToRoles[CinderDebugEnabled] = v1.AllRoles
	tuningToRoles[CinderExternalAccount] = v1.AllRoles
	tuningToRoles[CinderExternalDriver] = v1.AllRoles
	tuningToRoles[CinderExternalEndpoint] = v1.AllRoles
	tuningToRoles[CinderExternalName] = v1.AllRoles
	tuningToRoles[CinderExternalPool] = v1.AllRoles
	tuningToRoles[CinderExternalSecret] = v1.AllRoles
	tuningToRoles[CubesysAlertLevel] = v1.AllRoles
	tuningToRoles[CubesysAlertLevelS] = v1.AllRoles
	tuningToRoles[CubesysConntableMax] = v1.AllRoles
	tuningToRoles[CubesysLogDefaultRetention] = v1.AllRoles
	tuningToRoles[CubesysProviderExtra] = v1.AllRoles
	tuningToRoles[CyborgDebugEnabled] = v1.AllRoles
	tuningToRoles[DebugEnableCoreDumpS] = v1.AllRoles
	tuningToRoles[DebugEnableKdump] = v1.AllRoles
	tuningToRoles[DebugLevelS] = v1.AllRoles
	tuningToRoles[DebugMaxCoreDump] = v1.AllRoles
	tuningToRoles[DesignateDebugEnabled] = v1.AllRoles
	tuningToRoles[GlanceDebugEnabled] = v1.AllRoles
	tuningToRoles[GlanceExportRp] = v1.AllRoles
	tuningToRoles[HeatDebugEnabled] = v1.AllRoles
	tuningToRoles[InfluxdbCuratorRp] = v1.AllRoles
	tuningToRoles[IronicDebugEnabled] = v1.AllGeneralRoles
	tuningToRoles[IronicDeployServer] = v1.AllGeneralRoles
	tuningToRoles[KapacitorAlertCheckEnabled] = v1.ControlRoles
	tuningToRoles[KapacitorAlertCheckEventId] = v1.ControlRoles
	tuningToRoles[KapacitorAlertCheckInterval] = v1.ControlRoles
	tuningToRoles[KapacitorAlertExtraPrefix] = v1.ControlRoles
	tuningToRoles[KapacitorAlertFlowBase] = v1.ControlRoles
	tuningToRoles[KapacitorAlertFlowThreshold] = v1.ControlRoles
	tuningToRoles[KapacitorAlertFlowUnit] = v1.ControlRoles
	tuningToRoles[KeystoneDebugEnabled] = v1.AllRoles
	tuningToRoles[ManilaDebugEnabled] = v1.AllRoles
	tuningToRoles[ManilaVolumeType] = v1.AllRoles
	tuningToRoles[MasakariHostEvacuateAll] = v1.AllRoles
	tuningToRoles[MasakariWaitPeriod] = v1.ControlRoles
	tuningToRoles[MonascaDebugEnabled] = v1.AllRoles
	tuningToRoles[MysqlBackupCuratorRp] = v1.AllRoles
	tuningToRoles[NetIfMtuName] = v1.AllRoles
	tuningToRoles[NetIpv4TcpSyncookies] = v1.AllRoles
	tuningToRoles[NetLacpDefaultRate] = v1.AllRoles
	tuningToRoles[NetLacpDefaultXmit] = v1.AllRoles
	tuningToRoles[NeutronDebugEnabled] = v1.AllRoles
	tuningToRoles[NovaControlHostMemory] = v1.ComputeRoles
	tuningToRoles[NovaControlHostVcpu] = []*v1.Role{v1.GetControlConvergeRole(), v1.GetEdgeCoreRole()}
	tuningToRoles[NovaDebugEnabled] = v1.AllRoles
	tuningToRoles[NovaGpuType] = v1.ComputeRoles
	tuningToRoles[NovaOvercommitCpuRatio] = v1.AllRoles
	tuningToRoles[NovaOvercommitDiskRatio] = v1.AllRoles
	tuningToRoles[NovaOvercommitRamRatio] = v1.AllRoles
	tuningToRoles[NtpDebugEnabled] = v1.AllRoles
	tuningToRoles[OctaviaDebugEnabled] = v1.AllRoles
	tuningToRoles[OctaviaHa] = v1.AllRoles
	tuningToRoles[OpensearchCuratorRp] = v1.AllRoles
	tuningToRoles[OpensearchHeapSize] = v1.AllRoles
	tuningToRoles[SenlinDebugEnabled] = v1.AllRoles
	tuningToRoles[SkylineDebugEnabled] = v1.AllRoles
	tuningToRoles[SnapshotApplyAction] = v1.AllRoles
	tuningToRoles[SnapshotApplyPolicyIgnore] = v1.AllRoles
	tuningToRoles[SshdBindToAllInterfaces] = v1.AllRoles
	tuningToRoles[SshdSessionInactivity] = v1.AllRoles
	tuningToRoles[TimeTimezone] = v1.AllRoles
	tuningToRoles[UpdateSecurityAutoUpdate] = v1.AllRoles
	tuningToRoles[WatcherDebugEnabled] = v1.AllRoles
}

func setTuningToSelectors() {
	tuningToSelectors[NovaGpuType] = v1.Selector{
		Enabled: true,
		Labels:  map[string]string{"isGpuEnabled": "true"},
	}
}

func setTuningSpecs() {
	out, err := exec.Command("hex_sdk", "-f", "json", "tuning_dump").Output()
	if err != nil {
		log.Errorf("tunings: failed to get tuning specs: %s", err.Error())
		return
	}

	rawSpecs := []v1.RawTuningSpec{}
	err = json.Unmarshal(out, &rawSpecs)
	if err != nil {
		log.Errorf("tunings: failed to unmarshal tuning specs: %s", err.Error())
		return
	}

	for _, rawtuningSpec := range rawSpecs {
		v1.SetTuningSpec(rawtuningSpec.Name, convertToTuningSpec(rawtuningSpec))
	}
}

func convertToTuningSpec(rawSpec v1.RawTuningSpec) *v1.TuningSpec {
	spec := &v1.TuningSpec{
		Name:        rawSpec.Name,
		Description: rawSpec.Description,
		Limitation:  convertLimit(rawSpec),
		Roles:       getTuningRoles(rawSpec.Name),
		Selector:    getTuningSelectors(rawSpec.Name),
	}

	return spec
}

func convertLimit(raw v1.RawTuningSpec) v1.TuningLimitation {
	switch raw.Limitation.Type {
	case "int", "uint":
		return convertIntLimit(raw.Limitation)
	case "boolean":
		return convertBoolLimit(raw.Limitation)
	case "str":
		return convertStringLimit(raw.Limitation)
	}

	return v1.TuningLimitation{
		Type:    fmt.Sprintf("invalid tuning spec from cos(%s)", raw.Limitation.Type),
		Default: raw.Limitation.Default,
	}
}

func getTuningRoles(tuningName string) []*v1.Role {
	roles, found := tuningToRoles[tuningName]
	if found {
		return roles
	}

	return v1.AllRoles
}

func getTuningSelectors(tuningName string) v1.Selector {
	selectors, found := tuningToSelectors[tuningName]
	if found {
		return selectors
	}

	return v1.Selector{}
}

func convertIntLimit(raw v1.RawTuningLimitation) v1.TuningLimitation {
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

	return v1.TuningLimitation{
		Type:    raw.Type,
		Default: defaultVal,
		Min:     &min,
		Max:     &max,
		Regex:   raw.Regex,
	}
}

func convertBoolLimit(raw v1.RawTuningLimitation) v1.TuningLimitation {
	defaultVal, err := strconv.ParseBool(strings.ToLower(raw.Default))
	if err != nil {
		log.Errorf("tunings: failed to convert default value %s to bool: %v", raw.Default, err)
		defaultVal = false
	}

	return v1.TuningLimitation{
		Type:    raw.Type,
		Default: defaultVal,
		Regex:   raw.Regex,
	}
}

func convertStringLimit(raw v1.RawTuningLimitation) v1.TuningLimitation {
	return v1.TuningLimitation{
		Type:    raw.Type,
		Default: raw.Default,
		Regex:   raw.Regex,
	}
}

func init() {
	setTuningSpecs()
}

func GetTuningValue(name string) (string, error) {
	out, err := exec.Command("hex_tuning_helper", "/etc/settings.txt", "", name).Output()
	if err != nil {
		log.Errorf("tunings: failed to read hex tuning value: %s", err.Error())
		return "", err
	}

	keyValue := strings.Split(string(out), "'")
	if len(keyValue) < 2 {
		return "", cuberr.TuningNotFound
	}

	return keyValue[1], nil
}

func GetTuning(name string) (*v1.Tuning, error) {
	policy, err := GetTuningPolicy(TuningPolicyFile)
	if err != nil {
		return nil, err
	}

	for _, tuning := range policy.Tunings {
		if tuning.Name == name {
			return &tuning, nil
		}
	}

	return nil, cuberr.TuningNotFound
}

func ApplyTuning(isolatedDir string) error {
	out, err := exec.Command("hex_config", "apply", isolatedDir).CombinedOutput()
	if err != nil {
		log.Errorf("tunings: failed to apply hex tuning value: %s", string(out))
		return err
	}

	return nil
}

func IsTuningApplied(tuning v1.Tuning) error {
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

func isValueApplied(tuning v1.Tuning) bool {
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
	return errors.Is(err, cuberr.TuningNotFound)
}

func ApplyTunings(tunings []v1.Tuning) error {
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

func genTuningsAsYaml(tunings []v1.Tuning) ([]byte, error) {
	tuningTemplate := v1.TuningPolicy{
		Name:    "tuning",
		Version: "1.0",
		Enabled: true,
		Tunings: tunings,
	}

	yml, err := yaml.Marshal(&tuningTemplate)
	if err != nil {
		log.Errorf("tunings: failed to marshal batch tuning info: %s", err.Error())
		return nil, err
	}

	return yml, nil
}

func writeTuningToFile(tmpDir string, yml []byte) error {
	fullDir := fmt.Sprintf("%s/tuning", tmpDir)
	err := os.MkdirAll(fullDir, 0755)
	if err != nil {
		log.Errorf("tunings: failed to create isolated tuning directory: %s", err.Error())
		return err
	}

	file, err := os.Create(fmt.Sprintf("%s/tuning1_0.yml", fullDir))
	if err != nil {
		log.Errorf("tunings: failed to create isolated tuning file: %s", err.Error())
		return err
	}

	defer file.Close()
	_, err = io.Writer.Write(file, yml)
	if err != nil {
		log.Errorf("tunings: failed to write tuning info to isolated file: %s", err.Error())
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

func GetTuningPolicy(filePath string) (*v1.TuningPolicy, error) {
	b, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	policy := &v1.TuningPolicy{}
	err = yaml.Unmarshal(b, policy)
	if err != nil {
		return nil, err
	}

	return policy, nil
}

func IsTuningDeleted(tuning v1.Tuning) bool {
	_, valueErr := GetTuningValue(tuning.Name)
	policy, policyErr := GetTuningPolicy(TuningPolicyFile)
	if policyErr != nil {
		return false
	}

	return !policy.HasMatchedTuning(tuning) &&
		noValueInSettings(valueErr)
}

func ListTunings(opts v1.ListTuningOptions) ([]v1.Tuning, error) {
	localTunings := v1.ListLocalTunings()
	if !opts.AllNodes {
		return localTunings, nil
	}

	allTunings, err := ListTuningsFromOtherNodes()
	if err != nil {
		return nil, err
	}

	allTunings[v1.Hostname] = localTunings
	return aggregateTunings(allTunings), nil
}

func ListTuningsFromOtherNodes() (map[string][]v1.Tuning, error) {
	nodeTunings := map[string][]v1.Tuning{}
	for _, node := range v1.ListNodes() {
		if node.IsLocal() {
			continue
		}

		tunings, err := getNodeTunings(node)
		if err != nil {
			log.Errorf("tunings: failed to get tunings from node %s: %s", node.Hostname, err.Error())
			continue
		}

		nodeTunings[node.Hostname] = tunings
	}

	return nodeTunings, nil
}

func getNodeTunings(node v1.Node) ([]v1.Tuning, error) {
	h := http.GetGlobalHelper()
	resp, err := h.R().
		SetResult(&api.TuningListData{}).
		SetHeader(node.GenAuthHeader()).
		Get(node.GetTuningUrl())
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, fmt.Errorf(
			"tunings: failed to get tunings from %s: %d %s",
			node.Hostname,
			resp.StatusCode(),
			string(resp.Body()),
		)
	}

	tuningList := resp.Result().(*api.TuningListData)
	return tuningList.Data.Tunings, nil
}

func aggregateTunings(nodeToTuning map[string][]v1.Tuning) []v1.Tuning {
	mergedMap := make(map[string]v1.Tuning)
	for _, tunings := range nodeToTuning {
		setTunings(mergedMap, tunings)
	}

	tunings := []v1.Tuning{}
	for _, item := range mergedMap {
		tunings = append(tunings, item)
	}

	return tunings
}

func setTunings(mergedMap map[string]v1.Tuning, tunings []v1.Tuning) {
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

func SyncTunings() {
	for _, spec := range v1.ListTuningSpecs() {
		srcTuning, err := GetTuning(spec.Name)
		if err == nil {
			srcTuning.IsModified = true
			srcTuning.Description = spec.Description
			srcTuning.Limitation = spec.Limitation
			srcTuning.Hosts = []v1.Host{{Name: v1.Hostname, Ip: v1.AdvertiseIp}}
			checkAndUpdateTuning(spec.Name, *srcTuning)
		}

		if errors.Is(err, cuberr.TuningNotFound) {
			setDefaultTuning(spec)
		}
	}
}

func checkAndUpdateTuning(key string, sourceTuning v1.Tuning) {
	tuning := v1.GetLocalTuning(key)
	if !isTuningChanged(tuning, sourceTuning) {
		return
	}

	v1.SetLocalTuning(sourceTuning)
}

func isTuningChanged(tuning, fileTuning v1.Tuning) bool {
	if tuning.Value != fileTuning.Value {
		return true
	}

	if tuning.Enabled != fileTuning.Enabled {
		return true
	}

	return false
}

func setDefaultTuning(tuning v1.TuningSpec) {
	v1.SetLocalTuning(v1.Tuning{
		Enabled:     true,
		Name:        tuning.Name,
		Value:       tuning.Limitation.Default,
		Hosts:       []v1.Host{{Name: v1.Hostname, Ip: v1.AdvertiseIp}},
		Description: tuning.Description,
		Limitation:  tuning.Limitation,
		IsModified:  false,
	})
}
