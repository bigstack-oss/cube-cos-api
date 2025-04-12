package cubecos

import (
	"encoding/json"
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
	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	cuberr "github.com/bigstack-oss/cube-cos-api/internal/errors"
	"github.com/google/uuid"
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
	tuningToRoles     = map[string][]*definition.Role{}
	tuningToSelectors = map[string]definition.Selector{}
)

func init() {
	setTuningToRoles()
	setTuningToSelectors()
}

func setTuningToRoles() {
	tuningToRoles[BarbicanDebugEnabled] = definition.AllGeneralRoles
	tuningToRoles[CephDebugEnabled] = definition.AllGeneralRoles
	tuningToRoles[CephMirrorMetaSync] = definition.ControlRoles
	tuningToRoles[CinderBackupAccount] = definition.AllGeneralRoles
	tuningToRoles[CinderBackupEndpoint] = definition.AllRoles
	tuningToRoles[CinderBackupOverride] = definition.AllRoles
	tuningToRoles[CinderBackupPool] = definition.AllRoles
	tuningToRoles[CinderBackupSecret] = definition.AllRoles
	tuningToRoles[CinderBackupType] = definition.AllRoles
	tuningToRoles[CinderDebugEnabled] = definition.AllRoles
	tuningToRoles[CinderExternalAccount] = definition.AllRoles
	tuningToRoles[CinderExternalDriver] = definition.AllRoles
	tuningToRoles[CinderExternalEndpoint] = definition.AllRoles
	tuningToRoles[CinderExternalName] = definition.AllRoles
	tuningToRoles[CinderExternalPool] = definition.AllRoles
	tuningToRoles[CinderExternalSecret] = definition.AllRoles
	tuningToRoles[CubesysAlertLevel] = definition.AllRoles
	tuningToRoles[CubesysAlertLevelS] = definition.AllRoles
	tuningToRoles[CubesysConntableMax] = definition.AllRoles
	tuningToRoles[CubesysLogDefaultRetention] = definition.AllRoles
	tuningToRoles[CubesysProviderExtra] = definition.AllRoles
	tuningToRoles[CyborgDebugEnabled] = definition.AllRoles
	tuningToRoles[DebugEnableCoreDumpS] = definition.AllRoles
	tuningToRoles[DebugEnableKdump] = definition.AllRoles
	tuningToRoles[DebugLevelS] = definition.AllRoles
	tuningToRoles[DebugMaxCoreDump] = definition.AllRoles
	tuningToRoles[DesignateDebugEnabled] = definition.AllRoles
	tuningToRoles[GlanceDebugEnabled] = definition.AllRoles
	tuningToRoles[GlanceExportRp] = definition.AllRoles
	tuningToRoles[HeatDebugEnabled] = definition.AllRoles
	tuningToRoles[InfluxdbCuratorRp] = definition.AllRoles
	tuningToRoles[IronicDebugEnabled] = definition.AllGeneralRoles
	tuningToRoles[IronicDeployServer] = definition.AllGeneralRoles
	tuningToRoles[KapacitorAlertCheckEnabled] = definition.ControlRoles
	tuningToRoles[KapacitorAlertCheckEventId] = definition.ControlRoles
	tuningToRoles[KapacitorAlertCheckInterval] = definition.ControlRoles
	tuningToRoles[KapacitorAlertExtraPrefix] = definition.ControlRoles
	tuningToRoles[KapacitorAlertFlowBase] = definition.ControlRoles
	tuningToRoles[KapacitorAlertFlowThreshold] = definition.ControlRoles
	tuningToRoles[KapacitorAlertFlowUnit] = definition.ControlRoles
	tuningToRoles[KeystoneDebugEnabled] = definition.AllRoles
	tuningToRoles[ManilaDebugEnabled] = definition.AllRoles
	tuningToRoles[ManilaVolumeType] = definition.AllRoles
	tuningToRoles[MasakariHostEvacuateAll] = definition.AllRoles
	tuningToRoles[MasakariWaitPeriod] = definition.ControlRoles
	tuningToRoles[MonascaDebugEnabled] = definition.AllRoles
	tuningToRoles[MysqlBackupCuratorRp] = definition.AllRoles
	tuningToRoles[NetIfMtuName] = definition.AllRoles
	tuningToRoles[NetIpv4TcpSyncookies] = definition.AllRoles
	tuningToRoles[NetLacpDefaultRate] = definition.AllRoles
	tuningToRoles[NetLacpDefaultXmit] = definition.AllRoles
	tuningToRoles[NeutronDebugEnabled] = definition.AllRoles
	tuningToRoles[NovaControlHostMemory] = definition.ComputeRoles
	tuningToRoles[NovaControlHostVcpu] = []*definition.Role{definition.GetControlConvergeRole(), definition.GetEdgeCoreRole()}
	tuningToRoles[NovaDebugEnabled] = definition.AllRoles
	tuningToRoles[NovaGpuType] = definition.ComputeRoles
	tuningToRoles[NovaOvercommitCpuRatio] = definition.AllRoles
	tuningToRoles[NovaOvercommitDiskRatio] = definition.AllRoles
	tuningToRoles[NovaOvercommitRamRatio] = definition.AllRoles
	tuningToRoles[NtpDebugEnabled] = definition.AllRoles
	tuningToRoles[OctaviaDebugEnabled] = definition.AllRoles
	tuningToRoles[OctaviaHa] = definition.AllRoles
	tuningToRoles[OpensearchCuratorRp] = definition.AllRoles
	tuningToRoles[OpensearchHeapSize] = definition.AllRoles
	tuningToRoles[SenlinDebugEnabled] = definition.AllRoles
	tuningToRoles[SkylineDebugEnabled] = definition.AllRoles
	tuningToRoles[SnapshotApplyAction] = definition.AllRoles
	tuningToRoles[SnapshotApplyPolicyIgnore] = definition.AllRoles
	tuningToRoles[SshdBindToAllInterfaces] = definition.AllRoles
	tuningToRoles[SshdSessionInactivity] = definition.AllRoles
	tuningToRoles[TimeTimezone] = definition.AllRoles
	tuningToRoles[UpdateSecurityAutoUpdate] = definition.AllRoles
	tuningToRoles[WatcherDebugEnabled] = definition.AllRoles
}

