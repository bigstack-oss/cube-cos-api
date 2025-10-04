package cubecos

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/fixpacks"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/time"
	"github.com/google/uuid"
	log "go-micro.dev/v5/logger"
)

func ListFixpacks() ([]fixpacks.Fixpack, error) {
	ctx, canel := context.WithTimeout(wait.CtxSeconds(120))
	defer canel()
	out, err := exec.CommandContext(ctx, "hex_config", "fixpack_get_history").CombinedOutput()
	if err != nil {
		log.Errorf("fixpacks: failed to execute fixpack history cmd(%v)", err)
		return nil, err
	}

	historyFixpacks, err := convertHistoryToFixpacks(out)
	if err != nil {
		log.Errorf("fixpacks: failed to convert fixpacks(%v)", err)
		return nil, err
	}

	pkgFixpacks, err := convertPkgToFixpacks()
	if err != nil {
		log.Errorf("fixpacks: failed to convert pkg to fixpacks(%v)", err)
		return nil, err
	}

	merged := mergeFixpacks(historyFixpacks, pkgFixpacks)
	sortFixpackByVersion(&merged)
	setInstallableStatus(&merged)
	setRemovableStatus(&merged)
	return merged, nil
}

func GetLastFixpackOperation() (*fixpacks.Fixpack, error) {
	fixpacks, err := GetFixpackHistory()
	if err != nil {
		return nil, err
	}

	SortFixpackByTime(&fixpacks)
	return &fixpacks[0], nil
}

func GetFixpackHistory() ([]fixpacks.Fixpack, error) {
	ctx, canel := context.WithTimeout(wait.CtxSeconds(120))
	defer canel()
	out, err := exec.CommandContext(ctx, "hex_config", "fixpack_get_history").CombinedOutput()
	if err != nil {
		log.Errorf("fixpacks: failed to execute fixpack history cmd(%v)", err)
		return nil, err
	}

	historyFixpacks, err := convertHistoryToFixpacks(out)
	if err != nil {
		log.Errorf("fixpacks: failed to convert fixpacks(%v)", err)
		return nil, err
	}

	if len(historyFixpacks) == 0 {
		return nil, fmt.Errorf("no fixpack history found")
	}

	return historyFixpacks, nil
}

func GetFixpackByVersion(version string) (*fixpacks.Fixpack, bool) {
	fixpacks, err := ListFixpacks()
	if err != nil {
		return nil, false
	}

	for _, fixpack := range fixpacks {
		if strings.EqualFold(fixpack.Version, version) {
			return &fixpack, true
		}
	}

	return nil, false
}

func GetFixpackRawByVersion(version string) (*fixpacks.Raw, bool) {
	entries, err := os.ReadDir(fixpacks.UpdateDir)
	if err != nil {
		log.Errorf("fixpack(%s): failed to read update directory %s(%v)", fixpacks.UpdateDir, err)
		return nil, false
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		file := filepath.Join(fixpacks.UpdateDir, entry.Name())
		if !strings.HasSuffix(file, ".fixpack") {
			continue
		}

		info, err := GetFixpackInfo(file)
		if err != nil {
			continue
		}

		if strings.EqualFold(info.Id, version) {
			return info, true
		}
	}

	return nil, false
}

func GetFixpackPathByVersion(version string) (string, bool) {
	entries, err := os.ReadDir(fixpacks.UpdateDir)
	if err != nil {
		log.Errorf("fixpack(%s): failed to read update directory %s(%v)", fixpacks.UpdateDir, err)
		return "", false
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		file := filepath.Join(fixpacks.UpdateDir, entry.Name())
		if !strings.HasSuffix(file, ".fixpack") {
			continue
		}

		info, err := GetFixpackInfo(file)
		if err != nil {
			continue
		}

		if strings.EqualFold(info.Id, version) {
			return file, true
		}
	}

	return "", false
}

func InstallFixpack(req *fixpacks.ReqOpts) error {
	out, err := exec.Command("hex_fixpack_install", "-i", req.Path).CombinedOutput()
	if err != nil {
		err := fmt.Errorf("failed to execute the fixpack installation cmd %s(%v %s)", req.Version, err, string(out))
		log.Errorf("fixpack: %v", err)
		return err
	}

	if !IsHexSuccessful(err) {
		err := fmt.Errorf("failed to install fixpack %s(%s)", req.Version, string(out))
		log.Errorf("fixpack: %v", err)
		return err
	}

	return nil
}

