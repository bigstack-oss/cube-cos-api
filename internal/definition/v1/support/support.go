package support

import (
	"sync"

	"github.com/bigstack-oss/cube-cos-api/internal/status"
	"github.com/blevesearch/bleve/v2"
	json "github.com/json-iterator/go"
)

const (
	DefaultFileDir    = "/var/support"
	DefaultFileTmpDir = "/tmp/support-comment-file"
	Files             = "supportfiles"
	FileDB            = "supportfiles"
	FileReqCollection = "requests"
)

var (
	localFiles   = sync.Map{}
	fileSearcher bleve.Index
)

type ListFileOptions struct {
	AllNodes bool
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
	Group       string `json:"comment" bson:"comment"`
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
}

func (f *File) SetCompleted() {
	f.Status.Current = status.Completed
}

func (f *File) InitCreation(timeLocal string) {
	f.Status = status.SupportFile{
		Current:   status.Creating,
		Desired:   status.Create,
		CreatedAt: timeLocal,
	}
}

func (f *File) GenTaskUpdate() File {
	return File{
		Name:   f.Name,
		Status: f.Status,
	}
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

func InitFileSearchIndex() error {
	if fileSearcher != nil {
		return nil
	}

	var err error
	mapping := bleve.NewIndexMapping()
	fileSearcher, err = bleve.NewMemOnly(mapping)
	return err
}

func GetFileSetSearcher() bleve.Index {
	return fileSearcher
}
