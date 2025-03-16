package cubecos

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/http"
	"github.com/bigstack-oss/cube-cos-api/internal/api"
	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	log "go-micro.dev/v5/logger"
)

func ListSupportFiles(opts v1.ListSupportFileOptions) ([]v1.SupportFile, error) {
	localSupportFiles := definition.ListLocalSupportFiles()
	if !opts.AllNodes {
		return localSupportFiles, nil
	}

	allSupportFiles, err := ListSupportFilesFromOtherNodes()
	if err != nil {
		return nil, err
	}

	allSupportFiles[definition.Hostname] = localSupportFiles
	return aggregateSupportFiles(allSupportFiles), nil
}

func ListSupportFilesFromOtherNodes() (map[string][]v1.SupportFile, error) {
	nodes, err := definition.ListNodes()
	if err != nil {
		log.Errorf("failed to list nodes for supportFiles: %s", err.Error())
		return nil, err
	}

	supportFiles := map[string][]definition.SupportFile{}
	for _, node := range nodes {
		if node.IsLocal() {
			continue
		}

		supportFile, err := getNodeSupportFiles(*node)
		if err != nil {
			log.Errorf("failed to get supportFiles from node %s: %s", node.Name, err.Error())
			continue
		}

		supportFiles[node.Name] = supportFile
	}

	return supportFiles, nil
}

func CreateSupportFile(supportFile v1.SupportFile) error {
	path, err := CreateSupportCommentFile(supportFile.Comment)
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

	err = os.MkdirAll(v1.DefaultSupportFileTmpDir, 0755)
	if err != nil {
		return "", err
	}

	randomStr := hex.EncodeToString(randomSize)
	filePath := filepath.Join(v1.DefaultSupportFileTmpDir, randomStr)
	err = os.WriteFile(filePath, []byte(comment), 0644)
	if err != nil {
		return "", err
	}

	return filePath, nil
}

func GetSupportFileByComment(comment string) (string, error) {
	files, err := os.ReadDir(v1.DefaultSupportFileDir)
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
	filePath := filepath.Join(v1.DefaultSupportFileDir, file)
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

func getNodeSupportFiles(node definition.Node) ([]v1.SupportFile, error) {
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

func aggregateSupportFiles(nodeToSupportFile map[string][]definition.SupportFile) []definition.SupportFile {
	mergedMap := make(map[string]definition.SupportFile)
	for _, supportFiles := range nodeToSupportFile {
		setSupportFiles(mergedMap, supportFiles)
	}

	supportFiles := []definition.SupportFile{}
	for _, item := range mergedMap {
		supportFiles = append(supportFiles, item)
	}

	return supportFiles
}

func setSupportFiles(mergedMap map[string]definition.SupportFile, supportFiles []definition.SupportFile) {
	for _, supportFile := range supportFiles {
		key := supportFile.Comment
		existing, found := mergedMap[key]
		if found {
			existing.Hosts = slices.Concat(existing.Hosts, supportFile.Hosts)
			mergedMap[key] = existing
		} else {
			mergedMap[key] = supportFile
		}
	}
}

func SyncSupportFiles() {
	files, err := os.ReadDir(v1.DefaultSupportFileDir)
	if err != nil {
		log.Errorf("supportFile: failed to read support file directory: %s", err.Error())
		return
	}

	if len(files) == 0 {
		err := errors.New("no support file found")
		log.Errorf("supportFile: %v", err)
		return
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if !isSupportFile(file.Name()) {
			continue
		}

		comment, err := GetSupportFileComment(file.Name())
		if err != nil {
			log.Errorf("supportFile: %v", err)
			continue
		}

		if strings.Contains(comment, "Automatically") {
			continue
		}

		info, err := file.Info()
		if err != nil {
			log.Errorf("supportFile: %v", err)
			continue
		}

		definition.SetLocalSupportFile(definition.SupportFile{
			Name:        file.Name(),
			Comment:     comment,
			Hosts:       []v1.Host{{Name: v1.Hostname}},
			SizeMiB:     float64(info.Size()) / 1024 / 1024,
			Description: "",
		})
	}
}
