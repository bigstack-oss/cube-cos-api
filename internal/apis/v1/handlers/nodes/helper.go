package nodes

import (
	"fmt"
	"sort"

	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/queries"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/pages"
	"github.com/gin-gonic/gin"
	log "go-micro.dev/v5/logger"
)

type helper struct {
	c       *gin.Context
	reqId   string
	handler string

	node            string
	keyword         string
	products        []string
	licenseStatuses []string
	roles           []string

	page  *pages.Page
	watch bool
}

func initHelper(c *gin.Context, handler string) (*helper, error) {
	h := &helper{c: c, reqId: queries.GetReqId(c), handler: handler}
	return h, h.parseParamsByHandler()
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
	h.node = h.c.Param("nodeName")
	if h.node == "" {
		return fmt.Errorf("nodeName should be provided")
	}

	return h.parseWatch()
}

func (h *helper) listNodes() (*nodePage, error) {
	nodes := cubecos.ListNodesWithTimeSensitiveInfo()
	nodes = h.filterNodes(nodes)
	nodesPerPage := h.paginateNodes(nodes)
	h.sortNodes(&nodesPerPage)

	return &nodePage{
		Nodes: nodesPerPage,
		Page:  h.genPageInfo(nodes),
	}, nil
}

func (h *helper) sortNodes(node *[]nodes.Node) {
	sort.SliceStable(
		*node,
		func(i, j int) bool {
			return (*node)[i].Hostname < (*node)[j].Hostname
		},
	)
}

func (h *helper) getNode() (*nodes.Node, error) {
	node, err := cubecos.GetNodeWithTimeSensitiveInfo(h.node)
	if err != nil {
		log.Errorf("nodes(%s): failed to get node: %v", h.reqId, err)
		return nil, err
	}

	if node.IsLocal() {
		return node, nil
	}

	if node.IsDown() {
		return nil, fmt.Errorf("node %s is down", node.Hostname)
	}

	return node, nil
}
