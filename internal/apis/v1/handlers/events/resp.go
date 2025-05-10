package events

import (
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/pages"
)

type data struct {
	Events       any `json:"events"`
	*pages.Page  `json:"page,omitempty"`
	*pages.Limit `json:"limit,omitempty"`
}
