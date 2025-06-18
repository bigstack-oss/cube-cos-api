package supportfiles

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	ostime "time"

	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/queries"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/support"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/time"
)

func (h *helper) parseParamsByHandler() error {
	switch h.handler {
	case "listSupportFiles":
		return h.parseListParams()
	case "listHostSupportFiles":
		return h.parseHostListParams()
	case "createSupportFile":
		return h.parseCreateParams()
	case "deleteSupportFileGroup":
		return h.parseDeleteGroupParams()
	case "deleteSupportFile":
		return h.parseDeleteFileParams()
	case "downloadSupportFile":
		return h.parseDownloadParams()
	case "updateSupportFileTask":
		return h.parseUpdateParams()
	default:
		return nil
	}
}

func (h *helper) parseListParams() error {
	h.parseKeyword()
	h.parseHost()
	h.parseRoles()

	err := h.parsePage()
	if err != nil {
		return err
	}

	err = h.parseWatch()
	if err != nil {
		return err
	}

	err = h.parsePast()
	if err != nil {
		return err
	}

	err = h.parsePeriod()
	if err != nil {
		return err
	}

	if queries.ArePeriodAndPastEmpty(h.c) {
		h.past = "24h"
	}

	return nil
}

func (h *helper) parseHostListParams() error {
	h.host = h.c.Param("hostname")
	if h.host == "" {
		return errors.New("hostname is empty")
	}

	return nil
}

func (h *helper) parseCreateParams() error {
	return h.parseHosts()
}

func (h *helper) parseDeleteGroupParams() error {
	group, err := url.PathUnescape(h.c.Param("supportFileGroup"))
	if err != nil {
		return err
	}

	h.group.Name = group
	return nil
}

func (h *helper) parseDeleteFileParams() error {
	h.file = support.File{Name: h.c.Param("supportFileName")}
	return nil
}

func (h *helper) parseDownloadParams() error {
	group, err := url.PathUnescape(h.c.Param("supportFileGroup"))
	if err != nil {
		return err
	}

	h.group.Name = group
	h.file.Name = h.c.Param("supportFileName")
	return nil
}

func (h *helper) parseUpdateParams() error {
	return h.c.ShouldBindJSON(&h.file)
}

func (h *helper) parseKeyword() {
	keyword := h.c.DefaultQuery("keyword", "")
	h.keyword = strings.ToLower(keyword)
}

func (h *helper) parseWatch() error {
	var err error
	h.watch, err = queries.GetWatch(h.c)
	if err != nil {
		return err
	}

	return nil
}

func (h *helper) parseRoles() {
	h.roles = h.c.QueryArray("roles")
}

func (h *helper) parseHost() {
	h.host = h.c.DefaultQuery("host", "")
}

func (h *helper) parseHosts() error {
	err := h.c.ShouldBindJSON(&h.fileReq)
	if err != nil {
		return err
	}

	h.fileReq.CreatedAt = time.ISO8601Z(ostime.Now())
	return nil
}

func (h *helper) parsePast() error {
	var err error
	h.past, err = queries.GetPast(h.c)
	return err
}

func (h *helper) parsePeriod() error {
	var err error
	h.Period, err = queries.GetPeriod(h.c)
	return err
}

func (h *helper) parsePage() error {
	var err error
	h.Page, err = queries.GetPage(h.c)
	return err
}

func (h *helper) isPeriodRequired() bool {
	return h.c.DefaultQuery("stop", "") != "" || h.c.DefaultQuery("start", "") != ""
}

func (h *helper) isFilterRequired() bool {
	return h.isKeywordRequired() || h.isRoleRequired() || h.isPeriodRequired()
}

func (h *helper) isKeywordRequired() bool {
	return h.keyword != ""
}

func (h *helper) isRoleRequired() bool {
	return len(h.roles) > 0
}

func (h *helper) getHostsByGroup(group string) ([]string, error) {
	sets, err := h.listSupportFileSets()
	if err != nil {
		return nil, err
	}

	for _, set := range sets {
		if set.Name == group {
			return h.parseHostsFromSet(set)
		}
	}

	return nil, fmt.Errorf(
		"support file group(%s) not found",
		group,
	)
}

func (h *helper) parseHostsFromSet(set support.FileSet) ([]string, error) {
	hosts := []string{}
	for _, file := range set.Files {
		if file.IsCreating() {
			return nil, errors.New("has creating file, skip to delete")
		}

		if file.Source.Host == "" {
			continue
		}

		hosts = append(
			hosts,
			file.Source.Host,
		)
	}

	return hosts, nil
}
