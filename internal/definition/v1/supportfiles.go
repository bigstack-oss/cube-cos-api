package v1

import (
	"github.com/bigstack-oss/cube-cos-api/internal/status"
)

const (
	SupportFileDB            = "supportfiles"
	SupportFileReqCollection = "requests"
)

type SupportFile struct {
	Id          string         `json:"id"`
	Name        string         `json:"name"`
	Roles       []Role         `json:"roles"`
	SizeMiB     float64        `json:"sizeMiB"`
	Status      status.Details `json:"status"`
	Description string         `json:"description"`
}

func (s *SupportFile) SetError() {
	s.Status.Current = status.Error
}

func (s *SupportFile) SetCompleted() {
	s.Status.Current = status.Completed
}

func (s *SupportFile) GenTaskUpdate() SupportFile {
	return SupportFile{
		Id:     s.Id,
		Name:   s.Name,
		Status: s.Status,
	}
}
