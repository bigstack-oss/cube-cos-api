package cubecos

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/datacenter"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/integration"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/storages"
	json "github.com/json-iterator/go"
	log "go-micro.dev/v5/logger"
)

func ListBuiltInApplications() []integration.Service {
	if !datacenter.IsCloudType() {
		return integration.GetCommonServices()
	}

	return append(
		integration.GetCommonServices(),
		integration.GetCloudService(),
	)
}

func ListStorages() ([]storages.Cinder, error) {
	ctx, cancel := context.WithTimeout(wait.CtxMinutes(3))
	defer cancel()
	out, err := exec.CommandContext(ctx, "hex_sdk", "cinder_get_storages").CombinedOutput()
	if err != nil {
		err := genIntegrationErr("function exec failure")
		log.Errorf("storage: %s (%s)", err.Error(), string(out))
		return nil, err
	}

	if !IsHexSdkSuccess(err) {
		err := genIntegrationErr("function output failure")
		log.Errorf("storage: %s (%s)", err.Error(), string(out))
		return nil, err
	}

	list := []storages.Cinder{}
	err = json.Unmarshal(out, &list)
	if err != nil {
		err := genIntegrationErr("function parsing failure")
		log.Errorf("storage: %s (%s)", err.Error(), string(out))
		return nil, err
	}

	return list, nil
}

func genIntegrationErr(description string) error {
	return fmt.Errorf(
		"cubecos has unexpected hex error, please contact support(%s)",
		description,
	)
}
