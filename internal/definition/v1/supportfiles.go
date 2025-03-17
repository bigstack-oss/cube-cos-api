package v1

import (
	"sync"

	"github.com/bigstack-oss/cube-cos-api/internal/status"
	"github.com/blevesearch/bleve/v2"
	"github.com/google/uuid"
)

const (
	DefaultSupportFileDir    = "/var/support"
	DefaultSupportFileTmpDir = "/tmp/support-comment-file"
	SupportFiles             = "supportfiles"
	SupportFileDB            = "supportfiles"
	SupportFileReqCollection = "requests"
)

var (
	localSupportFiles   = sync.Map{}
	supportFileSearcher bleve.Index
)

type ListSupportFileOptions struct {
	AllNodes bool
}

type SupportFileRequest struct {
	Hosts []string `json:"hosts"`
}

type SupportFile struct {
	Id          string `json:"id,omitzero" bson:"id"`
	Name        string `json:"name" bson:"name"`
	Comment     string `json:"comment" bson:"comment"`
	Roles       []Role `json:"roles,omitzero" bson:"roles,omitzero"`
	Hosts       []Host `json:"hosts,omitzero" yaml:"-" bson:"-"`
	Node        `json:"node,omitzero" bson:"node,omitzero"`
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
		Id:      s.Id,
		Name:    s.Name,
		Comment: s.Comment,
		Node:    node,
		Status:  s.Status,
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

func GetLocalSupportFiles() *sync.Map {
	return &localSupportFiles
}

func GetLocalSupportFile(name string) SupportFile {
	val, loaded := localSupportFiles.Load(name)
	if !loaded {
		return SupportFile{}
	}

	return val.(SupportFile)
}

func SetLocalSupportFile(SupportFile SupportFile) {
	localSupportFiles.Store(SupportFile.Name, SupportFile)
}

func ListLocalSupportFiles() []SupportFile {
	supportFiles := []SupportFile{}
	localSupportFiles.Range(func(key, value any) bool {
		supportFiles = append(supportFiles, value.(SupportFile))
		return true
	})

	return supportFiles
}

func InitSupportFileSearchIndex() error {
	if supportFileSearcher != nil {
		return nil
	}

	var err error
	mapping := bleve.NewIndexMapping()
	supportFileSearcher, err = bleve.NewMemOnly(mapping)
	return err
}

func GetSupportFileSearcher() bleve.Index {
	return supportFileSearcher
}
