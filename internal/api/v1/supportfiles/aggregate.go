package supportfiles

import (
	"fmt"

	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/support"
)

func (h *helper) convertToFileSets(files []support.File) []support.FileSet {
	fileSets := h.aggregateToFileSets(files)
	fileSets = h.filterSupportFiles(fileSets)
	h.sortSupportFileSets(&fileSets)
	return fileSets
}

func (h *helper) aggregateToFileSets(files []support.File) []support.FileSet {
	fileSetMap := map[string]*support.FileSet{}

	for _, file := range files {
		key := file.Group
		_, found := fileSetMap[key]
		if !found {
			fileSetMap[key] = h.genSupportFileSet(file)
		}

		fileSetMap[key].SizeMiB += file.SizeMiB
		fileSetMap[key].Files = append(
			fileSetMap[key].Files,
			h.enrichSupportFileInfo(file),
		)
	}

	fileSets := []support.FileSet{}
	for _, fileSet := range fileSetMap {
		fileSets = append(fileSets, *fileSet)
	}

	return fileSets
}

func (h *helper) genSupportFileSet(file support.File) *support.FileSet {
	return &support.FileSet{
		Name:        fmt.Sprintf("%s Support File Set %s", v1.DataCenterVersion, file.Status.CreatedAt),
		Description: file.Description,
		SizeMiB:     0,
		Files:       []support.File{},
		Status:      file.Status,
	}
}

func (h *helper) enrichSupportFileInfo(file support.File) support.File {
	if file.Name == "" {
		file.Name = "filename will be generated later once the file is created"
		file.Url = "file url will be generated later once the file is created"
		file.Status.IsCreating = true
		return file
	}

	return support.File{
		Name:        file.Name,
		Group:       file.Group,
		Description: file.Description,
		SizeMiB:     file.SizeMiB,
		Url:         file.Url,
		Source:      file.Source,
	}
}
