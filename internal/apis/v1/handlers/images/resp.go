package images

import (
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/images"
)

type materials struct {
	ReservedImages []images.ReqOpts `json:"reservedImages"`
	Projects       []Project        `json:"projects"`
	Oses           []string         `json:"oses"`
	Destinations   []destination    `json:"destinations"`
	Domains        []string         `json:"domains"`
	Visibilities   []string         `json:"visibilities"`
}

type Project struct {
	Name        string `json:"name"`
	Domain      string `json:"domain"`
	Enabled     bool   `json:"enabled"`
	Description string `json:"description"`
}

type destination struct {
	Name      string `json:"name"`
	IsDefault bool   `json:"isDefault"`
}
