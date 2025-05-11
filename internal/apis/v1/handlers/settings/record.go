package settings

import (
	"maps"
	"net/http"

	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/email"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/settings"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/slack"
	log "go-micro.dev/v5/logger"
	"go.mongodb.org/mongo-driver/bson"
)

func (h *helper) updateEmailSenderRecord() error {
	return h.mongo.UpdateOne(
		settings.DB,
		email.SenderCollection,
		bson.M{"host": h.emailSender},
		bson.M{
			"$set": bson.M{
				"host":           h.task.Sender.Host,
				"port":           h.task.Sender.Port,
				"username":       h.task.Sender.Username,
				"password":       h.task.Sender.Password,
				"from":           h.task.Sender.From,
				"accessVerified": h.task.Sender.AccessVerified,
			},
		},
	)
}

func (h *helper) hideSenderPassword(senders *[]email.Sender) {
	for i := range *senders {
		(*senders)[i].ErasePassword()
	}
}

func (h *helper) updateLocal() {
	h.addReqRecord(*h.task)
	reqQueue.Add(h.task)
}

func (h *helper) updatePeerControllers() {
	peerNodes, err := nodes.GetPeerControls()
	if err != nil {
		log.Errorf("settings: failed to get peer controller nodes: %v", err)
		return
	}

	for _, node := range peerNodes {
		h.updateSettingToPeerNode(node)
	}
}

func (h *helper) updateSettingToPeerNode(node nodes.Node) {
	req := h.http.R().
		SetHeaders(h.convertHeadersToMap(h.c.Request.Header)).
		SetQueryParam("clusterWise", "false").
		SetBody(string(h.rawBody))

	url := node.GenUrl() + h.c.Request.RequestURI
	resp, err := req.Execute(h.c.Request.Method, url)
	if err != nil {
		log.Errorf("settings: failed to update setting to peer node %s: %v", node.Hostname, err)
		return
	}

	if resp.IsError() {
		log.Errorf("settings: has resp error during updating setting to peer node %s: %s", node.Hostname, resp.String())
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

func (h *helper) isSenderExist(host string) bool {
	policy, err := cubecos.GetAlertSetting()
	if err != nil {
		return false
	}

	return policy.HasSender(host)
}

func (h *helper) isTitlePrefixUpdating() bool {
	count, err := h.mongo.GetCount(
		settings.DB,
		settings.ReqCollection,
		bson.M{"type": "titlePrefix"},
	)
	if err != nil {
		log.Errorf("settings(%s): failed to check title prefix update: %v", h.reqId, err)
		return false
	}

	return count > 0
}

func (h *helper) isRecipientUpdating(recipient *email.Recipient) bool {
	count, err := h.mongo.GetCount(
		settings.DB,
		settings.ReqCollection,
		bson.M{
			"type": "emailRecipient",
			"key":  recipient.Address,
		},
	)
	if err != nil {
		log.Errorf("settings(%s): failed to get sender count: %s", h.reqId, err.Error())
		return false
	}

	return count > 0
}

func (h *helper) isSenderUpdating(sender email.Sender) bool {
	count, err := h.mongo.GetCount(
		settings.DB,
		settings.ReqCollection,
		bson.M{
			"type": "emailSender",
			"key":  sender.Host,
		},
	)
	if err != nil {
		log.Errorf("settings(%s): failed to get sender count: %s", h.reqId, err.Error())
		return false
	}

	return count > 0
}

func (h *helper) isSlackUpdating(channel *slack.ApiChannel) bool {
	count, err := h.mongo.GetCount(
		settings.DB,
		settings.ReqCollection,
		bson.M{
			"type": "slackChannel",
			"key":  channel.URL,
		},
	)
	if err != nil {
		log.Errorf("settings(%s): failed to get slack channel count: %s", h.reqId, err.Error())
		return false
	}

	return count > 0
}

func (h *helper) updateSettingTask() error {
	return h.mongo.DeleteOne(
		settings.DB,
		settings.ReqCollection,
		h.genTaskFilter(),
	)
}

func (h *helper) genTaskFilter() bson.M {
	return bson.M{
		"type": h.task.Type,
		"key":  h.task.Key,
	}
}

func (h *helper) resetAccessVerification() error {
	return h.mongo.UpdateMany(
		settings.DB,
		email.SenderCollection,
		bson.M{"host": h.c.Param("senderHost")},
		bson.M{"$set": bson.M{"accessVerified": false}},
	)
}
