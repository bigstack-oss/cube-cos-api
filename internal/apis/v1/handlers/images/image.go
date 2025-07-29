package images

import (
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/images"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"
	log "go-micro.dev/v5/logger"
)

func (h *helper) listConvertedImages() ([]images.Image, error) {
	list, err := h.openstack.ListImages()
	if err != nil {
		log.Errorf("images(%s): failed to list images(%v)", h.reqId, err)
		return nil, err
	}

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

	return converted, nil
}
