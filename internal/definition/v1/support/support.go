package support

import (
	"path/filepath"
	"slices"
	"sync"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"
	json "github.com/json-iterator/go"
)

const (
	DefaultBucket     = "supportfiles"
	DefaultFileDir    = "/var/support"
	DefaultFileTmpDir = "/tmp/support-comment-file"
	Files             = "supportfiles"
	FileDB            = "supportfiles"
	FileReqCollection = "requests"
	ReqTTL            = 3600
)

var (
	localFiles = sync.Map{}
)

type ListFileOptions struct {
	AllNodes bool
	Host     string
}

type FileRequest struct {
	Name        string   `json:"name" bson:"name"`
	Description string   `json:"description" bson:"description"`
	Hosts       []string `json:"hosts"`
	CreatedAt   string   `json:"createdAt" bson:"createdAt"`
}

type FileSet struct {
	Name        string             `json:"name" bson:"name"`
	Description string             `json:"description" bson:"description"`
	Files       []File             `json:"files" bson:"files"`
	SizeMiB     float64            `json:"sizeMiB" bson:"sizeMiB"`
	Status      status.SupportFile `json:"status" bson:"status"`
}

type File struct {
	Name        string `json:"name" bson:"name"`
	Group       string `json:"group" bson:"group"`
	Description string `json:"description" bson:"description"`
	Source      `json:"source" bson:"source"`
	SizeMiB     float64            `json:"sizeMiB" bson:"sizeMiB"`
	Url         string             `json:"url" bson:"url"`
	Status      status.SupportFile `json:"status" bson:"status"`
}

type Source struct {
	Role string `json:"role" bson:"role"`
	Host string `json:"host" bson:"host"`
}

func (f *File) Bytes() []byte {
	b, err := json.Marshal(f)
	if err != nil {
		return []byte{}
	}

	return b
}

func (f *File) SetError() {
	f.Status.Current = status.Error
	f.Status.IsCreating = false
}

func (f *File) SetCompleted() {
	f.Status.Current = status.Completed
	f.Status.IsCreating = false
}

func (f *File) InitCreation(timeLocal string) {
	f.Status = status.SupportFile{
		Current:    status.Creating,
		Desired:    status.Create,
		IsCreating: true,
		CreatedAt:  timeLocal,
	}
}

func (f *File) GenTaskUpdate() File {
	return File{
		Group:  f.Group,
		Name:   f.Name,
		Url:    f.Url,
		Source: f.Source,
		Status: f.Status,
	}
}

func (f *File) ObjectKey() string {
	return filepath.Base(f.Name)
}

func (f *FileSet) IncludeRoles(roles []string) bool {
	for _, file := range f.Files {
		if slices.Contains(roles, file.Source.Role) {
			return true
		}
	}

	return false
}

func GetLocalFiles() *sync.Map {
	return &localFiles
}

func GetLocalFile(name string) File {
	val, loaded := localFiles.Load(name)
	if !loaded {
		return File{}
	}

	return val.(File)
}

func SetLocalFile(File File) {
	localFiles.Store(File.Name, File)
}

func ListLocalFiles() []File {
	files := []File{}
	localFiles.Range(func(key, value any) bool {
		files = append(files, value.(File))
		return true
	})

	return files
}
