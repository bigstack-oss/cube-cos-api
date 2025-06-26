package nodes

import (
	"fmt"
	"strings"

	query "github.com/bigstack-oss/cube-cos-api/internal/apis/v1/queries"
)

func (h *helper) parseParamsByHandler() error {
	switch h.handler {
	case "listNodes":
		return h.parseListOptions()
	case "getNode", "getNodeIpmi", "disconnectNodeIpmi":
		return h.parseGetOptions()
	case "setNodeIpmi", "verifyNodeIpmi":
		return h.parseSetOrVerifyOptions()
	case "ipmiOperateNode":
		return h.parseIpmiOperateOptions()
	default:
		return fmt.Errorf(
			"unknown node handler: %s",
			h.handler,
		)
	}
}

func (h *helper) parseSetOrVerifyOptions() error {
	h.node = h.c.Param("nodeName")
	if h.node == "" {
		return fmt.Errorf("node name should be provided")
	}

	err := h.c.ShouldBindJSON(&h.ipmi)
	if err != nil {
		return err
	}

	return h.ipmi.CheckInvalidValues()
}

func (h *helper) parseIpmiOperateOptions() error {
	h.node = h.c.Param("nodeName")
	if h.node == "" {
		return fmt.Errorf("nodeName should be provided")
	}

	h.operation = strings.ToLower(h.c.Param("operation"))
	return h.checkIpmiOperation()
}

func (h *helper) parseKeyword() {
	keyword := h.c.DefaultQuery("keyword", "")
	h.keyword = strings.ToLower(keyword)
}

func (h *helper) parseProduct() {
	h.products = h.c.QueryArray("products")
}

func (h *helper) parseRoles() {
	h.roles = h.c.QueryArray("roles")
}

func (h *helper) parseLicenseStatus() {
	h.licenseStatuses = h.c.QueryArray("licenseStatuses")
}

func (h *helper) parsePage() error {
	var err error
	h.page, err = query.GetPage(h.c)
	return err
}

func (h *helper) parseWatch() error {
	var err error
	h.watch, err = query.GetWatch(h.c)
	if err != nil {
		return err
	}

	return nil
}

func (h *helper) getIpmiOperation() (string, error) {
	switch h.operation {
	case "poweron":
		return "on", nil
	case "poweroff":
		return "off", nil
	case "powercycle":
		return "cycle", nil
	default:
		return "", fmt.Errorf("unknown ipmi operation: %s", h.operation)
	}
}
