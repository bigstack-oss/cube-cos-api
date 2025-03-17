package cubecos

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/http"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/math"
	"github.com/bigstack-oss/cube-cos-api/internal/api"
	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/support"
	"github.com/bigstack-oss/cube-cos-api/internal/status"
	log "go-micro.dev/v5/logger"
)

func ListSupportFiles(opts support.ListFileOptions) ([]support.File, error) {
	localFiles := support.ListLocalFiles()
	if !opts.AllNodes {
		return localFiles, nil
	}

	otherNodeFiles, err := ListSupportFilesFromOtherNodes()
	if err != nil {
		return nil, err
	}

	return append(localFiles, otherNodeFiles...), nil
}

func ListSupportFilesFromOtherNodes() ([]support.File, error) {
	nodes, err := v1.ListNodes()
	if err != nil {
		log.Errorf("failed to list nodes for supportFiles: %s", err.Error())
		return nil, err
	}

	files := []support.File{}
	for _, node := range nodes {
		if node.IsLocal() {
			continue
		}

		file, err := getNodeSupportFiles(*node)
		if err != nil {
			log.Errorf("failed to get supportFiles from node %s: %s", node.Name, err.Error())
			continue
		}

		files = append(files, file...)
	}

	return files, nil
}

func CreateSupportFile(supportFile support.File) error {
	path, err := CreateSupportCommentFile(supportFile.Group)
	if err != nil {
		log.Errorf("supportFile: failed to create support comment file: %s", err.Error())
		return err
	}

	out, err := exec.Command("hex_config", "create_support_file", path).CombinedOutput()
	if err != nil {
		log.Errorf("supportFile: failed to create support file: %s", string(out))
		return err
	}

	return nil
}

func CreateSupportCommentFile(comment string) (string, error) {
	randomSize := make([]byte, 8)
	_, err := rand.Read(randomSize)
	if err != nil {
		return "", err
	}

	err = os.MkdirAll(support.DefaultFileTmpDir, 0755)
	if err != nil {
		return "", err
	}

	randomStr := hex.EncodeToString(randomSize)
	filePath := filepath.Join(support.DefaultFileTmpDir, randomStr)
	err = os.WriteFile(filePath, []byte(comment), 0644)
	if err != nil {
		return "", err
	}

	return filePath, nil
}

func GetSupportFileByComment(comment string) (string, error) {
	files, err := os.ReadDir(support.DefaultFileDir)
	if err != nil {
		log.Errorf("supportFile: failed to read support file directory: %s", err.Error())
		return "", err
	}

	if len(files) == 0 {
		err := errors.New("no support file found")
		log.Errorf("supportFile: %v", err)
		return "", err
	}

	file, err := findSupportFile(files, comment)
	if err != nil {
		log.Errorf("supportFile: %v", err)
		return "", err
	}

	return filepath.Join("/var/support", file.Name()), nil
}

func findSupportFile(files []os.DirEntry, comment string) (os.DirEntry, error) {
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if !isSupportFile(file.Name()) {
			continue
		}

		content, err := GetSupportFileComment(file.Name())
		if err != nil {
			continue
		}

		if string(content) == comment {
			return file, nil
		}
	}

	return nil, errors.New("no support file found")
}

func GetSupportFileComment(file string) (string, error) {
	filePath := filepath.Join(support.DefaultFileDir, file)
	out, err := exec.Command("hex_config", "get_support_file_comment", filePath).CombinedOutput()
	if err != nil {
		return "", err
	}

	return string(out), nil
}

func isSupportFile(file string) bool {
	return strings.HasPrefix(file, fmt.Sprintf("CUBE_%s", v1.DataCenterNumericVersion)) &&
		strings.HasSuffix(file, fmt.Sprintf("%s.support", v1.Hostname))
}

func getNodeSupportFiles(node v1.Node) ([]support.File, error) {
	h := http.GetGlobalHelper()
	resp, err := h.R().
		SetResult(&api.SupportFileListData{}).
		SetHeader(node.GenAuthHeader()).
		Get(node.GetSupportFileUrl())
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, fmt.Errorf(
			"failed to get support file from %s: %d %s",
			node.Hostname,
			resp.StatusCode(),
			string(resp.Body()),
		)
	}

	supportFileList := resp.Result().(*api.SupportFileListData)
	return supportFileList.Data, nil
}

func SyncSupportFiles() {
	files, err := os.ReadDir(support.DefaultFileDir)
	if err != nil {
		log.Errorf("supportFile: failed to read support file directory: %s", err.Error())
		return
	}

	if len(files) == 0 {
		log.Warnf("supportFile: no support file found")
		return
	}

	findAndParseSupportFiles(files)
}

func findAndParseSupportFiles(files []os.DirEntry) {
	for _, file := range files {
		supportFile, err := parseSupportFile(file)
		if err != nil {
			continue
		}

		comment, err := GetSupportFileComment(supportFile.Name())
		if err != nil {
			continue
		}

		support.SetLocalFile(genSupportFile(
			supportFile,
			comment,
		))
	}
}

func parseSupportFile(file os.DirEntry) (fs.FileInfo, error) {
	if file.IsDir() {
		return nil, errors.New("not a file")
	}

	if !isSupportFile(file.Name()) {
		return nil, errors.New("not a support file")
	}

	return file.Info()
}

func genSupportFile(file fs.FileInfo, comment string) support.File {
	return support.File{
		Name:  file.Name(),
		Group: comment,
		Source: support.Source{
			Role: v1.CurrentRole,
			Host: v1.Hostname,
		},
		SizeMiB:     math.RoundDown(float64(file.Size())/1024/1024, 4),
		Description: "",
		Status: status.SupportFile{
			Current:    status.Completed,
			IsCreating: false,
			CreatedAt:  v1.TimeISO8601Z(file.ModTime()),
		},
	}
}
