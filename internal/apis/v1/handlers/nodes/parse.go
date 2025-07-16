package nodes

import (
	"fmt"
	"strings"

	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/queries"
	nodes "github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
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
	case "listNodeDevices":
		return h.parseListDevicesOptions()
	case "addNodeDevice":
		return h.parseCreateDeviceOptions()
	case "promoteOrDemoteNodeDevice":
		return h.parsePromoteOrDemoteOptions()
	case "removeNodeDevice":
		return h.parseRemoveDeviceOptions()
	case "restartNodeOsd":
		return h.parseRestartOsdOptions()
	case "removeNodeOsd":
		return h.parseRemoveOsdOptions()
	case "patchNodeOsd":
		return h.parsePatchOsdOptions()
	case "patchDeviceTask":
		return h.parsePatchDeviceTaskOptions()
	case "patchOsdTask":
		return h.parsePatchOsdTaskOptions()
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
	h.page, err = queries.GetPage(h.c)
	return err
}

func (h *helper) parseWatch() error {
	var err error
	h.watch, err = queries.GetWatch(h.c)
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

func (h *helper) parseListDevicesOptions() error {
	h.node = h.c.Param("nodeName")
	if h.node == "" {
		return fmt.Errorf("node name should be provided")
	}

	var err error
	h.watch, err = queries.GetWatch(h.c)
	if err != nil {
		return err
	}

	return nil
}

func (h *helper) parseCreateDeviceOptions() error {
	h.node = h.c.Param("nodeName")
	if h.node == "" {
		return fmt.Errorf("node name should be provided")
	}

	err := h.c.ShouldBindJSON(&h.deviceReqOpts)
	if err != nil {
		return fmt.Errorf(
			"failed to parse create device options(%v)",
			err,
		)
	}
	if h.deviceReqOpts.Device == "" {
		return fmt.Errorf("device name should be provided")
	}

	h.deviceReqOpts.Hostname = h.node
	h.deviceReqOpts.Device = fmt.Sprintf("/dev/%s", h.deviceReqOpts.Device)
	h.deviceReqOpts.SetAdding()
	return nil
}

func (h *helper) parsePromoteOrDemoteOptions() error {
	h.node = h.c.Param("nodeName")
	if h.node == "" {
		return fmt.Errorf("node name should be provided")
	}

	h.device = h.c.Param("device")
	if h.device == "" {
		return fmt.Errorf("device name should be provided")
	}

	err := h.c.ShouldBindJSON(&h.deviceReqOpts)
	if err != nil {
		return fmt.Errorf(
			"failed to parse promote or demote options(%v)",
			err,
		)
	}

	return nil
}

func (h *helper) parseRemoveDeviceOptions() error {
	h.node = h.c.Param("nodeName")
	if h.node == "" {
		return fmt.Errorf("node name should be provided")
	}

	h.device = h.c.Param("device")
	if h.device == "" {
		return fmt.Errorf("device name should be provided")
	}

	h.deviceReqOpts = nodes.DeviceReqOpts{}
	h.deviceReqOpts.Hostname = h.node
	h.deviceReqOpts.Device = fmt.Sprintf("/dev/%s", h.device)
	h.deviceReqOpts.SetRemoving()
	return nil
}

func (h *helper) parseRestartOsdOptions() error {
	h.node = h.c.Param("nodeName")
	if h.node == "" {
		return fmt.Errorf("node name should be provided")
	}

	h.osdId = h.c.Param("id")
	if h.osdId == "" {
		return fmt.Errorf("osd id should be provided")
	}

	return nil
}

func (h *helper) parseRemoveOsdOptions() error {
	h.node = h.c.Param("nodeName")
	if h.node == "" {
		return fmt.Errorf("node name should be provided")
	}

	h.osdId = h.c.Param("id")
	if h.osdId == "" {
		return fmt.Errorf("osd id should be provided")
	}

	return nil
}

func (h *helper) parsePatchOsdOptions() error {
	h.node = h.c.Param("nodeName")
	if h.node == "" {
		return fmt.Errorf("node name should be provided")
	}

	h.osdId = h.c.Param("id")
	if h.osdId == "" {
		return fmt.Errorf("osd id should be provided")
	}

	err := h.c.ShouldBindJSON(&h.osdReqOpts)
	if err != nil {
		return fmt.Errorf(
			"failed to parse patch osd options(%v)",
			err,
		)
	}

	return nil
}

func (h *helper) parsePatchDeviceTaskOptions() error {
	h.node = h.c.Param("nodeName")
	if h.node == "" {
		return fmt.Errorf("node name should be provided")
	}

	err := h.c.ShouldBindJSON(&h.deviceReqOpts)
	if err != nil {
		return fmt.Errorf(
			"failed to parse patch device task options(%v)",
			err,
		)
	}

	return nil
}

func (h *helper) parsePatchOsdTaskOptions() error {
	h.node = h.c.Param("nodeName")
	if h.node == "" {
		return fmt.Errorf("node name should be provided")
	}

	err := h.c.ShouldBindJSON(&h.osdReqOpts)
	if err != nil {
		return fmt.Errorf(
			"failed to parse patch osd task options(%v)",
			err,
		)
	}

	return nil
}
