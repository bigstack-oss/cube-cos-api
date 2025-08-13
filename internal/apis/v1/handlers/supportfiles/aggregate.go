package supportfiles

import (
	"fmt"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/base"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/support"
)

func (h *helper) genFileSets(files []support.File) []support.FileSet {
	sets := h.aggregateToFileSets(files)
	sets = h.filterFiles(sets)
	h.sortFileSets(&sets)
	return sets
}

func (h *helper) aggregateToFileSets(files []support.File) []support.FileSet {
	setMap := h.groupFileSet(files)
	h.syncCreatingStatus(&setMap)

	sets := []support.FileSet{}
	for _, set := range setMap {
		sets = append(sets, *set)
	}

	return sets
}

func (h *helper) groupFileSet(files []support.File) map[string]*support.FileSet {
	setMap := map[string]*support.FileSet{}

	for _, file := range files {
		key := file.Group
		_, found := setMap[key]
		if !found {
			setMap[key] = h.genSupportFileSet(file)
		}

		setMap[key].SizeMiB += file.SizeMiB
		setMap[key].Files = append(
			setMap[key].Files,
			h.enrichFileInfo(file),
		)
	}

	return setMap
}

func (h *helper) genSupportFileSet(file support.File) *support.FileSet {
	return &support.FileSet{
		Name:        fmt.Sprintf("%s Support File Set %s", base.DataCenterFirmwareVersion, file.Status.CreatedAt),
		Description: file.Description,
		SizeMiB:     0,
		Files:       []support.File{},
		Status:      file.Status,
	}
}

func (h *helper) enrichFileInfo(file support.File) support.File {
	if file.IsCreating() {
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

func (h *helper) syncCreatingStatus(setMap *map[string]*support.FileSet) {
	for key, set := range *setMap {
		if len(set.Files) == 0 {
			continue
		}

		if h.areAllFilesCreated(set) {
			(*setMap)[key].Status.Current = status.Completed
			(*setMap)[key].Status.IsCreating = false
		} else {
			(*setMap)[key].Status.Current = status.Creating
			(*setMap)[key].Status.IsCreating = true
		}
	}
}

func (h *helper) areAllFilesCreated(set *support.FileSet) bool {
	for _, file := range set.Files {
		if file.Status.IsCreating {
			return false
		}
	}

	return true
}
