package nodes

import (
	"fmt"
	"strconv"

	"github.com/bigstack-oss/cube-cos-api/internal/api"
	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/gin-gonic/gin"
	log "go-micro.dev/v5/logger"
)

type helper struct {
	c *gin.Context
	definition.Page
	watch bool
}

func initReqHelper(c *gin.Context) (*helper, error) {
	h := &helper{c: c}

	err := h.parsePage()
	if err != nil {
		return nil, err
	}

	err = h.parseWatch()
	if err != nil {
		return nil, err
	}

	return h, nil
}

func (h *helper) parsePage() error {
	num := h.c.DefaultQuery("pageNum", "")
	size := h.c.DefaultQuery("pageSize", "")
	if !isPageReceived(num, size) {
		return nil
	}

	if num == "" {
		return fmt.Errorf("pageNum should be provided if pageSize is provided")
	}

	if size == "" {
		return fmt.Errorf("pageSize should be provided if pageNum is provided")
	}

	var err error
	h.Page.Number, err = strconv.Atoi(num)
	if err != nil {
		return fmt.Errorf("pageNum should be an integer: %s", num)
	}

	h.Page.Size, err = strconv.Atoi(size)
	if err != nil {
		return fmt.Errorf("pageSize should be an integer: %s", size)
	}

	if h.Page.Number <= 0 {
		return fmt.Errorf("pageNum should be greater than 0 if pageSize is provided")
	}

	if h.Page.Size <= 0 {
		return fmt.Errorf("pageSize should be greater than 0 if pageNum is provided")
	}

	return nil
}

func (h *helper) parseWatch() error {
	var err error
	h.watch, err = api.ParseWatch(h.c)
	if err != nil {
		return err
	}

	return nil
}

func (h *helper) getNodesResp() (*data, error) {
	nodes, err := definition.ListNodes()
	if err != nil {
		log.Errorf("request(%s): failed to get nodes: %s", api.GetReqId(h.c), err.Error())
		return nil, err
	}

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

	addLicenseInfoToNodes(h.c, &pagedNodes)
	addNodeDetailsToNodes(h.c, &pagedNodes)
	return &data{
		Nodes: pagedNodes,
		Page:  page,
	}, nil
}

func isPageReceived(num, size string) bool {
	return num != "" || size != ""
}
