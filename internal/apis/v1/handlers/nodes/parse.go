package nodes

import (
	"fmt"
	"strings"

	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/queries"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/ceph"
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
	case "updateNodeDevice":
		return h.parsePromoteOrDemoteOptions()
	case "updateNodeDeviceOsds":
		return h.parseUpdateDeviceOsdsOptions()
	case "removeNodeDevice":
		return h.parseRemoveDeviceOptions()
	case "restartNodeOsd":
		return h.parseRestartOsdOptions()
	case "removeNodeOsd":
		return h.parseRemoveOsdOptions()
	case "updateNodeOsd":
		return h.parseUpdateOsdOptions()
	case "updateDeviceTask":
		return h.parseUpdateDeviceTaskOptions()
	case "updateOsdTask":
		return h.parseUpdateOsdTaskOptions()
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

	h.deviceReqOpts.ReqId = h.reqId
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

	h.deviceReqOpts.ReqId = h.reqId
	h.deviceReqOpts.Hostname = h.node
	h.deviceReqOpts.Device = fmt.Sprintf("/dev/%s", h.device)
	h.deviceReqOpts.SetUpdating()
	return nil
}

func (h *helper) parseUpdateDeviceOsdsOptions() error {
	h.node = h.c.Param("nodeName")
	if h.node == "" {
		return fmt.Errorf("node name should be provided")
	}

	h.device = h.c.Param("device")
	if h.device == "" {
		return fmt.Errorf("device name should be provided")
	}

	err := h.c.ShouldBindJSON(&h.osdReqOpts)
	if err != nil {
		return fmt.Errorf(
			"failed to parse patch osd options(%v)",
			err,
		)
	}

	if !h.isValidReweight(h.osdReqOpts.Reweight) {
		return fmt.Errorf("reweight should be between 0.0 ~ 1.0 and only allow two decimal places")
	}

	h.deviceReqOpts.Device = fmt.Sprintf("/dev/%s", h.device)
	return h.setOsdReqOptses()
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

	h.deviceReqOpts = nodes.DeviceReqOpts{ReqId: h.reqId}
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

	device, err := ceph.GetDeviceByOsdId(h.node, h.osdId)
	if err != nil {
		return fmt.Errorf("failed to get device by osd id(%s): %v", h.osdId, err)
	}

	h.osdReqOpts = nodes.OsdReqOpts{ReqId: h.reqId}
	h.osdReqOpts.Device = fmt.Sprintf("/dev/%s", device.Dev)
	h.osdReqOpts.Hostname = h.node
	h.osdReqOpts.OsdId = h.osdId
	h.osdReqOpts.SetRestarting()
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

	device, err := ceph.GetDeviceByOsdId(h.node, h.osdId)
	if err != nil {
		return fmt.Errorf("failed to get device by osd id(%s): %v", h.osdId, err)
	}

	h.osdReqOpts = nodes.OsdReqOpts{ReqId: h.reqId}
	h.osdReqOpts.Device = fmt.Sprintf("/dev/%s", device.Dev)
	h.osdReqOpts.Hostname = h.node
	h.osdReqOpts.OsdId = h.osdId
	h.osdReqOpts.SetRemoving()
	return nil
}

func (h *helper) parseUpdateOsdOptions() error {
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

	if !h.isValidReweight(h.osdReqOpts.Reweight) {
		return fmt.Errorf("reweight should be between 0.0 ~ 1.0 and only allow two decimal places")
	}

	device, err := ceph.GetDeviceByOsdId(h.node, h.osdId)
	if err != nil {
		return fmt.Errorf("failed to get device by osd id(%s): %v", h.osdId, err)
	}

	h.osdReqOpts.ReqId = h.reqId
	h.osdReqOpts.Device = fmt.Sprintf("/dev/%s", device.Dev)
	h.osdReqOpts.Hostname = h.node
	h.osdReqOpts.OsdId = h.osdId
	h.osdReqOpts.SetReweighting()
	return nil
}

func (h *helper) parseUpdateDeviceTaskOptions() error {
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

func (h *helper) parseUpdateOsdTaskOptions() error {
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

func (h *helper) setOsdReqOptses() error {
	deviceMap, err := ceph.GetDeviceMapByHost(h.node)
	if err != nil {
		return fmt.Errorf("failed to get device map by host %s(%v)", h.node, err)
	}

	device, found := deviceMap[h.device]
	if !found {
		return fmt.Errorf("device %s not found on host %s", h.device, h.node)
	}
	if len(device.Osds) == 0 {
		return fmt.Errorf("device %s has no OSDs on host %s", h.device, h.node)
	}

	for _, osd := range device.Osds {
		odsReqOpts := nodes.OsdReqOpts{
			Device:   fmt.Sprintf("/dev/%s", device.Dev),
			Hostname: h.node,
			OsdId:    osd.Id,
			Reweight: h.osdReqOpts.Reweight,
		}

		odsReqOpts.SetReweighting()
		h.osdReqOptses = append(h.osdReqOptses, odsReqOpts)
	}

	return nil
}
