package volumes

import (
	"encoding/csv"
	"fmt"
	"strconv"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/volumes"
)

var (
	csvHeaders = []string{
		"id", "name", "type", "diskTag", "attachedTo", "bootable", "shared", "sizeMiB", "createdAt", "status",
	}
)

func (h *helper) convertToCsv(volumes []volumes.Volume) *csv.Writer {
	writer := csv.NewWriter(h.c.Writer)
	writer.Write(csvHeaders)
	for _, volume := range volumes {
		writer.Write([]string{
			volume.Id,
			volume.Name,
			volume.Type,
			volume.DiskTag,
			volume.AttachedTo,
			strconv.FormatBool(volume.Bootable),
			strconv.FormatBool(volume.Shared),
			strconv.FormatInt(volume.SizeMiB, 10),
			volume.CreatedAt,
			h.genStatusDesc(volume.Status),
		})
	}

	return writer
}

func (h *helper) genStatusDesc(status status.Volume) string {
	if !status.IsProcessing {
		return status.Current
	}

	return fmt.Sprintf(
		"%s(%.2f%%)",
		status.Current, status.ProcessPercent,
	)
}
