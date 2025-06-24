package nodes

import (
	"fmt"
	"sort"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/ipmi"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/mongo"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/password"
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/queries"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/base"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/pages"
	"github.com/gin-gonic/gin"
	log "go-micro.dev/v5/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type helper struct {
	c       *gin.Context
	mongo   *mongo.Helper
	reqId   string
	handler string

	node            string
	keyword         string
	products        []string
	licenseStatuses []string
	roles           []string
	ipmi            nodes.Ipmi
	operation       string

	page  *pages.Page
	watch bool
}

func initHelper(c *gin.Context, handler string) (*helper, error) {
	h := &helper{
		c:       c,
		mongo:   mongo.GetGlobalHelper(),
		reqId:   queries.GetReqId(c),
		handler: handler,
	}

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
		log.Errorf("nodes(%s): failed to get node(%v)", h.reqId, err)
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

func (h *helper) getNodeIpmi() (*nodes.Ipmi, error) {
	doc, err := h.mongo.Get(nodes.Db, nodes.CollectionIpmi, bson.M{"host": h.node})
	if err != nil {
		log.Errorf("nodes(%s): failed to get node ipmi(%v)", h.reqId, err)
		return nil, err
	}
	if doc == nil {
		return nil, fmt.Errorf("node %s ipmi not found", h.node)
	}

	impi := &nodes.Ipmi{}
	err = doc.Decode(impi)
	if err != nil {
		log.Errorf("nodes(%s): failed to decode node ipmi(%v)", h.reqId, err)
		return nil, err
	}

	return impi, nil
}

func (h *helper) setNodeIpmi() error {
	var err error
	h.ipmi.Password, err = password.Encrypt(h.ipmi.Password, base.SystemSeed)
	if err != nil {
		log.Errorf("nodes(%s): failed to encrypt ipmi password(%v)", h.reqId, err)
		return err
	}

	return h.mongo.UpdateOne(
		nodes.Db,
		nodes.CollectionIpmi,
		bson.M{"host": h.node},
		h.genUpsertPayload(),
		options.Update().SetUpsert(true),
	)
}

func (h *helper) ipmiOperateNode() error {
	info, err := h.getNodeIpmi()
	if err != nil {
		log.Errorf("nodes(%s): failed to get node ipmi(%v)", h.reqId, err)
		return err
	}

	decryptedPassword, err := password.Decrypt(info.Password, base.SystemSeed)
	if err != nil {
		log.Errorf("nodes(%s): failed to decrypt ipmi password(%v)", h.reqId, err)
		return err
	}

	op, err := h.getIpmiOperation()
	if err != nil {
		log.Errorf("nodes(%s): failed to get ipmi operation(%v)", h.reqId, err)
		return err
	}

	c, err := ipmi.NewHelper(
		ipmi.Host(info.Ip),
		ipmi.Port(info.Port),
		ipmi.Username(info.Username),
		ipmi.Password(decryptedPassword),
	)
	if err != nil {
		log.Errorf("nodes(%s): failed to create ipmi helper(%v)", h.reqId, err)
		return err
	}

	_, err = c.Operate(op)
	if err != nil {
		log.Errorf("nodes(%s): failed to operate ipmi(%v)", h.reqId, err)
		return err
	}

	go traceNodeStatus(h.node, h.operation)
	log.Infof(
		"nodes(%s): successfully operated ipmi(%s) on node(%s)",
		h.reqId,
		h.operation,
		info.Ip,
	)

	return nil
}

func (h *helper) disconnectNodeIpmi() error {
	return h.mongo.DeleteOne(
		nodes.Db,
		nodes.CollectionIpmi,
		bson.M{"host.name": h.node},
	)
}
