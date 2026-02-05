package cubecos

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	ostime "time"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/http"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/ssh"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/bodies"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/firmwares"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	defssh "github.com/bigstack-oss/cube-cos-api/internal/definition/v1/ssh"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/time"
	json "github.com/json-iterator/go"
	log "go-micro.dev/v5/logger"
	cryptossh "golang.org/x/crypto/ssh"
	"gopkg.in/yaml.v3"
)

func ListFirmwares() ([]firmwares.Firmware, error) {
	update, err := parseUpdateHistory()
	if err != nil {
		return nil, err
	}

	firmwares := convertHistoryToFirmwares(update)
	appendUninstalledFirmwares(&firmwares)
	return firmwares, nil
}

// note:
// please DO NOT use exec.CommandContext with timeout for hex_install
// because the duration of firmware upgrade is not predictable from CubeCOS, and it may take a long time to complete.
// use timeout might makes the situation to be worse.
func UpgradeFirmware(req *firmwares.ReqOpts) error {
	out, err := exec.Command("hex_install", "-v", "update", req.PkgPath).CombinedOutput()
	if err != nil {
		errDesc := strings.ReplaceAll(string(out), "\n", " ")
		log.Errorf("firmwares: failed to execute firmware upgrade %s(%s %s)", req.Version, err, errDesc)
		code, stderr := getUpdateFirmwareStatus()
		return fmt.Errorf("UPG200%d: %s", code, stderr)
	}

	if !IsHexSuccessful(err) {
		err := fmt.Errorf("%v %s", err, string(out))
		log.Errorf("firmwares: failed to upgrade firmware(%v)", err)
		code, stderr := getUpdateFirmwareStatus()
		return fmt.Errorf("UPG200%d: %s", code, stderr)
	}

	log.Infof("firmwares: %s", string(out))
	return nil
}

func getUpdateFirmwareStatus() (int, string) {
	out, err := exec.Command("hex_sdk", "-v", "stats_partition").CombinedOutput()
	if err != nil {
		intgErr := genIntegrationErr("firmware fetch status exec failure")
		log.Errorf("firmwares: %s (%s)", intgErr.Error(), string(out))
		return GetCmdReturnCode(err), string(out)
	}

	if !IsHexSuccessful(err) {
		intgErr := genIntegrationErr("firmware fetch status output failure")
		log.Errorf("firmwares: %s (%s)", intgErr.Error(), string(out))
		return GetCmdReturnCode(err), string(out)
	}

	return 0, string(out)
}

func GetUpdateInterruptedNode() (*nodes.Node, error) {
	return nil, fmt.Errorf("waiting COS to provide the SDK, so not implemented yet")
}

func GetBootstrappingProgress() ([]firmwares.BootstrappingStatus, error) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(180))
	defer cancel()
	out, _ := exec.CommandContext(ctx, "hex_sdk", "-f", "json", "stats_bootstrap").Output()

	var status []firmwares.BootstrappingStatus
	err := json.Unmarshal(out, &status)
	if err != nil {
		err := genIntegrationErr("bootstrapping progress output parsing failure")
		log.Errorf("firmwares: %s (%s)", err.Error(), string(out))
		return nil, err
	}

	return status, nil
}

func GetUpgradeProgress() (*firmwares.Upgrade, error) {
	out, err := os.ReadFile(firmwares.UpdateProgress)
	if err != nil {
		log.Errorf("firmwares: failed to read progress file(%v)", err)
		return nil, err
	}

	upgrade := &firmwares.Upgrade{}
	err = json.Unmarshal(out, upgrade)
	if err != nil {
		log.Errorf("firmwares: failed to unmarshal progress file(%v)", err)
		return nil, err
	}

	return upgrade, nil
}

