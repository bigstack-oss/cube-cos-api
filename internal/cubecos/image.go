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
	"strings"

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
)

func ImportImage(opts *images.CreateOpts) error {
	ctx, cancel := context.WithTimeout(wait.CtxMinutes(180))
	defer cancel()
	cmd, err := genCmdByPoolType(ctx, opts)
	if err != nil {
		return err
	}

	defer removeWorkaroundScriptIfExists()
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

func genCmdByPoolType(ctx context.Context, opts *images.CreateOpts) (*exec.Cmd, error) {
	switch opts.PoolType {
	case "cinder-volumes":
		return genCinderWorkaroundCmd(ctx, opts)
	default:
		return genGlanceCmd(ctx, opts)
	}
}

func genCinderWorkaroundCmd(ctx context.Context, opts *images.CreateOpts) (*exec.Cmd, error) {
	script := fmt.Sprintf(
		"#!/bin/bash\nhex_sdk %s\n",
		strings.Join(genImageArgs(opts), " "),
	)

	scriptPath := "image-import-cinder.sh"
	if err := os.WriteFile(scriptPath, []byte(script), 0755); err != nil {
		log.Errorf("images: failed to write script file(%v)", err)
		return nil, err
	}

	absPath, err := filepath.Abs(scriptPath)
	if err != nil {
		log.Errorf("images: failed to get abs path(%v)", err)
		return nil, err
	}

	err = os.Chmod(absPath, 0755)
	if err != nil {
		log.Errorf("images: failed to chmod script file(%v)", err)
		return nil, err
	}

	return exec.CommandContext(
		ctx,
		absPath,
	), nil
}

func genGlanceCmd(ctx context.Context, opts *images.CreateOpts) (*exec.Cmd, error) {
	return exec.CommandContext(
		ctx,
		"hex_sdk",
		genImageArgs(opts)...,
	), nil
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
			fmt.Sprintf(`"--os-project-domain-name %s --os-project-name %s --visibility %s"`, opts.Domain, opts.Project, opts.Visibility),
			opts.PoolType, opts.Destination,
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

func removeWorkaroundScriptIfExists() {
	os.Remove("/image-import-cinder.sh")
}
