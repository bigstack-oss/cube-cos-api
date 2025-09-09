package integrations

import (
	"sort"
	ostime "time"

	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
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
				Type:         h.convertType(cinder.IsExternal),
				Vendor:       cinder.Device.Vendor,
				ManagementIp: base.ManagementIp,
				UpdatedAt:    h.convertTime(cinder.Storage.UpdateTime),
				IsDefault:    cubecos.IsDefaultStorage(cinder.Name),
				Status: status.Integration{
					Current:      status.Ok,
					IsProcessing: false,
				},
			},
		)
	}

	return storages
}

func (h *helper) convertType(isExternal bool) string {
	if isExternal {
		return "external"
	}

	return "built-in"
}

func (h *helper) convertTime(updateTime string) string {
	if updateTime == "" {
		return ""
	}

	update, err := ostime.Parse(updateTime, time.FormatRFC3339Z)
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

func (h *helper) sortModels(models *[]storages.Model) {
	sort.Slice(*models, func(i, j int) bool {
		return (*models)[i].Vendor > (*models)[j].Vendor
	})
}
