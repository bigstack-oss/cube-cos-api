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
		err := genIntegrationErr("storage exec failure")
		log.Errorf("storage: %s (%s)", err.Error(), string(out))
		return nil, err
	}

	if !IsHexSuccessful(err) {
		err := genIntegrationErr("storage output failure")
		log.Errorf("storage: %s (%s)", err.Error(), string(out))
		return nil, err
	}

	list := []storages.Cinder{}
	err = json.Unmarshal(out, &list)
	if err != nil {
		err := genIntegrationErr("storage parsing failure")
		log.Errorf("storage: %s (%s)", err.Error(), string(out))
		return nil, err
	}

	return list, nil
}

func CheckStorageExist(name string) (bool, error) {
	storages, err := ListStorages()
	if err != nil {
		return false, err
	}

	for _, storage := range storages {
		if storage.Name == name {
			return true, nil
		}
	}

	return false, nil
}

func GetStorage(name string) (*storages.Cinder, error) {
	storages, err := ListStorages()
	if err != nil {
		return nil, err
	}

	for _, storage := range storages {
		if storage.Name == name {
			return &storage, nil
		}
	}

	return nil, fmt.Errorf(
		"storage %s not found", name,
	)
}

func ListModels() ([]storages.Model, error) {
	ctx, cancel := context.WithTimeout(wait.CtxMinutes(3))
	defer cancel()
	out, err := exec.CommandContext(ctx, "hex_sdk", "host_get_models").CombinedOutput()
	if err != nil {
		err := genIntegrationErr("model exec failure")
		log.Errorf("storage: %s (%s)", err.Error(), string(out))
		return nil, err
	}

	if !IsHexSuccessful(err) {
		err := genIntegrationErr("model output failure")
		log.Errorf("storage: %s (%s)", err.Error(), string(out))
		return nil, err
	}

	list := []storages.Model{}
	err = json.Unmarshal(out, &list)
	if err != nil {
		err := genIntegrationErr("model parsing failure")
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
