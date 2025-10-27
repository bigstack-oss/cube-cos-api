package images

import (
	"encoding/csv"
	"fmt"
	"strconv"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/images"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"
	opsimage "github.com/gophercloud/gophercloud/v2/openstack/image/v2/images"
)

var (
	csvHeaders = []string{
		"name", "os", "destination", "domain", "project", "visibility", "sizeMiB", "createdAt", "status",
	}
)

func (h *helper) convertToCsv(images []images.Image) *csv.Writer {
	writer := csv.NewWriter(h.c.Writer)
	writer.Write(csvHeaders)
	for _, image := range images {
		writer.Write([]string{
			image.Name,
			image.Os,
			image.Destination,
			image.Domain,
			image.Project,
			image.Visibility,
			strconv.FormatInt(image.SizeMiB, 10),
			image.CreatedAt,
			h.genStatusDesc(image.Status),
		})
	}

	return writer
}

func (h *helper) genStatusDesc(status status.Image) string {
	if !status.IsProcessing {
		return status.Current
	}

	return fmt.Sprintf(
		"%s(%.2f%%)",
		status.Current, status.ProcessPercent,
	)
}

func (h *helper) genImageUpdateOpts() opsimage.UpdateOpts {
	opts := opsimage.UpdateOpts{}
	if h.reqOpts.Name != "" {
		opts = append(opts, opsimage.ReplaceImageName{NewName: h.reqOpts.Name})
	}

	if h.reqOpts.Os != "" {
		opts = append(opts, opsimage.UpdateImageProperty{
			Op:    opsimage.ReplaceOp,
			Name:  images.DefaultOsDistro,
			Value: h.reqOpts.Os,
		})
	}

	if h.reqOpts.Visibility != "" {
		opts = append(opts, opsimage.UpdateVisibility{
			Visibility: h.convertVisibility(h.reqOpts.Visibility),
		})
	}

	return opts
}

func (h *helper) convertVisibility(visibility string) opsimage.ImageVisibility {
	switch visibility {
	case "public":
		return "public"
	case "private":
		return "private"
	case "shared":
		return "shared"
	case "community":
		return "community"
	default:
		return "unknown"
	}
}
