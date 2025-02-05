package healths

import definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"

func isPageRequired(page, pageSize string) bool {
	return page != "" || pageSize != ""
}

func (h *helper) isPageRequired() bool {
	return h.Page.Number > 0 || h.Page.Size > 0
}

func (h *helper) genPageInfo() (definition.Page, error) {
	return definition.Page{
		Total:  1,
		Number: 1,
		Size:   1,
	}, nil
}
