package triggers

import (
	"context"
	"encoding/json"
	"maps"
	"net/http"

	bsmongo "github.com/bigstack-oss/bigstack-dependency-go/pkg/mongo"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/base"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/triggers"
	"github.com/mohae/deepcopy"
	log "go-micro.dev/v5/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (h *helper) addReqRecord() {
	mongo := bsmongo.GetGlobalHelper()
	err := mongo.UpdateOne(
		triggers.DB,
		triggers.ReqCollection,
		bson.M{"name": h.reqOpts.Name},
		bson.M{"$set": h.reqOpts},
		options.Update().SetUpsert(true),
	)
	if err != nil {
		log.Errorf(
			"triggers(%s): failed to sync trigger %s record(%v)",
			h.reqId,
			h.reqOpts.Name,
			err,
		)
	}
}

func (h *helper) updateTaskStatus() error {
	mongo := bsmongo.GetGlobalHelper()
	return mongo.DeleteOne(
		triggers.DB,
		triggers.ReqCollection,
		bson.M{"name": h.reqOpts.Name},
	)
}

func (h *helper) hasUpdateHistory(trigger triggerResp) bool {
	mongo := bsmongo.GetGlobalHelper()
	count, err := mongo.GetCount(
		triggers.DB,
		triggers.ReqCollection,
		bson.M{"name": trigger.Name},
	)
	if err != nil {
		return false
	}

	return count > 0
}

func (h *helper) getUpdateRecord(trigger triggerResp) (*triggerResp, error) {
	mongo := bsmongo.GetGlobalHelper()
	record, err := mongo.Get(
		triggers.DB,
		triggers.ReqCollection,
		bson.M{"name": trigger.Name},
	)
	if err != nil {
		return nil, err
	}

	resp := &triggerResp{}
	err = record.Decode(resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (h *helper) updateToControllers() {
	h.updateToLocal()
	h.updateToPeerTriggers()
}

func (h *helper) updateToLocal() {
	if cubecos.IsVirtualIpOwner(base.Hostname) {
		h.addReqRecord()
	}

	reqQueue.Add(&h.reqOpts)
}

func (h *helper) updateToPeerTriggers() {
	if h.isRequestFromVirtualIp() {
		return
	}

	nodes, err := nodes.GetPeerControls()
	if err != nil {
		log.Errorf("triggers(%s): failed to get peer controller nodes: %v", h.reqId, err)
		return
	}

	for _, node := range nodes {
		h.updatePeerTrigger(node)
	}
}

func (h *helper) updatePeerTrigger(node nodes.Node) error {
	reqOpts, err := h.genPeerTriggerReq(node.Hostname)
	if err != nil {
		return nil
	}

	url := node.PostTriggerUrl()
	req := h.http.R().
		SetHeaders(h.convertHeadersToMap(h.c.Request.Header)).
		SetBody(string(reqOpts))
	resp, err := req.Execute(h.c.Request.Method, url)
	if err != nil {
		log.Errorf(
			"triggers(%s): failed to update peer trigger %s(%v)",
			h.reqId, node.Hostname, err,
		)
		return err
	}

	if resp.IsError() {
		log.Errorf(
			"triggers(%s): has resp error during updating peer trigger on node %s(%s)",
			h.reqId, node.Hostname, resp.String(),
		)
		return err
	}

	return nil
}

func (h *helper) genPeerTriggerReq(hostname string) ([]byte, error) {
	reqOpts := deepcopy.Copy(h.reqOpts).(triggers.ReqOpts)
	reqOpts.Nodes = append(reqOpts.Nodes, hostname)
	req, err := json.Marshal(reqOpts)
	if err != nil {
		log.Errorf("triggers(%s): failed to marshal trigger request for node %s(%v)", h.reqId, hostname, err)
		return nil, err
	}

	return req, nil
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

func (h *helper) addCreatingTriggers(list *[]triggerResp) {
	if !h.hasCreatingTriggers() {
		return
	}

	creatings, err := h.getCreatingTriggers()
	if err != nil {
		return
	}

	for _, creating := range creatings {
		*list = append(*list, triggerResp{
			Name:        creating.Name,
			Description: creating.Description,
			Enabled:     creating.Enabled,
			Response:    Response{Types: h.getCreatingResponseTypes(creating)},
			Status:      &creating.Status,
		})
	}
}

func (h *helper) hasCreatingTriggers() bool {
	count, err := h.mongo.GetCount(
		triggers.DB,
		triggers.ReqCollection,
		bson.M{"status.current": status.Creating},
	)
	if err != nil {
		log.Errorf("triggers(%s): failed to get creating triggers count: %v", h.reqId, err)
		return false
	}

	return count > 0
}

func (h *helper) getCreatingTriggers() ([]triggers.ReqOpts, error) {
	cursor, err := h.mongo.GetQueryCursor(
		triggers.DB,
		triggers.ReqCollection,
		bson.M{"status.current": status.Creating},
	)
	if err != nil {
		log.Errorf("triggers(%s): failed to get creating triggers(%v)", h.reqId, err)
		return nil, err
	}

	ctx, cancel := context.WithTimeout(wait.CtxSeconds(30))
	defer cancel()
	defer cursor.Close(ctx)
	return h.parseCreatingTriggers(cursor)
}

func (h *helper) parseCreatingTriggers(c *mongo.Cursor) ([]triggers.ReqOpts, error) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(30))
	defer cancel()
	reqs := []triggers.ReqOpts{}

	for c.Next(ctx) {
		reqOpts := &triggers.ReqOpts{}
		err := c.Decode(reqOpts)
		if err != nil {
			log.Warnf("triggers(%s): failed to decode creating trigger record(%v)", h.reqId, err)
			continue
		}

		reqs = append(reqs, *reqOpts)
	}

	err := c.Err()
	if err != nil {
		log.Errorf("triggers(%s): error while iterating creating triggers(%v)", h.reqId, err)
		return nil, err
	}

	return reqs, nil
}

func (h *helper) getCreatingResponseTypes(record triggers.ReqOpts) []string {
	types := []string{}
	if len(record.Response.Emails) > 0 {
		types = append(types, "email")
	}

	if len(record.Response.Slacks) > 0 {
		types = append(types, "slack")
	}

	if record.Response.Script.Name != "" {
		types = append(types, "script")
	}

	return types
}
