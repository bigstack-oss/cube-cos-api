package triggers

import (
	"maps"
	"net/http"

	bsmongo "github.com/bigstack-oss/bigstack-dependency-go/pkg/mongo"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/triggers"
	log "go-micro.dev/v5/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (h *helper) addReqRecord() {
	mongo := bsmongo.GetGlobalHelper()
	err := mongo.UpdateOne(
		triggers.DB,
		triggers.ReqCollection,
		bson.M{"name": h.trigger.Name},
		bson.M{"$set": h.trigger},
		options.Update().SetUpsert(true),
	)
	if err != nil {
		log.Errorf(
			"triggers(%s): failed to sync trigger record for %s (%s)",
			h.reqId,
			h.trigger.Name,
			err.Error(),
		)
	}
}

func (h *helper) updateTaskStatus() error {
	mongo := bsmongo.GetGlobalHelper()
	return mongo.DeleteOne(
		triggers.DB,
		triggers.ReqCollection,
		bson.M{"name": h.trigger.Name},
	)
}

func (h *helper) hasUpdateHistory(t triggers.ApiSchema) bool {
	mongo := bsmongo.GetGlobalHelper()
	count, err := mongo.GetCount(
		triggers.DB,
		triggers.ReqCollection,
		bson.M{"name": t.Name},
	)
	if err != nil {
		return false
	}

	return count > 0
}

func (h *helper) getUpdateRecord(trigger triggers.ApiSchema) (*triggers.ApiSchema, error) {
	mongo := bsmongo.GetGlobalHelper()
	record, err := mongo.Get(
		triggers.DB,
		triggers.ReqCollection,
		bson.M{"name": trigger.Name},
	)
	if err != nil {
		return nil, err
	}

	schema := &triggers.ApiSchema{}
	err = record.Decode(schema)
	if err != nil {
		return nil, err
	}

	return schema, nil
}

func (h *helper) updateToAllControllers() {
	h.delegateToLocal()
	if h.isClusterWiseRequired {
		h.delegateToPeerControlNodes()
	}
}

func (h *helper) delegateToLocal() {
	if h.isClusterWiseRequired {
		h.addReqRecord()
	}

	reqQueue.Add(&h.trigger)
}

func (h *helper) delegateToPeerControlNodes() {
	peerNodes, err := nodes.GetPeerControls()
	if err != nil {
		log.Errorf("triggers(%s): failed to get peer controller nodes: %v", h.reqId, err)
		return
	}

	for _, node := range peerNodes {
		h.updateTriggerToPeerNode(node)
	}
}

func (h *helper) updateTriggerToPeerNode(node nodes.Node) {
	req := h.http.R().
		SetHeaders(h.convertHeadersToMap(h.c.Request.Header)).
		SetQueryParam("clusterWise", "false").
		SetBody(string(h.rawBody))

	url := node.GenUrl() + h.c.Request.RequestURI
	resp, err := req.Execute(h.c.Request.Method, url)
	if err != nil {
		log.Errorf(
			"triggers(%s): failed to update trigger to peer node %s: %s",
			h.reqId,
			node.Hostname,
			err.Error(),
		)
		return
	}

	if resp.IsError() {
		log.Errorf(
			"triggers(%s): has resp error during updating trigger to peer node %s: %s",
			h.reqId,
			node.Hostname,
			resp.String(),
		)
		return
	}
}

func (h *helper) convertHeadersToMap(headers http.Header) map[string]string {
	headerMap := map[string]string{}
	for key, values := range headers {
		if len(values) > 0 {
			headerMap[key] = values[0]
		}
	}

	maps.Copy(headerMap, nodes.GetSecretHeaders())
	return headerMap
}
