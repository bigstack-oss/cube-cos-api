package v1

type Page struct {
	Total          int64 `json:"total"`
	Number         int   `json:"number"`
	Size           int   `json:"size"`
	TotalItemCount int64 `json:"totalItemCount"`
}

func (p Page) IsRequired() bool {
	return p.Number > 0 || p.Size > 0
}

func IsPageRequired(pageNum, pageSize string) bool {
	return pageNum != "" || pageSize != ""
}

type Limit struct {
	Number      int    `json:"number"`
	Description string `json:"description"`
}
