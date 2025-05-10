package triggers

import (
	"maps"
	"net/http"

	cubeMongo "github.com/bigstack-oss/bigstack-dependency-go/pkg/mongo"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/trigger"
	log "go-micro.dev/v5/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (h *helper) addReqRecord() {
	mongo := cubeMongo.GetGlobalHelper()
	err := mongo.UpdateOne(
		trigger.DB,
		trigger.ReqCollection,
		bson.M{"name": h.trigger.Name},
		bson.M{"$set": h.trigger},
		options.Update().SetUpsert(true),
	)
	if err != nil {
		log.Errorf(
			"triggers: failed to sync trigger record for %s (%s)",
			h.trigger.Name,
			err.Error(),
		)
	}
}

func (h *helper) updateTaskStatus() error {
	mongo := cubeMongo.GetGlobalHelper()
	return mongo.DeleteOne(
		trigger.DB,
		trigger.ReqCollection,
		bson.M{"name": h.trigger.Name},
	)
}

func (h *helper) hasUpdateHistory(t trigger.ApiOptions) bool {
	mongo := cubeMongo.GetGlobalHelper()
	count, err := mongo.GetCount(
		trigger.DB,
		trigger.ReqCollection,
		bson.M{"name": t.Name},
	)
	if err != nil {
		return false
	}

	return count > 0
}

func (h *helper) getUpdateRecord(t trigger.ApiOptions) (*trigger.ApiOptions, error) {
	mongo := cubeMongo.GetGlobalHelper()
	pending, err := mongo.Get(
		trigger.DB,
		trigger.ReqCollection,
		bson.M{"name": t.Name},
	)
	if err != nil {
		return nil, err
	}

	record := &trigger.ApiOptions{}
	err = pending.Decode(record)
	if err != nil {
		return nil, err
	}

	return record, nil
}

func (h *helper) updateClusterWiseTrigger() {
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
		log.Errorf("triggers: failed to get peer controller nodes: %v", err)
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
		log.Errorf("triggers: failed to update trigger to peer node %s: %v", node.Hostname, err)
		return
	}

	if resp.IsError() {
		log.Errorf("triggers: has resp error during updating trigger to peer node %s: %s", node.Hostname, resp.String())
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
