package cubecos

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/aws"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/http"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/math"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/openstack/v2"
	"github.com/bigstack-oss/cube-cos-api/internal/api"
	"github.com/bigstack-oss/cube-cos-api/internal/config"
	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/support"
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

	return append(
		localFiles,
		otherNodeFiles...,
	), nil
}

func ListHostSupportFiles(opts support.ListFileOptions) ([]support.File, error) {
	node, err := v1.GetNodeByHostname(opts.Host)
	if err != nil {
		log.Errorf("failed to get node by hostname: %s", err.Error())
		return nil, err
	}

	if node.IsLocal() {
		return support.ListLocalFiles(), nil
	}

	return getNodeSupportFiles(*node)
}

func ListSupportFilesFromOtherNodes() ([]support.File, error) {
	files := []support.File{}
	for _, node := range v1.ListNodes() {
		if node.IsLocal() {
			continue
		}

		file, err := getNodeSupportFiles(node)
		if err != nil {
			log.Errorf("failed to get supportFiles from node %s: %s", node.Hostname, err.Error())
			continue
		}

		files = append(files, file...)
	}

	return files, nil
}

func CreateSupportFile(file support.File) error {
	path, err := CreateSupportCommentFile(file)
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

func SetSupportFileComment(file support.File) error {

	log.Infof("create comment: %s", string(file.Bytes()))

	path, err := CreateSupportCommentFile(file)
	if err != nil {
		log.Errorf("supportFile: failed to create support comment file: %s", err.Error())
		return err
	}

	out, err := exec.Command("hex_config", "set_support_file_comment", file.Name, path).CombinedOutput()
	if err != nil {
		log.Errorf("supportFile: failed to set support file comment: %s", string(out))
		return err
	}

	return nil
}

func CreateSupportCommentFile(file support.File) (string, error) {
	err := os.MkdirAll(support.DefaultFileTmpDir, 0755)
	if err != nil {
		return "", err
	}

	filePath, err := genRandomFilePath()
	if err != nil {
		return "", err
	}

	err = os.WriteFile(filePath, []byte(file.Bytes()), 0644)
	if err != nil {
		return "", err
	}

	return filePath, nil
}

func GetSupportFile(file support.File) (string, error) {
	files, err := os.ReadDir(support.DefaultFileDir)
	if err != nil {
		return "", err
	}

	if len(files) == 0 {
		err := errors.New("no support file found")
		return "", err
	}

	info, err := findSupportFile(files, file)
	if err != nil {
		return "", err
	}

	return info.Name(), nil
}

func UploadSupportFileToObjectStore(supportFile support.File) error {
	file, err := os.Open(supportFile.Name)
	if err != nil {
		return err
	}

	defer file.Close()
	err = SyncBucketSecret()
	if err != nil {
		return err
	}

	err = SyncBucketStore()
	if err != nil {
		return err
	}

	return PutSupportFileToBucket(
		supportFile.ObjectKey(),
		file,
	)
}

func GetSupportFileUrl(file support.File) string {
	u := url.URL{}
	u.Scheme = "https"
	u.Host = fmt.Sprintf("%s:%d", v1.DataCenterVip, config.Opts.Saml.ServiceProvider.Host.Port)
	u.Path = fmt.Sprintf(
		"/api/v1/datacenters/%s/supportFiles/%s/%s",
		v1.DataCenterName,
		url.PathEscape(file.Group),
		url.PathEscape(file.Name),
	)

	return u.String()
}

func SyncBucketStore() error {
	h := aws.GetGlobalHelper()
	bucket := support.DefaultBucket
	_, err := h.CreateBucket(s3.CreateBucketInput{Bucket: &bucket})
	if err == nil {
		return nil
	}

	var isBucketAlreadyExists *types.BucketAlreadyExists
	if !errors.As(err, &isBucketAlreadyExists) {
		return err
	}

	return nil
}

func SyncBucketSecret() error {
	h := openstack.GetGlobalHelper()
	accessKey := config.Opts.Spec.Aws.AccessKey
	secretKey := config.Opts.Spec.Aws.SecretKey
	userId, err := h.GetUserIdByName(accessKey)
	if err != nil {
		return err
	}

	projectId, err := h.GetProjectIdByName(secretKey)
	if err != nil {
		return err
	}

	_, err = h.CreateEc2Credential(userId, projectId, accessKey, secretKey)
	if err == nil {
		return nil
	}
	if !strings.Contains(err.Error(), "Conflict") {
		return err
	}

	return nil
}

func PutSupportFileToBucket(key string, file io.Reader) error {
	h := aws.GetGlobalHelper()
	_, err := h.PutObject(support.DefaultBucket, key, file)
	if err != nil {
		return err
	}

	return nil
}

func findSupportFile(files []os.DirEntry, supportFile support.File) (os.DirEntry, error) {
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if !IsSupportFile(file.Name()) {
			continue
		}

		comment, err := GetSupportFileComment(file.Name())
		if err != nil {
			continue
		}

		if isCommentMatchWithFile(comment, supportFile) {
			return file, nil
		}
	}

	return nil, errors.New("no support file found")
}

func GetSupportFileComment(file string) (*support.File, error) {
	filePath := filepath.Join(support.DefaultFileDir, file)
	out, err := exec.Command("hex_config", "get_support_file_comment", filePath).CombinedOutput()
	if err != nil {
		return nil, err
	}

	s := &support.File{}
	err = json.Unmarshal([]byte(out), s)
	if err != nil {
		return nil, err
	}

	return s, nil
}

func IsSupportFile(file string) bool {
	prefix := fmt.Sprintf("CUBE_%s", v1.DataCenterNumericVersion)
	suffix := fmt.Sprintf("%s.support", v1.Hostname)
	return strings.HasPrefix(file, prefix) && strings.HasSuffix(file, suffix)
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

		info, err := GetSupportFileComment(supportFile.Name())
		if err != nil {
			continue
		}

		support.SetLocalFile(enrichSupportFile(
			supportFile,
			*info,
		))
	}
}

func isCommentMatchWithFile(comment *support.File, file support.File) bool {
	if comment.Group != file.Group {
		return false
	}

	if comment.Source.Host != file.Source.Host {
		return false
	}

	if comment.Status.CreatedAt != file.Status.CreatedAt {
		return false
	}

	return true
}

func genRandomFilePath() (string, error) {
	size := make([]byte, 8)
	_, err := rand.Read(size)
	if err != nil {
		return "", err
	}

	random := hex.EncodeToString(size)
	return filepath.Join(
		support.DefaultFileTmpDir,
		random,
	), nil
}

func parseSupportFile(file os.DirEntry) (fs.FileInfo, error) {
	if file.IsDir() {
		return nil, errors.New("not a file")
	}

	if !IsSupportFile(file.Name()) {
		return nil, errors.New("not a support file")
	}

	return file.Info()
}

func enrichSupportFile(file fs.FileInfo, info support.File) support.File {
	info.Name = file.Name()
	info.SizeMiB = math.RoundDown(float64(file.Size())/1024/1024, 4)
	return info
}
