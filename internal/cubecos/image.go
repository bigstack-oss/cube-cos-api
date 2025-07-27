package cubecos

import (
	"context"
	"fmt"
	"os/exec"

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

func ImportImage(opts images.CreateOpts) error {
	ctx, cancel := context.WithTimeout(wait.CtxMinutes(180))
	defer cancel()

	out, err := exec.CommandContext(
		ctx, "hex_sdk", "os_image_import_with_attrs",
		opts.AttributesType, opts.Dir, opts.File, opts.Name,
		opts.Domain, opts.Project, opts.PoolType, opts.Visibility,
	).CombinedOutput()
	if err != nil {
		err := fmt.Errorf("failed to execute image import cmd %s(%v %s)", opts.Name, err, string(out))
		log.Errorf("images: %v", err)
		return err
	}

	if !IsHexSdkSuccess(err) {
		err := fmt.Errorf("failed to import image %s(%v %s)", opts.Name, err, string(out))
		return err
	}

	return nil
}
