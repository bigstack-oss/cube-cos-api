package cubecos

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/openstack/v2"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/images"
	tty "github.com/creack/pty"
	opsimage "github.com/gophercloud/gophercloud/v2/openstack/image/v2/images"
	log "go-micro.dev/v5/logger"
)

var (
	imageProgress  = regexp.MustCompile(`\[[= >]+\]\s+(\d+)%`)
	volumeProgress = regexp.MustCompile(`Importing image:\s+(\d+)%\s+complete`)
	importParam    = regexp.MustCompile(`^[\p{L}\p{N}][\p{L}\p{N}._+-]*$`)
)

const maxImportParamLen = 255

// ValidateImportOpts limits every import parameter to a plain token, since
// hex_sdk passes them on to further shell commands. The file must also be a
// bare name so it stays inside the glance directory.
func ValidateImportOpts(opts images.ReqOpts) error {
	params := []struct{ field, value string }{
		{"file", opts.File},
		{"name", opts.Name},
		{"os", opts.Os},
		{"destination", opts.Destination},
		{"domain", opts.Domain},
		{"project", opts.Project},
		{"visibility", opts.Visibility},
	}

	for _, p := range params {
		if len(p.value) > maxImportParamLen {
			return fmt.Errorf("invalid %s: longer than %d bytes", p.field, maxImportParamLen)
		}

		if !importParam.MatchString(p.value) {
			return fmt.Errorf(
				"invalid %s %q: only letters, digits, '.', '_', '+' and '-' are allowed, starting with a letter or digit",
				p.field, p.value,
			)
		}
	}

	if filepath.Base(opts.File) != opts.File {
		return fmt.Errorf("invalid file %q: must be a file name, not a path", opts.File)
	}

	return nil
}

func ImportImage(opts *images.CreateOpts) error {
	ctx, cancel := context.WithTimeout(wait.CtxMinutes(180))
	defer cancel()
	cmd := genCmd(ctx, opts)
	out, err := tty.Start(cmd)
	if err != nil {
		log.Errorf("images: failed to start command(%v)", err)
		return err
	}

	defer out.Close()
	traceImportProgress(opts, out)
	err = cmd.Wait()
	if err != nil {
		log.Errorf("images: failed to wait for image %s import command(%v)", opts.Name, err)
		return err
	}

	if !IsHexSuccessful(err) {
		err := fmt.Errorf("failed to import image %s(%v)", opts.Name, err)
		log.Errorf("images: %v", err)
		return err
	}

	return nil
}

func RemoveRawImage(path string) {
	err := os.Remove(path)
	if err == nil {
		return
	}

	if !os.IsNotExist(err) {
		log.Warnf("images: failed to remove raw image file %s(%v)", path, err)
	}
}

func genCmd(ctx context.Context, opts *images.CreateOpts) *exec.Cmd {
	return exec.CommandContext(ctx, "hex_sdk", genImageArgs(opts)...)
}

func genImageArgs(opts *images.CreateOpts) []string {
	switch opts.ReservedType {
	case "lb":
		return []string{
			"os_octavia_image_import",
			opts.Dir, opts.File,
		}
	case "fs":
		return []string{
			"os_manila_image_import",
			opts.Dir, opts.File,
		}
	default:
		return []string{
			"os_image_import",
			opts.Dir, opts.File, opts.Name,
			fmt.Sprintf("--os-project-domain-name %s --os-project-name %s --visibility %s", opts.Domain, opts.Project, opts.Visibility),
			opts.PoolType, opts.Destination, opts.Os,
		}
	}
}

func SetImagePropertiesByName(name string, opts opsimage.UpdateOpts) error {
	h := openstack.GetGlobalHelper()
	image, err := h.GetImageByName(name)
	if err != nil {
		log.Errorf("images: failed to get image %s(%v)", name, err)
		return err
	}

	return h.UpdateImageProperty(image.ID, opts)
}

func UpdateImage(id string, opts opsimage.UpdateOpts) error {
	h := openstack.GetGlobalHelper()
	return h.UpdateImageProperty(id, opts)
}

func traceImportProgress(opts *images.CreateOpts, out io.Reader) {
	buf := bytes.Buffer{}
	last := float64(0)

	for {
		tmp := make([]byte, 1)
		n, err := out.Read(tmp)
		if err != nil || n == 0 {
			break
		}

		bytes := tmp[0]
		if bytes == '\n' {
			buf.Reset()
			continue
		}

		if bytes == '\r' {
			streamImportProgress(
				opts.PoolType, &buf, &last, &opts.StreamingLogs,
			)
			buf.Reset()
			continue
		}

		buf.WriteByte(bytes)
	}
}

func streamImportProgress(poolType string, buf *bytes.Buffer, last *float64, streamingLogs *chan float64) {
	if streamingLogs == nil {
		return
	}

	line := buf.String()
	if len(line) <= 0 {
		return
	}

	matches := []string{}
	switch poolType {
	case "glance-images":
		matches = imageProgress.FindStringSubmatch(line)
	case "cinder-volumes":
		matches = volumeProgress.FindStringSubmatch(line)
	}
	if len(matches) < 2 {
		return
	}

	percent, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return
	}

	if percent != *last {
		*last = percent
		*streamingLogs <- percent
	}
}
