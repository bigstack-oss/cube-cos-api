package integrations

import (
	"sort"
	ostime "time"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/base"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/integration"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/storages"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/time"
	log "go-micro.dev/v5/logger"
)

func (h *helper) convertToStorages(cinders []storages.Cinder) []integration.Storage {
	storages := []integration.Storage{}
	for _, cinder := range cinders {
		storages = append(
			storages,
			integration.Storage{
				Name:         cinder.Name,
				Type:         h.convertType(cinder.IsBuiltIn),
				Driver:       cinder.Driver,
				Vendor:       cinder.Vendor,
				ManagementIp: base.ManagementIp,
				UpdatedAt:    h.convertTime(cinder.UpdateTime),
				IsDefault:    cinder.IsDefault,
				Status: status.Storage{
					Current:      status.Ok,
					IsProcessing: false,
				},
			},
		)
	}

	return storages
}

func (h *helper) convertType(isBuiltIn bool) string {
	if isBuiltIn {
		return "built-in"
	}

	return "external"
}

func (h *helper) convertTime(updateTime string) string {
	if updateTime == "" {
		return ""
	}

	update, err := ostime.Parse(time.FormatRFC3339Z, updateTime)
	if err != nil {
		log.Warnf("integrations: failed to parse update time %s (%v)", updateTime, err)
		return updateTime
	}

	return time.RFC3339Z(update)
}

func (h *helper) sortStorages(storages *[]integration.Storage) {
	sort.Slice(*storages, func(i, j int) bool {
		return (*storages)[i].Name > (*storages)[j].Name
	})
}

func (h *helper) sortVendors(vendors *[]string) {
	sort.Slice(*vendors, func(i, j int) bool {
		return (*vendors)[i] > (*vendors)[j]
	})
}

func (h *helper) sortModels(models *[]storages.Model) {
	sort.Slice(*models, func(i, j int) bool {
		return (*models)[i].Vendor > (*models)[j].Vendor
	})
}

func (h *helper) eraseStorageExtraConfigFiles(storage *storages.CinderDetails) {
	if storage == nil || storage.ExtraConfigFiles == nil {
		return
	}

	storage.ExtraConfigFiles = []storages.ExtraConfigFile{}
}

func (h *helper) eraseModelExtraConfigFiles(models []storages.Model) {
	for i := range models {
		if len(models[i].ExtraConfigFiles) == 0 {
			continue
		}

		for j := range models[i].ExtraConfigFiles {
			models[i].ExtraConfigFiles[j].Content = ""
		}
	}
}
