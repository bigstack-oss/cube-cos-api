package images

import (
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/images"
)

type materials struct {
	ReservedImages []images.Reserved `json:"reservedImages"`
	Projects       []Project         `json:"projects"`
	Oses           []string          `json:"oses"`
	Destinations   []string          `json:"destinations"`
	Domains        []string          `json:"domains"`
	Visibilities   []string          `json:"visibilities"`
}

type Project struct {
	Name        string `json:"name"`
	Domain      string `json:"domain"`
	Enabled     bool   `json:"enabled"`
	Description string `json:"description"`
}
