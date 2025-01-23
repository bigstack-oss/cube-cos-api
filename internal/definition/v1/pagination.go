package v1

type Page struct {
	Total  int64 `json:"total"`
	Number int   `json:"number"`
	Size   int   `json:"size"`
}

func (p Page) IsRequired() bool {
	return p.Number > 0 || p.Size > 0
}

func IsPaginationRequired(pageNum, pageSize string) bool {
	return pageNum != "" || pageSize != ""
}