func GetUpgradeProgressFromVip() (*firmwares.Upgrade, error) {
	node, err := GetVirtualIpController()
	if err != nil {
		log.Errorf("firmwares: failed to get virtual IP controller(%v)", err)
		return nil, err
	}

	http := http.GetGlobalHelper()
	resp, err := http.R().
		SetResult(&bodies.FirmwareUpgradeProgress{}).
		SetHeaders(nodes.GetSecretHeaders()).
		Get(node.GetFirmwareUpgradeProgressUrl())
	if err != nil {
		log.Errorf("firmwares: unable to get firmware upgrade progress from node %s (%v)", node.Hostname, err)
		return nil, err
	}

	if resp.IsError() {
		err := fmt.Errorf("HTTP %d: %s", resp.StatusCode(), string(resp.Body()))
		log.Errorf("firmwares: failed to get firmware upgrade progress from node %s (%v)", node.Hostname, err)
		return nil, err
	}

	return &resp.Result().(*bodies.FirmwareUpgradeProgress).Data, nil
}

func SetNodeUpdateProgress(hostname, phase, status string) error {
	update, err := GetUpgradeProgressFromVip()
	if err != nil {
		log.Errorf("firmwares: failed to get update progress(%v)", err)
		return err
	}

	for i, progress := range update.Progresses {
		if progress.Host != hostname {
			continue
		}

		update.Progresses[i].Phase = phase
		update.Progresses[i].Status.Current = status
	}

	return SetProgressDetails(update)
}

func SetNodeAsContinueAnywaied(hostname string) error {
	update, err := GetUpgradeProgress()
	if err != nil {
		log.Errorf("firmwares: failed to get update progress(%v)", err)
		return err
	}

	for i, progress := range update.Progresses {
		if progress.Host == hostname {
			update.Progresses[i].Status.IsContinueAnywaied = true
			break
		}
	}

	return SetProgressDetails(update)
}

func SetProgressDetails(progress *firmwares.Upgrade) error {
	file, err := os.Create(firmwares.UpdateProgress)
	if err != nil {
		log.Errorf("firmwares: failed to create progress file for update(%v)", err)
		return err
	}

	defer file.Close()
	content, err := json.MarshalIndent(progress, "", "  ")
	if err != nil {
		log.Errorf("firmwares: failed to marshal progress details(%v)", err)
		return err
	}

	_, err = file.WriteString(string(content))
	if err != nil {
		log.Errorf("firmwares: failed to write progress file(%v)", err)
		return err
	}

	return nil
}

func SyncFirmwareUpgradeProgressToAllNodes() {
	for _, node := range nodes.List() {
		if node.IsLocal() {
			continue
		}

		err := MoveFirmwareUpgradeProgress(node.Hostname)
		if err != nil {
			log.Errorf("nodes: failed to move firmware upgrade progress to controller %s(%v)", node.Hostname, err)
		}
	}
}

func MoveFirmwareUpgradeProgress(node string) error {
	if !IsProgressFileExist() {
		return errors.New("firmware upgrade progress file does not exist")
	}

	sshAuth, err := defssh.GenSshAuth(defssh.DefaultPrivateKey)
	if err != nil {
		return err
	}

	ssh, err := ssh.NewHelper(
		ssh.Host(fmt.Sprintf("%s:22", node)),
		ssh.User("root"),
		ssh.AuthMethod(sshAuth),
		ssh.HostKeyCallback(cryptossh.InsecureIgnoreHostKey()),
	)
	if err != nil {
		return err
	}

	defer ssh.Close()
	err = ssh.Copy(firmwares.UpdateProgress, firmwares.UpdateProgress)
	if err != nil {
		log.Errorf("firmwares: failed to copy firmware upgrade progress to node %s(%v)", node, err)
		return err
	}

	return nil
}

func IsProgressFileExist() bool {
	_, err := os.Stat(firmwares.UpdateProgress)
	if err == nil {
		return true
	}

	if os.IsNotExist(err) {
		return false
	}

	log.Errorf(
		"firmwares: failed to check if firmware upgrade progress file exists(%v)",
		err,
	)

	return false
}

