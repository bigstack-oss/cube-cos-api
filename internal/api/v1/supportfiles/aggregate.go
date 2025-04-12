package supportfiles

import (
	"fmt"

	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/support"
)

func (h *helper) convertToFileSets(files []support.File) []support.FileSet {
	fileSets := aggregateToFileSets(files)
	fileSets = h.filterSupportFiles(fileSets)
	h.sortSupportFileSets(&fileSets)
	return fileSets
}

func aggregateToFileSets(files []support.File) []support.FileSet {
	fileSetMap := map[string]*support.FileSet{}

	for _, file := range files {
		key := file.Group
		if _, exists := fileSetMap[key]; !exists {
			fileSetMap[key] = &support.FileSet{
				Name:        fmt.Sprintf("%s Support File Set %s", v1.DataCenterVersion, file.Status.CreatedAt),
				Description: file.Description,
				SizeMiB:     0,
				Files:       []support.File{},
				Status:      file.Status,
			}
		}

		fileSetMap[key].SizeMiB += file.SizeMiB
		file := support.File{
			Name:        file.Name,
			Group:       file.Group,
			Description: file.Description,
			SizeMiB:     file.SizeMiB,
			Url:         file.Url,
			Source:      file.Source,
		}

		fileSetMap[key].Files = append(
			fileSetMap[key].Files,
			file,
		)
	}

	fileSets := []support.FileSet{}
	for _, fileSet := range fileSetMap {
		fileSets = append(fileSets, *fileSet)
	}

	return fileSets
}
