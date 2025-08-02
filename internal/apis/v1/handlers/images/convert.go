package images

import (
	"encoding/csv"
	"fmt"
	"strconv"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/images"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"
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
