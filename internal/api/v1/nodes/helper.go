package nodes

import (
	"fmt"
	"sort"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/http"
	"github.com/bigstack-oss/cube-cos-api/internal/api"
	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
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

	*v1.Page
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

func (h *helper) listNodes() (*nodePage, error) {
	nodes := h.filterNodes(v1.ListNodes())
	nodesPerPage := h.paginateNodes(nodes)
	h.sortNodes(&nodesPerPage)

	return &nodePage{
		Nodes: nodesPerPage,
		Page:  h.genPageInfo(nodes),
	}, nil
}

func (h *helper) sortNodes(node *[]v1.Node) {
	sort.SliceStable(
		*node,
		func(i, j int) bool {
			return (*node)[i].Hostname < (*node)[j].Hostname
		},
	)
}

func (h *helper) getNode() (*v1.Node, error) {
	node, err := v1.GetNodeByHostname(h.nodeName)
	if err != nil {
		log.Errorf("nodes(%s): failed to get node: %s", api.GetReqId(h.c), err.Error())
		return nil, err
	}

	if node.IsLocal() {
		return node, nil
	}

	if node.IsDown() {
		return nil, fmt.Errorf("node %s is down", node.Hostname)
	}

	return h.askPeerNode(node)
}

func (h *helper) askPeerNode(node *v1.Node) (*v1.Node, error) {
	helper := http.GetGlobalHelper()
	resp, err := helper.R().
		SetResult(&api.NodeData{}).
		SetHeaders(v1.GenNodeAuthHeaders()).
		Get(node.GetNodeDetailsUrl())
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, fmt.Errorf(
			"nodes(%s): failed to get node details %s: %d %s",
			api.GetReqId(h.c),
			node.Hostname,
			resp.StatusCode(),
			string(resp.Body()),
		)
	}

	nodeDetails := &resp.Result().(*api.NodeData).Data
	return nodeDetails, nil
}
