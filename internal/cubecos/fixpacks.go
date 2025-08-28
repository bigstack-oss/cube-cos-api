package cubecos

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/fixpacks"
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

	return mergeFixpacks(
		historyFixpacks,
		pkgFixpacks,
	), nil
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

		info, err := getFixpackInfo(file)
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

		info, err := getFixpackInfo(file)
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
	out, err := exec.Command("hex_config", "fixpack", req.Path).CombinedOutput()
	if err != nil {
		err := fmt.Errorf("failed to execute the fixpack installation cmd %s(%v %s)", req.Version, err, string(out))
		log.Errorf("fixpack: %v", err)
		return err
	}

	if !IsHexSdkSuccess(err) {
		err := fmt.Errorf("failed to install fixpack %s(%s)", req.Version, string(out))
		log.Errorf("fixpack: %v", err)
		return err
	}

	return nil
}

func RollbackFixpack() error {
	out, err := exec.Command("hex_config", "fixpack_rollback").CombinedOutput()
	if err != nil {
		err := fmt.Errorf("failed to execute the fixpack rollback cmd(%v %s)", err, string(out))
		log.Errorf("fixpack: %v", err)
		return err
	}

	if !IsHexSdkSuccess(err) {
		err := fmt.Errorf("failed to rollback fixpack(%s)", string(out))
		log.Errorf("fixpack: %v", err)
		return err
	}

	return nil
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

		info, err := getFixpackInfo(file)
		if err != nil {
			continue
		}

		list = append(list, fixpacks.Fixpack{
			Version: info.Id,
			Name:    info.Name,
			Note:    info.Description,
			Details: info.Details,
			Status: status.Fixpack{
				Current:        status.Available,
				IsInstallable:  true,
				IsRollbackable: true,
			},
		})
	}

	return list, nil
}

func mergeFixpacks(history, pkgs []fixpacks.Fixpack) []fixpacks.Fixpack {
	merged := make(map[string]fixpacks.Fixpack)
	for _, fixpack := range history {
		merged[fixpack.Version] = fixpack
	}

	for _, pkg := range pkgs {
		history, found := merged[pkg.Version]
		if !found {
			merged[pkg.Version] = pkg
			continue
		}

		history.Name = pkg.Name
		history.Note = pkg.Note
		history.Details = pkg.Details
		merged[pkg.Version] = history
	}

	fixpacks := make([]fixpacks.Fixpack, 0, len(merged))
	for _, fixpack := range merged {
		fixpacks = append(fixpacks, fixpack)
	}

	return fixpacks
}

func convertFixpackStatus(rollback, action string) status.Fixpack {
	s := status.Fixpack{Current: "installed"}
	if strings.EqualFold(rollback, "yes") {
		s.IsRollbackable = true
	}

	if strings.EqualFold(action, "uninstalled") {
		s.IsInstallable = true
		s.Current = status.Available
	}

	return s
}

func getFixpackInfo(file string) (*fixpacks.Raw, error) {
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
			raw.SupportedFirmwares = strings.Split(val, ",")
		case "FIXPACK_DESCRIPTION":
			raw.Description = val
		}
	}

	return raw, nil
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
