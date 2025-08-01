package cubecos

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"strconv"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/images"
	log "go-micro.dev/v5/logger"
)

func GetReservedImages() []images.ReqOpts {
	return []images.ReqOpts{
		{
			File:                        "amphora-x64-haproxy-yoga.qcow2",
			Name:                        "amphora-x64-haproxy",
			Os:                          "Ubuntu",
			Destination:                 "CubeStorage",
			Domain:                      "default",
			Project:                     "admin",
			SourceFromAnotherHypervisor: false,
			Visibility:                  "private",
		},
		{
			File:                        "manila-service-image_yoga.qcow2",
			Name:                        "manila-service-image",
			Os:                          "Ubuntu",
			Destination:                 "CubeStorage",
			Domain:                      "default",
			Project:                     "admin",
			SourceFromAnotherHypervisor: false,
			Visibility:                  "private",
		},
	}
}

func ImportImage(opts *images.CreateOpts) error {
	ctx, cancel := context.WithTimeout(wait.CtxMinutes(180))
	defer cancel()
	cmd := exec.CommandContext(
		ctx, "hex_sdk", "os_image_import_with_attrs",
		opts.AttributesType, opts.Dir, opts.File, opts.Name,
		opts.Domain, opts.Project, opts.PoolType, opts.Visibility,
	)

	stdout, err := cmd.StderrPipe()
	cmd.Stdout = cmd.Stderr
	if err != nil {
		log.Errorf("images: failed to get stdout pipe(%v)", err)
		return err
	}

	err = cmd.Start()
	if err != nil {
		log.Errorf("images: failed to start command(%v)", err)
		return err
	}

	traceImportProgress(opts, stdout)
	err = cmd.Wait()
	if err != nil {
		log.Errorf("images: failed to wait for image %s import command(%v)", opts.Name, err)
		return err
	}

	if !IsHexSdkSuccess(err) {
		err := fmt.Errorf("failed to import image %s(%v)", opts.Name, err)
		return err
	}

	return nil
}

func traceImportProgress(opts *images.CreateOpts, stdout io.Reader) {
	reader := bufio.NewReader(stdout)
	buffer := bytes.Buffer{}
	if opts.StreamingLogs != nil {
		defer close(opts.StreamingLogs)
	}

	for {
		bytes, err := reader.ReadByte()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Errorf("images: failed to read byte from image %s import stdout(%v)", opts.Name, err)
			return
		}

		if bytes != '\r' {
			buffer.WriteByte(bytes)
			continue
		}

		line := buffer.String()
		buffer.Reset()
		regex := regexp.MustCompile(`\(([\d.]+)\/100%`)
		match := regex.FindStringSubmatch(line)
		if len(match) <= 1 {
			continue
		}

		if opts.StreamingLogs != nil {
			progress, err := strconv.ParseFloat(match[1], 64)
			if err == nil {
				opts.StreamingLogs <- progress
			}
		}
	}
}