func setTuningToSelectors() {
	tuningToSelectors[NovaGpuType] = definition.Selector{
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

	rawSpecs := []definition.RawTuningSpec{}
	err = json.Unmarshal(out, &rawSpecs)
	if err != nil {
		log.Errorf("tunings: failed to unmarshal tuning specs: %s", err.Error())
		return
	}

	for _, rawtuningSpec := range rawSpecs {
		definition.SetTuningSpec(rawtuningSpec.Name, convertToTuningSpec(rawtuningSpec))
	}
}

func convertToTuningSpec(rawSpec definition.RawTuningSpec) *definition.TuningSpec {
	spec := &definition.TuningSpec{
		Name:        rawSpec.Name,
		Description: rawSpec.Description,
		Limitation:  convertLimit(rawSpec),
		Roles:       getTuningRoles(rawSpec.Name),
		Selector:    getTuningSelectors(rawSpec.Name),
	}

	return spec
}

func convertLimit(raw definition.RawTuningSpec) definition.TuningLimitation {
	switch raw.Limitation.Type {
	case "int", "uint":
		return convertIntLimit(raw.Limitation)
	case "boolean":
		return convertBoolLimit(raw.Limitation)
	case "str":
		return convertStringLimit(raw.Limitation)
	}

	return definition.TuningLimitation{
		Type:    fmt.Sprintf("invalid tuning spec from cos(%s)", raw.Limitation.Type),
		Default: raw.Limitation.Default,
	}
}

func getTuningRoles(tuningName string) []*definition.Role {
	roles, found := tuningToRoles[tuningName]
	if found {
		return roles
	}

	return definition.AllRoles
}

func getTuningSelectors(tuningName string) definition.Selector {
	selectors, found := tuningToSelectors[tuningName]
	if found {
		return selectors
	}

	return definition.Selector{}
}

func convertIntLimit(raw definition.RawTuningLimitation) definition.TuningLimitation {
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

	return definition.TuningLimitation{
		Type:    raw.Type,
		Default: defaultVal,
		Min:     &min,
		Max:     &max,
		Regex:   raw.Regex,
	}
}

func convertBoolLimit(raw definition.RawTuningLimitation) definition.TuningLimitation {
	defaultVal, err := strconv.ParseBool(strings.ToLower(raw.Default))
	if err != nil {
		log.Errorf("tunings: failed to convert default value %s to bool: %v", raw.Default, err)
		defaultVal = false
	}

	return definition.TuningLimitation{
		Type:    raw.Type,
		Default: defaultVal,
		Regex:   raw.Regex,
	}
}

func convertStringLimit(raw definition.RawTuningLimitation) definition.TuningLimitation {
	return definition.TuningLimitation{
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

func GetTuning(name string) (*definition.Tuning, error) {
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

func IsTuningApplied(tuning definition.Tuning) error {
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

func isValueApplied(tuning definition.Tuning) bool {
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

func ApplyTunings(tunings []definition.Tuning) error {
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

func genTuningsAsYaml(tunings []definition.Tuning) ([]byte, error) {
	tuningTemplate := definition.TuningPolicy{
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

func GetTuningPolicy(filePath string) (*definition.TuningPolicy, error) {
	b, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	policy := &definition.TuningPolicy{}
	err = yaml.Unmarshal(b, policy)
	if err != nil {
		return nil, err
	}

	return policy, nil
}

func IsTuningDeleted(tuning definition.Tuning) bool {
	_, valueErr := GetTuningValue(tuning.Name)
	policy, policyErr := GetTuningPolicy(TuningPolicyFile)
	if policyErr != nil {
		return false
	}

	return !policy.HasMatchedTuning(tuning) &&
		noValueInSettings(valueErr)
}

func ListTunings(opts definition.ListTuningOptions) ([]definition.Tuning, error) {
	localTunings := definition.ListLocalTunings()
	if !opts.AllNodes {
		return localTunings, nil
	}

	allTunings, err := ListTuningsFromOtherNodes()
	if err != nil {
		return nil, err
	}

	allTunings[definition.Hostname] = localTunings
	return aggregateTunings(allTunings), nil
}

func ListTuningsFromOtherNodes() (map[string][]definition.Tuning, error) {
	nodeTunings := map[string][]definition.Tuning{}
	for _, node := range definition.ListNodes() {
		if node.IsLocal() {
			continue
		}

		tunings, err := getNodeTunings(node)
		if err != nil {
			log.Errorf("tunings: failed to get tunings from node %s: %s", node.Name, err.Error())
			continue
		}

		nodeTunings[node.Name] = tunings
	}

	return nodeTunings, nil
}

func getNodeTunings(node definition.Node) ([]definition.Tuning, error) {
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
			"failed to get tunings from %s: %d %s",
			node.Hostname,
			resp.StatusCode(),
			string(resp.Body()),
		)
	}

	tuningList := resp.Result().(*api.TuningListData)
	return tuningList.Data, nil
}

func aggregateTunings(nodeToTuning map[string][]definition.Tuning) []definition.Tuning {
	mergedMap := make(map[string]definition.Tuning)
	for _, tunings := range nodeToTuning {
		setTunings(mergedMap, tunings)
	}

	tunings := []definition.Tuning{}
	for _, item := range mergedMap {
		tunings = append(tunings, item)
	}

	return tunings
}

func setTunings(mergedMap map[string]definition.Tuning, tunings []definition.Tuning) {
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
	for _, spec := range definition.ListTuningSpecs() {
		srcTuning, err := GetTuning(spec.Name)
		if err == nil {
			srcTuning.IsModified = true
			srcTuning.Description = spec.Description
			srcTuning.Limitation = spec.Limitation
			srcTuning.Hosts = []definition.Host{{Name: definition.Hostname, Ip: definition.AdvertiseIp}}
			checkAndUpdateTuning(spec.Name, *srcTuning)
		}

		if errors.Is(err, cuberr.TuningNotFound) {
			setDefaultTuning(spec)
		}
	}
}

func checkAndUpdateTuning(key string, sourceTuning definition.Tuning) {
	tuning := definition.GetLocalTuning(key)
	if !isTuningChanged(tuning, sourceTuning) {
		return
	}

	definition.SetLocalTuning(sourceTuning)
}

func isTuningChanged(tuning, fileTuning definition.Tuning) bool {
	if tuning.Value != fileTuning.Value {
		return true
	}

	if tuning.Enabled != fileTuning.Enabled {
		return true
	}

	return false
}

func setDefaultTuning(tuning definition.TuningSpec) {
	definition.SetLocalTuning(definition.Tuning{
		Enabled:     true,
		Name:        tuning.Name,
		Value:       tuning.Limitation.Default,
		Hosts:       []definition.Host{{Name: definition.Hostname, Ip: definition.AdvertiseIp}},
		Description: tuning.Description,
		Limitation:  tuning.Limitation,
		IsModified:  false,
	})
}
