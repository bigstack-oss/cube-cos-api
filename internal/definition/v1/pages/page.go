package pages

type Page struct {
	Total          int64 `json:"total"`
	Number         int   `json:"number"`
	Size           int   `json:"size"`
	TotalItemCount int64 `json:"totalItemCount"`
}

type Limit struct {
	Number      int    `json:"number"`
	Description string `json:"description"`
}

func (p Page) IsRequired() bool {
	return p.Number > 0 || p.Size > 0
}
