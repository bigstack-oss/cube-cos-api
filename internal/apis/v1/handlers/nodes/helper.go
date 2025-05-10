package nodes

import (
	"fmt"
	"sort"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/http"
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/bodies"
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/queries"
	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/auth"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
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

	page  *v1.Page
	watch bool
}

func initHelper(c *gin.Context, handler string) (*helper, error) {
	h := &helper{c: c, handler: handler}
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
	h.nodeName = h.c.Param("nodeName")
	if h.nodeName == "" {
		return fmt.Errorf("nodeName should be provided")
	}

	return h.parseWatch()
}

func (h *helper) listNodes() (*nodePage, error) {
	nodes := h.filterNodes(nodes.List())
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
	node, err := nodes.Get(h.nodeName)
	if err != nil {
		log.Errorf("nodes(%s): failed to get node: %s", queries.GetReqId(h.c), err.Error())
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

func (h *helper) askPeerNode(node *nodes.Node) (*nodes.Node, error) {
	http := http.GetGlobalHelper()
	resp, err := http.R().SetResult(&bodies.Node{}).SetHeaders(auth.GetNodeSecret()).Get(node.GetNodeUrl())
	if err != nil {
		return nil, err
	}

	if !resp.IsError() {
		return &resp.Result().(*bodies.Node).Data, nil
	}

	return nil, fmt.Errorf(
		"nodes(%s): has err resp for node details %s: %s",
		queries.GetReqId(h.c),
		node.Hostname,
		string(resp.Body()),
	)
}
