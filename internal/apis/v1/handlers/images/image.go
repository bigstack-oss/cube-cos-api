package images

import (
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/images"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"
	opsimage "github.com/gophercloud/gophercloud/v2/openstack/image/v2/images"
	log "go-micro.dev/v5/logger"
)

func (h *helper) listConvertedImages() ([]images.Image, error) {
	list, err := h.openstack.ListImages()
	if err != nil {
		log.Errorf("images(%s): failed to list images(%v)", h.reqId, err)
		return nil, err
	}

	images := h.convertToImages(list)
	h.syncProcessingImages(&images)
	return images, nil
}

func (h *helper) convertToImages(list []opsimage.Image) []images.Image {
	converted := []images.Image{}

	for _, image := range list {
		converted = append(
			converted,
			images.Image{
				Id:          image.ID,
				Name:        image.Name,
				Os:          h.parseOs(image.Properties),
				Destination: h.parseDestination(image.Properties),
				Domain:      h.parseDomain(image.Owner),
				Project:     h.parseProjectName(image.Owner),
				Visibility:  h.parseVisibility(image.Visibility),
				SizeMiB:     h.parseSizeMiB(image.SizeBytes),
				CreatedAt:   h.parseCreatedAt(image.CreatedAt),
				Status: status.Image{
					Current: h.parseStatus(image.Status),
				},
			},
		)
	}

	return converted
}

func (h *helper) syncProcessingImages(images *[]images.Image) {
	if !h.hasProcessingImages() {
		return
	}

	processings, err := h.getProcessingImages()
	if err != nil {
		log.Errorf("images(%s): failed to get processing images(%v)", h.reqId, err)
		return
	}

	existings := map[string]int{}
	for i, image := range *images {
		existings[image.Name] = i
	}

	for _, processing := range processings {
		updateIdx, found := existings[processing.Name]
		if !found {
			*images = append(*images, processing)
			continue
		}

		(*images)[updateIdx].Status.Current = processing.Status.Current
		(*images)[updateIdx].Status.IsProcessing = processing.Status.IsProcessing
		(*images)[updateIdx].Status.ProcessPercent = processing.Status.ProcessPercent
	}
}