func RollbackFixpack() error {
	out, err := exec.Command("hex_fixpack_install", "-u").CombinedOutput()
	if err != nil {
		err := fmt.Errorf("failed to execute the fixpack rollback cmd(%v %s)", err, string(out))
		log.Errorf("fixpack: %v", err)
		return err
	}

	if !IsHexSuccessful(err) {
		err := fmt.Errorf("failed to rollback fixpack(%s)", string(out))
		log.Errorf("fixpack: %v", err)
		return err
	}

	return nil
}

func GetLatestFixpackInfo() (*fixpacks.Fixpack, error) {
	entries, err := os.ReadDir(fixpacks.RollbackDir)
	if err != nil {
		log.Errorf("fixpack: failed to read rollback directory %s(%v)", fixpacks.RollbackDir, err)
		return nil, err
	}

	versions := []int{}
	dirMap := make(map[int]string)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		m := fixpacks.RollbackFileRegex.FindStringSubmatch(entry.Name())
		if m == nil {
			continue
		}

		version, err := strconv.Atoi(m[1])
		if err != nil {
			log.Warnf("fixpack: failed to parse rollback dir version %s(%v)", entry.Name(), err)
			continue
		}

		versions = append(versions, version)
		dirMap[version] = filepath.Join(
			fixpacks.RollbackDir,
			entry.Name(),
		)
	}

	if len(versions) == 0 {
		err := fmt.Errorf("no fixpack directories found")
		log.Errorf("fixpack: %v", err)
		return nil, err
	}

	sort.Ints(versions)
	latestDir := dirMap[versions[len(versions)-1]]
	infoFile := filepath.Join(latestDir, fixpacks.Info)
	file, err := os.Open(infoFile)
	if err != nil {
		log.Errorf("fixpack: failed to open fixpack info file %s(%v)", infoFile, err)
		return nil, err
	}

	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "FIXPACK_ID=") {
			continue
		}

		segments := strings.Split(line, "=")
		if len(segments) < 2 {
			continue
		}

		version := strings.NewReplacer("\"", "", " ", "").Replace(segments[1])
		fixpack, found := GetFixpackByVersion(version)
		if !found {
			err := fmt.Errorf("fixpack version %s not found", version)
			log.Errorf("fixpack: %v", err)
			return nil, err
		}

		return fixpack, nil
	}

	err = scanner.Err()
	if err != nil {
		log.Errorf("fixpack: failed to scan fixpack info file %s(%v)", infoFile, err)
		return nil, err
	}

	return nil, fmt.Errorf(
		"failed to find fixpack id in info file",
	)
}

func SortFixpackByTime(list *[]fixpacks.Fixpack) {
	sort.Slice(*list, func(i, j int) bool {
		return (*list)[i].UpdatedAt > (*list)[j].UpdatedAt
	})
}

func convertHistoryToFixpacks(out []byte) ([]fixpacks.Fixpack, error) {
	fixpacks := parseHistoryFixpacks(out)
	fixpacks = sortFixpacksByUpdatedAt(fixpacks)
	fixpacks = deduplicateFixpacks(fixpacks)
	return filterOutUninstalledFixpacks(fixpacks), nil
}

func sortFixpacksByUpdatedAt(fixpacks []fixpacks.Fixpack) []fixpacks.Fixpack {
	sort.Slice(fixpacks, func(i, j int) bool {
		return (fixpacks)[i].UpdatedAt > (fixpacks)[j].UpdatedAt
	})

	return fixpacks
}

func filterOutUninstalledFixpacks(list []fixpacks.Fixpack) []fixpacks.Fixpack {
	filtered := make([]fixpacks.Fixpack, 0, len(list))
	for _, fixpack := range list {
		if fixpack.Status.Current == status.Available {
			continue
		}

		filtered = append(filtered, fixpack)
	}

	return filtered
}

