package cubecos

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	ostime "time"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/base"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/datacenter"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/integration"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/storages"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/time"
	json "github.com/json-iterator/go"
	log "go-micro.dev/v5/logger"
)

func SetDefaultStorage(name string) error {
	nameMap := map[string]string{"name": name}
	input, err := json.Marshal(nameMap)
	if err != nil {
		err := genIntegrationErr("set default storage req parsing failure")
		log.Errorf("storage: %s (%v)", err.Error(), err)
		return err
	}

	ctx, cancel := context.WithTimeout(wait.CtxMinutes(3))
	defer cancel()
	out, err := exec.CommandContext(ctx, "hex_sdk", "cinder_set_default_storage", string(input)).CombinedOutput()
	if err != nil {
		err := genIntegrationErr("set default storage exec failure")
		log.Errorf("storage: %s (%s)", err.Error(), string(out))
		return err
	}

	if !IsHexSuccessful(err) {
		err := genIntegrationErr("set default storage output failure")
		log.Errorf("storage: %s (%s)", err.Error(), string(out))
		return err
	}

	return nil
}

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

	convertStorageTimes(&list)
	return list, nil
}

func GetStorage(name string) (*storages.CinderDetails, error) {
	nameMap := map[string]string{"name": name}
	input, err := json.Marshal(nameMap)
	if err != nil {
		err := genIntegrationErr("storage name parsing failure")
		log.Errorf("storage: %s (%v)", err.Error(), err)
		return nil, err
	}

	ctx, cancel := context.WithTimeout(wait.CtxMinutes(3))
	defer cancel()
	out, err := exec.CommandContext(ctx, "hex_sdk", "cinder_get_storage", string(input)).CombinedOutput()
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

	cinder := &storages.CinderDetails{}
	err = json.Unmarshal(out, cinder)
	if err != nil {
		err := genIntegrationErr("storage parsing failure")
		log.Errorf("storage: %s (%s)", err.Error(), string(out))
		return nil, err
	}

	return cinder, nil
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

func UpdateStorage(details storages.CinderDetails) error {
	input, err := json.Marshal(details)
	if err != nil {
		err := genIntegrationErr("storage cinder details parsing failure")
		log.Errorf("storage: %s (%v)", err.Error(), err)
		return err
	}

	ctx, cancel := context.WithTimeout(wait.CtxMinutes(3))
	defer cancel()
	out, err := exec.CommandContext(ctx, "hex_sdk", "cinder_put_storage", string(input)).CombinedOutput()
	if err != nil {
		err := genIntegrationErr("storage exec failure")
		log.Errorf("storage: %s (%s)", err.Error(), string(out))
		return err
	}

	if !IsHexSuccessful(err) {
		err := genIntegrationErr("storage output failure")
		log.Errorf("storage: %s (%s)", err.Error(), string(out))
		return err
	}

	return nil
}

func VerifyStorage(name string) (*storages.VerficationResult, error) {
	nameMap := map[string]string{"name": name}
	input, err := json.Marshal(nameMap)
	if err != nil {
		err := genIntegrationErr("storage req parsing failure")
		log.Errorf("storage: %s (%v)", err.Error(), err)
		return nil, err
	}

	ctx, cancel := context.WithTimeout(wait.CtxMinutes(10))
	defer cancel()
	out, err := exec.CommandContext(ctx, "hex_sdk", "cinder_test_storage", string(input)).CombinedOutput()
	if err != nil {
		err := genIntegrationErr("storage verify exec failure")
		log.Errorf("storage: %s (%s)", err.Error(), string(out))
		return nil, err
	}

	if !IsHexSuccessful(err) {
		err := genIntegrationErr("storage verify output failure")
		log.Errorf("storage: %s (%s)", err.Error(), string(out))
		return nil, err
	}

	result := &storages.VerficationResult{}
	err = json.Unmarshal(out, result)
	if err != nil {
		err := genIntegrationErr("storage verify parsing failure")
		log.Errorf("storage: %s (%s)", err.Error(), string(out))
		return nil, err
	}

	return result, err
}

func DeleteStorage(name string) error {
	nameMap := map[string]string{"name": name}
	input, err := json.Marshal(nameMap)
	if err != nil {
		err := genIntegrationErr("storage req parsing failure")
		log.Errorf("storage: %s (%v)", err.Error(), err)
		return err
	}

	ctx, cancel := context.WithTimeout(wait.CtxMinutes(3))
	defer cancel()
	out, err := exec.CommandContext(ctx, "hex_sdk", "cinder_delete_storage", string(input)).CombinedOutput()
	if err != nil {
		err := genIntegrationErr("storage exec failure")
		log.Errorf("storage: %s (%s)", err.Error(), string(out))
		return err
	}

	if !IsHexSuccessful(err) {
		err := genIntegrationErr("storage output failure")
		log.Errorf("storage: %s (%s)", err.Error(), string(out))
		return err
	}

	return nil
}

func ListVendors() ([]string, error) {
	models, err := ListModels()
	if err != nil {
		return nil, err
	}

	vendors := []string{}
	isAdded := map[string]bool{}
	for _, model := range models {
		_, found := isAdded[model.Vendor]
		if !found {
			vendors = append(vendors, model.Vendor)
			isAdded[model.Vendor] = true
		}
	}

	return vendors, nil
}

func ListModels() ([]storages.Model, error) {
	ctx, cancel := context.WithTimeout(wait.CtxMinutes(3))
	defer cancel()
	out, err := exec.CommandContext(ctx, "hex_sdk", "cinder_get_models").CombinedOutput()
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
		log.Errorf("storage: failed to parse model list (%s)", err.Error())
		err := genIntegrationErr("model parsing failure")
		log.Errorf("storage: %s (%s)", err.Error(), string(out))
		return nil, err
	}

	return list, nil
}

