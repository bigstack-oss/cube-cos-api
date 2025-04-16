package nodes

import (
	"fmt"
	"sort"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/http"
	"github.com/bigstack-oss/cube-cos-api/internal/api"
	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/gin-gonic/gin"
	log "go-micro.dev/v5/logger"
)

type helper struct {
	c       *gin.Context
	handler string

	nodeName        string
	keyword         string
	products        []string
	licenseStatuses []string
	roles           []string

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

	h.parseProduct()
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
	nodes := definition.ListNodes()
	nodes = h.filterNodes(nodes)
	nodesPerPage := h.paginateNodes(nodes)
	h.sortNodes(&nodesPerPage)

	return &data{
		Nodes: nodesPerPage,
		Page:  genPageInfo(nodes, h.Page),
	}, nil
}

func (h *helper) sortNodes(node *[]definition.Node) {
	sort.SliceStable(
		*node,
		func(i, j int) bool {
			return (*node)[i].Hostname < (*node)[j].Hostname
		},
	)
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