func parseHistoryFixpacks(out []byte) []fixpacks.Fixpack {
	list := []fixpacks.Fixpack{}
	lines := strings.SplitSeq(string(out), "\n")
	for line := range lines {
		segments := strings.Split(line, "|")
		if len(segments) < 6 {
			continue
		}

		list = append(list, fixpacks.Fixpack{
			Version:   segments[1],
			Action:    segments[4],
			Note:      segments[5],
			UpdatedAt: convertRawTime(time.FormatFixpack, segments[0]),
			Status:    convertFixpackStatus(segments[3], segments[4]),
		})
	}

	return list
}

func deduplicateFixpacks(list []fixpacks.Fixpack) []fixpacks.Fixpack {
	seen := make(map[string]fixpacks.Fixpack)
	for _, fixpack := range list {
		_, found := seen[fixpack.Version]
		if !found {
			seen[fixpack.Version] = fixpack
		}
	}

	deduplicated := make([]fixpacks.Fixpack, 0, len(seen))
	for _, fixpack := range seen {
		deduplicated = append(deduplicated, fixpack)
	}

	return deduplicated
}

func convertPkgToFixpacks() ([]fixpacks.Fixpack, error) {
	list := []fixpacks.Fixpack{}
	entries, err := os.ReadDir(fixpacks.UpdateDir)
	if err != nil {
		log.Errorf("fixpack(%s): failed to read update directory %s(%v)", fixpacks.UpdateDir, err)
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		file := filepath.Join(fixpacks.UpdateDir, entry.Name())
		if !strings.HasSuffix(file, ".fixpack") {
			continue
		}

		info, err := GetFixpackInfo(file)
		if err != nil {
			continue
		}

		list = append(list, fixpacks.Fixpack{
			Version:        info.Id,
			Name:           info.Name,
			Note:           info.Description,
			Details:        info.Details,
			RebootRequired: parseRebootRequired(info.RebootRequired),
			TargetNodes:    parseRebootTargetNodes(info.RebootRequired),
			Status: status.Fixpack{
				Current:        status.Available,
				IsInstallable:  true,
				IsRollbackable: info.NoRollback,
			},
		})
	}

	return list, nil
}

func parseRebootRequired(list []string) bool {
	return len(list) > 0
}

func parseRebootTargetNodes(roles []string) []string {
	list := []string{}
	for _, role := range roles {
		nodes, err := nodes.GetNodesByRole(strings.ToLower(role))
		if err != nil {
			log.Warnf("fixpack: failed to get nodes by role %s(%v)", role, err)
			continue
		}

		for _, n := range nodes {
			list = append(list, n.Hostname)
		}
	}

	return list
}

func mergeFixpacks(histories, pkgs []fixpacks.Fixpack) []fixpacks.Fixpack {
	merged := make(map[string]fixpacks.Fixpack)
	for _, fixpack := range pkgs {
		merged[fixpack.Version] = fixpack
	}

	for _, history := range histories {
		pkg, found := merged[history.Version]
		if !found {
			continue
		}

		pkg.Status.Current = history.Status.Current
		pkg.Status.IsInstallable = history.Status.IsInstallable
		pkg.Status.IsProcessing = history.Status.IsProcessing
		merged[pkg.Version] = pkg
	}

	fixpacks := make([]fixpacks.Fixpack, 0, len(merged))
	for _, fixpack := range merged {
		fixpacks = append(fixpacks, fixpack)
	}

	return fixpacks
}

func sortFixpackByVersion(list *[]fixpacks.Fixpack) {
	sort.Slice(*list, func(i, j int) bool {
		return (*list)[i].Version > (*list)[j].Version
	})
}

func setInstallableStatus(fixpacks *[]fixpacks.Fixpack) {
	if len(*fixpacks) == 0 {
		return
	}

	for i := len(*fixpacks) - 1; i >= 0; i-- {
		if (*fixpacks)[i].Status.Current == status.Available {
			(*fixpacks)[i].Status.IsInstallable = true
		}

		previous := i + 1
		isLast := previous > len(*fixpacks)-1
		if isLast {
			continue
		}

		previousIsInstalled := (*fixpacks)[previous].Status.Current == status.Installed
		currentIsAvailable := (*fixpacks)[i].Status.Current == status.Available
		if previousIsInstalled && currentIsAvailable {
			(*fixpacks)[i].Status.IsInstallable = true
		}
	}
}

