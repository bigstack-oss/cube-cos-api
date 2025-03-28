package nodes

import (
	"fmt"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/http"
	"github.com/bigstack-oss/cube-cos-api/internal/api"
	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/gin-gonic/gin"
	log "go-micro.dev/v5/logger"
)

type helper struct {
	c       *gin.Context
	handler string

	nodeName      string
	keyword       string
	licenseStatus string
	roles         []string

	definition.Page
	watch bool
}

func initHelper(c *gin.Context, handler string) (*helper, error) {
	h := &helper{c: c, handler: handler}

	switch handler {
	case "listNodes":
		return h, h.parseListOptions()
	case "getNode":
		return h, h.parseGetOptions()
	}

	return h, nil
}

func (h *helper) parseListOptions() error {
	err := h.parsePage()
	if err != nil {
		return err
	}

	h.parseKeyword()
	h.parseRoles()
	h.parseLicenseStatus()

	return h.parseWatch()
}

func (h *helper) parseGetOptions() error {
	h.nodeName = h.c.Param("nodeName")
	if h.nodeName == "" {
		return fmt.Errorf("nodeName should be provided")
	}

	return h.parseWatch()
}

func (h *helper) listNodes() (*data, error) {
	nodes, err := definition.ListNodes()
	if err != nil {
		log.Errorf("request(%s): failed to get nodes: %s", api.GetReqId(h.c), err.Error())
		return nil, err
	}

	h.addLicenseInfoToNodes(&nodes)
	h.addDetailsToNodes(&nodes)
	nodes = h.filterNodes(nodes)

	pagedNodes, err := paginateNodes(nodes, h.Page)
	if err != nil {
		log.Errorf("request(%s): failed to paginate nodes: %s", api.GetReqId(h.c), err.Error())
		return nil, err
	}

	page, err := genPageInfo(nodes, h.Page)
	if err != nil {
		log.Errorf("request(%s): failed to gen page info: %s", api.GetReqId(h.c), err.Error())
		return nil, err
	}

	return &data{
		Nodes: pagedNodes,
		Page:  page,
	}, nil
}

func isPageReceived(num, size string) bool {
	return num != "" || size != ""
}

func (h *helper) getNode() (*definition.Node, error) {
	node, err := definition.GetNodeByHostname(h.nodeName)
	if err != nil {
		log.Errorf("request(%s): failed to get node: %s", api.GetReqId(h.c), err.Error())
		return nil, err
	}

	if node.IsLocal() {
		h.addLicenseToNode(node)
		h.addMetricsToNode(node)
		return node, nil
	}

	return h.askFromOtherNode(node)
}

func (h *helper) askFromOtherNode(node *definition.Node) (*definition.Node, error) {
	helper := http.GetGlobalHelper()
	resp, err := helper.R().
		SetResult(&api.NodeData{}).
		SetHeader(node.GenAuthHeader()).
		Get(node.GetNodeDetailsUrl())
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, fmt.Errorf(
			"failed to get node details %s: %d %s",
			node.Hostname,
			resp.StatusCode(),
			string(resp.Body()),
		)
	}

	nodeDetails := &resp.Result().(*api.NodeData).Data
	return nodeDetails, nil
}