func parseUpdateHistory() (*firmwares.Upadte, error) {
	data, err := os.ReadFile(firmwares.UpdateHistory)
	if err != nil {
		log.Errorf("firmwares: failed to read update history file %s (%v)", firmwares.UpdateHistory, err)
		return nil, err
	}

	update := &firmwares.Upadte{}
	err = yaml.Unmarshal(data, update)
	if err != nil {
		log.Errorf("firmwares: failed to unmarshal update history file %s (%v)", firmwares.UpdateHistory, err)
		return nil, err
	}

	return update, nil
}

func convertHistoryToFirmwares(update *firmwares.Upadte) []firmwares.Firmware {
	firmwaresList := make([]firmwares.Firmware, 0, len(update.History))

	for _, raw := range update.History {
		date := convertRawTime(time.FormatFirmware, raw.CreatedAt)
		dayBaseDate := convertRawTimeToDayBaseDate(raw.BuiltAt)
		firmwaresList = append(firmwaresList, firmwares.Firmware{
			Version:      convertFirmwareVersion(raw.Version, dayBaseDate),
			ReleaseNotes: convertReleaseNotes(raw.Version, raw.Variant, dayBaseDate),
			UpdatedAt:    date,
			Status: status.Firmware{
				Current:     status.Succeeded,
				IsRemovable: false,
			},
		})
	}

	return firmwaresList
}

func convertRawTimeToDayBaseDate(rawTime string) string {
	segments := strings.Split(rawTime, " ")
	if len(segments) < 2 {
		return ""
	}

	return segments[0]
}

func convertRfc3339ToDayBaseDate(rfc3339Time string) string {
	t, err := ostime.Parse(ostime.RFC3339, rfc3339Time)
	if err != nil {
		panic(err)
	}

	return t.Format("20060102-1504")
}

func appendUninstalledFirmwares(list *[]firmwares.Firmware) {
	isInstallted := map[string]bool{}
	for _, firmware := range *list {
		isInstallted[firmware.Version] = true
	}

	entries, err := os.ReadDir(firmwares.UpdateDir)
	if err != nil {
		log.Errorf("firmwares: failed to read update directory %s (%v)", firmwares.UpdateDir, err)
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if !strings.HasSuffix(entry.Name(), ".pkg") {
			continue
		}

		firmware, err := ConvertPkgNameToFirmware(entry.Name())
		if err != nil {
			continue
		}

		if !isInstallted[firmware.Version] {
			(*list) = append(*list, *firmware)
		}
	}
}

func convertRawTime(layout, rawTime string) string {
	t, err := ostime.ParseInLocation(layout, rawTime, time.LocalFixedZone)
	if err != nil {
		log.Errorf("firmwares: failed to parse time %s (%v)", rawTime, err)
		return ""
	}

	return time.RFC3339Z(t)
}

func ConvertPkgNameToFirmware(pkgname string) (*firmwares.Firmware, error) {
	pkgname = strings.TrimSuffix(pkgname, ".pkg")
	segment := strings.Split(pkgname, "_")
	if len(segment) < 3 {
		err := fmt.Errorf("invalid firmware package name: %s", pkgname)
		log.Errorf("firmwares: %v", err)
		return nil, err
	}

	date := convertRawTime(time.FormatFirmwarePkg, segment[2])
	return &firmwares.Firmware{
		Version:      convertFirmwareVersion(segment[1], segment[2]),
		ReleaseNotes: convertReleaseNotes(segment[1], segment[3], convertRfc3339ToDayBaseDate(date)),
		Status: status.Firmware{
			Current:     status.Available,
			IsUpdatable: true,
			IsRemovable: true,
		},
	}, nil
}

func convertFirmwareVersion(version, date string) string {
	return fmt.Sprintf("Cube Appliance %s %s", version, date)
}

func convertReleaseNotes(version, variant, date string) string {
	return fmt.Sprintf("The CubeCOS %s(%s) firmware release since %s", version, variant, date)
}
