package cubecos

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/fixpacks"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/time"
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

	fixpacks, err := convertToFixpacks(out)
	if err != nil {
		log.Errorf("fixpacks: failed to convert fixpacks(%v)", err)
		return nil, err
	}

	err = addUninstalledFixpacks(&fixpacks)
	if err != nil {
		log.Errorf("fixpacks: failed to add uninstalled fixpacks(%v)", err)
		return nil, err
	}

	return fixpacks, nil
}

func convertToFixpacks(out []byte) ([]fixpacks.Fixpack, error) {
	fixpacksList := []fixpacks.Fixpack{}
	lines := strings.SplitSeq(string(out), "\n")
	for line := range lines {
		segments := strings.Split(line, "|")
		if len(segments) < 6 {
			log.Warnf("fixpacks: invalid fixpack line(%s)", line)
			continue
		}

		fixpack := fixpacks.Fixpack{
			Version:   segments[1],
			Note:      segments[5],
			UpdatedAt: convertRawTime(time.FormatFixpack, segments[0]),
			Status:    convertFixpackStatus(segments[3], segments[4]),
		}

		fixpacksList = append(fixpacksList, fixpack)
	}

	return fixpacksList, nil
}

func addUninstalledFixpacks(list *[]fixpacks.Fixpack) error {
	entries, err := os.ReadDir(fixpacks.UpdateDir)
	if err != nil {
		log.Errorf("fixpack(%s): failed to read update directory %s(%v)", fixpacks.UpdateDir, err)
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		file := filepath.Join(fixpacks.UpdateDir, entry.Name())
		if !strings.HasSuffix(file, ".fixpack") {
			continue
		}

		*list = append(*list, fixpacks.Fixpack{
			Version: entry.Name(),
			Note:    "Uninstalled fixpack",
			Status: status.Fixpack{
				Current:       status.Available,
				IsInstallable: true,
			},
		})
	}

	return nil
}

func convertFixpackStatus(rollback, action string) status.Fixpack {
	status := status.Fixpack{Current: "installed"}

	if strings.EqualFold(rollback, "yes") {
		status.IsRollbackable = true
	}

	if strings.EqualFold(action, "installed") {
		status.IsInstallable = false
	}

	return status
}