func GetStorageModel(driver string) (*storages.Model, error) {
	models, err := ListModels()
	if err != nil {
		return nil, err
	}

	for _, model := range models {
		if model.Driver == driver {
			return &model, nil
		}
	}

	return nil, fmt.Errorf(
		"storage model %s not found", driver,
	)
}

func CheckStorageModelExist(driver string) (bool, error) {
	models, err := ListModels()
	if err != nil {
		return false, err
	}

	for _, m := range models {
		if m.Driver == driver {
			return true, nil
		}
	}

	return false, nil
}

func UpdateStorageModel(model storages.Model) error {
	input, err := json.Marshal(model)
	if err != nil {
		err := genIntegrationErr("model req parsing failure")
		log.Errorf("storage: %s (%v)", err.Error(), err)
		return err
	}

	ctx, cancel := context.WithTimeout(wait.CtxMinutes(3))
	defer cancel()
	out, err := exec.CommandContext(ctx, "hex_sdk", "cinder_put_model", string(input)).CombinedOutput()
	if err != nil {
		err := genIntegrationErr("model exec failure")
		log.Errorf("storage: %s (%s)", err.Error(), string(out))
		return err
	}

	if !IsHexSuccessful(err) {
		err := genIntegrationErr("model output failure")
		log.Errorf("storage: %s (%s)", err.Error(), string(out))
		return err
	}

	return nil
}

func DeleteStorageModel(driver string) error {
	model := map[string]string{"driver": driver}
	input, err := json.Marshal(model)
	if err != nil {
		err := genIntegrationErr("model req parsing failure")
		log.Errorf("storage: %s (%v)", err.Error(), err)
		return err
	}

	ctx, cancel := context.WithTimeout(wait.CtxMinutes(3))
	defer cancel()
	out, err := exec.CommandContext(ctx, "hex_sdk", "cinder_delete_model", string(input)).CombinedOutput()
	if err != nil {
		err := identifyStorageModelDeleteErr(driver, out)
		log.Errorf("storage: %s (%s)", err.Error(), string(out))
		return err
	}

	if !IsHexSuccessful(err) {
		err := genIntegrationErr("model output failure")
		log.Errorf("storage: %s (%s)", err.Error(), string(out))
		return err
	}

	wait.Seconds(3)
	if DoesStorageModelExist(driver) {
		return fmt.Errorf("%s is built-in model, can't be deleted", driver)
	}

	return nil
}

func DoesStorageModelExist(driver string) bool {
	model, err := GetStorageModel(driver)
	return err == nil && model != nil
}

func genIntegrationErr(description string) error {
	return fmt.Errorf(
		"cubecos has unexpected hex error, please contact support(%s)",
		description,
	)
}

func identifyStorageModelDeleteErr(driver string, output []byte) error {
	if !strings.Contains(string(output), "does not exist") {
		return genIntegrationErr("model exec failure")
	}

	wait.Seconds(3)
	if DoesStorageModelExist(driver) {
		return fmt.Errorf("%s is built-in model, can't be deleted", driver)
	}

	return nil
}

func convertStorageTimes(list *[]storages.Cinder) {
	for i, storage := range *list {
		if storage.IsBuiltIn {
			(*list)[i].UpdateTime = base.ActiveFirmwareUpdatedAt
			continue
		}

		if storage.UpdateTime == "" {
			(*list)[i].UpdateTime = time.NowRFC3339()
			continue
		}

		updateTime, err := ostime.Parse(time.FormatRFC3339ZUTC, storage.UpdateTime)
		if err != nil {
			log.Warnf("integrations: failed to parse storage %s update time %s (%v)", storage.Name, storage.UpdateTime, err)
			continue
		}

		(*list)[i].UpdateTime = time.LocalRFC3339(updateTime)
	}
}
