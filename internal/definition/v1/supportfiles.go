package v1

import (
	"github.com/bigstack-oss/cube-cos-api/internal/status"
	"github.com/google/uuid"
)

const (
	SupportFiles             = "supportfiles"
	SupportFileDB            = "supportfiles"
	SupportFileReqCollection = "requests"
)

type SupportFileRequest struct {
	Hosts []string `json:"hosts"`
}

type SupportFile struct {
	Id          string `json:"id" bson:"id"`
	Name        string `json:"name" bson:"name"`
	Group       string `json:"group" bson:"group"`
	Roles       []Role `json:"roles" bson:"roles"`
	Node        `json:"node" bson:"node"`
	SizeMiB     float64            `json:"sizeMiB" bson:"sizeMiB"`
	Url         string             `json:"url" bson:"url"`
	Status      status.SupportFile `json:"status" bson:"status"`
	Description string             `json:"description" bson:"description"`
}

func (s *SupportFile) SetError() {
	s.Status.Current = status.Error
}

func (s *SupportFile) SetCompleted() {
	s.Status.Current = status.Completed
}

func (s *SupportFile) InitCreateStatus() {
	s.Id = uuid.New().String()
	s.Status = status.SupportFile{
		Current:   status.Creating,
		Desired:   status.Create,
		CreatedAt: TimeLocal(),
	}
}

func (s *SupportFile) GenTaskUpdate() SupportFile {
	return SupportFile{
		Id:     s.Id,
		Name:   s.Name,
		Status: s.Status,
	}
}

func (s *SupportFile) GenTask(node Node) SupportFile {
	return SupportFile{
		Id:     s.Id,
		Name:   s.Name,
		Node:   node,
		Status: s.Status,
	}
}

func (s *SupportFile) SetRoleByHosts(hosts []string) {
	roleNodeMap := make(map[string][]*Node)
	for _, host := range hosts {
		node, err := GetNodeByHostname(host)
		if err != nil {
			continue
		}

		roleNodeMap[node.Role] = append(
			roleNodeMap[node.Role],
			node,
		)
	}

	for role, nodes := range roleNodeMap {
		s.Roles = append(
			s.Roles,
			Role{
				Name:  role,
				Nodes: nodes,
			},
		)
	}
}