func setRemovableStatus(fixpacks *[]fixpacks.Fixpack) {
	if len(*fixpacks) == 0 {
		return
	}

	for i, fixpack := range *fixpacks {
		if fixpack.Status.Current == status.Available {
			(*fixpacks)[i].Status.IsRemovable = true
		}

		if i != 0 {
			(*fixpacks)[i].Status.IsRemovable = false
		}
	}
}

func convertFixpackStatus(rollback, action string) status.Fixpack {
	s := status.Fixpack{}
	if strings.EqualFold(rollback, "yes") {
		s.IsRollbackable = true
	}

	if strings.EqualFold(action, "installed") {
		s.Current = status.Installed
	}

	if strings.EqualFold(action, "uninstalled") {
		s.IsInstallable = true
		s.Current = status.Available
	}

	if isUnknownAction(action) {
		s.Current = status.Failed
	}

	return s
}

func isUnknownAction(action string) bool {
	return !strings.EqualFold(action, "installed") &&
		!strings.EqualFold(action, "uninstalled")
}

func GetFixpackInfo(file string) (*fixpacks.Raw, error) {
	info, err := parseFixpackInfo(file)
	if err != nil {
		return nil, err
	}

	raw := &fixpacks.Raw{Details: string(info)}
	lines := strings.SplitSeq(string(info), "\n")
	for line := range lines {
		segment := strings.Split(line, "=")
		if len(segment) < 2 {
			continue
		}

		key := segment[0]
		val := strings.ReplaceAll(segment[1], "\"", "")
		switch key {
		case "FIXPACK_ID":
			raw.Id = val
		case "FIXPACK_NAME":
			raw.Name = val
		case "SUPPORTED_FIRMWARES":
			raw.SupportedFirmwares = strings.Split(val, " ")
		case "FIXPACK_DESCRIPTION":
			raw.Description = val
		case "REBOOT_REQUIRED":
			raw.RebootRequired = parseRebootRequiredRoles(val)
		case "NOROLLBACK":
			raw.NoRollback = parseNoRollback(val)
		}
	}

	return raw, nil
}

func parseRebootRequiredRoles(val string) []string {
	roles := []string{}
	for _, str := range strings.Split(val, ",") {
		role := strings.ToLower(strings.TrimSpace(str))
		if nodes.HasRole(role) {
			roles = append(roles, role)
		}
	}

	return roles
}

func parseNoRollback(val string) bool {
	intVal, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		log.Warnf("fixpack: failed to parse no rollback value %s(%v)", val, err)
		return false
	}

	return intVal == 0
}

func genTmpFixpackDir() (string, error) {
	hash := uuid.New().String()[:8]
	dir := fmt.Sprintf("/tmp/fixpack-%s", hash)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		log.Errorf("fixpack: failed to create tmp fixpack dir %s(%v)", dir, err)
		return "", err
	}

	return dir, nil
}

func parseFixpackInfo(file string) ([]byte, error) {
	tmpDir, err := genTmpFixpackDir()
	if err != nil {
		return nil, err
	}

	defer unmountTmpDir(tmpDir)
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(30))
	defer cancel()

	out, err := exec.CommandContext(ctx, "mount", file, tmpDir).CombinedOutput()
	if err != nil {
		err := fmt.Errorf("failed to mount fixpack %s(%v %s)", file, err, string(out))
		log.Errorf("fixpack: %v", err)
		return nil, err
	}

	infoPath := filepath.Join(tmpDir, "fixpack.info")
	bytes, err := os.ReadFile(infoPath)
	if err != nil {
		err := fmt.Errorf("failed to read fixpack info(%v)", err)
		log.Errorf("fixpack: %v", err)
		return nil, err
	}

	return bytes, nil
}

func unmountTmpDir(tmpDir string) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(30))
	defer cancel()

	out, err := exec.CommandContext(ctx, "umount", tmpDir).CombinedOutput()
	if err != nil {
		log.Errorf("fixpack: failed to unmount tmp dir %s(%v %s)", tmpDir, err, string(out))
	}

	err = os.RemoveAll(tmpDir)
	if err != nil {
		log.Errorf("fixpack: failed to remove tmp dir %s(%v)", tmpDir, err)
	}
}
