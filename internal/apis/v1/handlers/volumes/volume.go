package volumes

import (
	"strings"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/volumes"
	opsvolumes "github.com/gophercloud/gophercloud/v2/openstack/blockstorage/v3/volumes"
	log "go-micro.dev/v5/logger"
)

func (h *helper) listConvertedVolumes() ([]volumes.Volume, error) {
	opts, err := h.genListOpts()
	if err != nil {
		log.Errorf("volumes(%s): failed to generate volume list opts(%v)", h.reqId, err)
		return nil, err
	}

	list, err := h.openstack.ListVolumes(*opts)
	if err != nil {
		log.Errorf("volumes(%s): failed to list volumes(%v)", h.reqId, err)
		return nil, err
	}

	volumes := h.convertToVolumes(list)
	h.syncProcessingVolumes(&volumes)
	return volumes, nil
}

func (h *helper) genListOpts() (*opsvolumes.ListOpts, error) {
	id, err := h.openstack.GetProjectIdByName(h.project)
	if err != nil {
		log.Errorf("volumes(%s): failed to get project id(%v)", h.reqId, err)
		return nil, err
	}

	return &opsvolumes.ListOpts{
		AllTenants: true,
		TenantID:   id,
	}, nil
}

func (h *helper) convertToVolumes(list []opsvolumes.Volume) []volumes.Volume {
	converted := []volumes.Volume{}

	for _, volume := range list {
		converted = append(
			converted,
			volumes.Volume{
				Id:         volume.ID,
				Name:       volume.Name,
				Type:       volume.VolumeType,
				DiskTag:    h.parseDiskTag(volume.Attachments),
				AttachedTo: h.parseAttachedTo(volume.Attachments),
				Bootable:   strings.EqualFold(volume.Bootable, "true"),
				Shared:     volume.Multiattach,
				SizeMiB:    h.parseSizeToMiB(volume.Size),
				CreatedAt:  h.parseCreatedAt(volume.CreatedAt),
				Status: status.Volume{
					Current: volume.Status,
				},
			},
		)
	}

	return converted
}

func (h *helper) syncProcessingVolumes(volumes *[]volumes.Volume) {
	if !h.hasProcessingVolumes() {
		return
	}

	processings, err := h.getProcessingVolumes()
	if err != nil {
		log.Errorf("volumes(%s): failed to get processing volumes(%v)", h.reqId, err)
		return
	}

	existings := map[string]int{}
	for i, volume := range *volumes {
		existings[volume.Name] = i
	}

	for _, processing := range processings {
		updateIdx, found := existings[processing.Name]
		if !found {
			*volumes = append(*volumes, processing)
			continue
		}

		(*volumes)[updateIdx].Status.Current = processing.Status.Current
		(*volumes)[updateIdx].Status.IsProcessing = processing.Status.IsProcessing
		(*volumes)[updateIdx].Status.ProcessPercent = processing.Status.ProcessPercent
	}
}
