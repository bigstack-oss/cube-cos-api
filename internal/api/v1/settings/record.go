package settings

import (
	"maps"
	"net/http"

	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/email"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/setting"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/slack"
	log "go-micro.dev/v5/logger"
	"go.mongodb.org/mongo-driver/bson"
)

func (h *helper) eraseSenderPassword(senders *[]email.Sender) {
	for i := range *senders {
		(*senders)[i].ErasePassword()
	}
}

func (h *helper) updateClusterWiseSetting() {
	h.delegateToLocal()
	if h.isClusterWiseRequired {
		h.delegateToPeerControlNodes()
	}
}

func (h *helper) delegateToLocal() {
	h.addReqRecord(*h.task)
	reqQueue.Add(h.task)
}

func (h *helper) delegateToPeerControlNodes() {
	peerNodes, err := v1.GetPeerControlNodes()
	if err != nil {
		log.Errorf("settings: failed to get peer controller nodes: %v", err)
		return
	}

	for _, node := range peerNodes {
		h.updateSettingToPeerNode(node)
	}
}

func (h *helper) updateSettingToPeerNode(node v1.Node) {
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
	mapHeaderMap := v1.GenNodeAuthHeaders()

	headerMap := map[string]string{}
	for key, values := range headers {
		if len(values) > 0 {
			headerMap[key] = values[0]
		}
	}

	maps.Copy(headerMap, mapHeaderMap)
	return headerMap
}

func (h *helper) isSenderExist(host string) bool {
	policy, err := cubecos.GetAlertSetting()
	if err != nil {
		return false
	}

	return policy.HasSender(host)
}

func isRecipientExist(recipient string) bool {
	policy, err := cubecos.GetAlertSetting()
	if err != nil {
		return false
	}

	return policy.HasRecipient(recipient)
}

func (h *helper) isTitlePrefixUpdating() bool {
	count, err := h.mongo.GetCount(
		setting.DB,
		setting.ReqCollection,
		bson.M{"type": "titlePrefix"},
	)
	if err != nil {
		log.Errorf("settings: failed to check title prefix update: %v", err)
		return false
	}

	return count > 0
}

func (h *helper) isEmailRecipientUpdating(recipient *email.Recipient) bool {
	count, err := h.mongo.GetCount(
		setting.DB,
		setting.ReqCollection,
		bson.M{
			"type": "emailRecipient",
			"key":  recipient.Address,
		},
	)
	if err != nil {
		log.Errorf("settings: failed to get email count: %s", err.Error())
		return false
	}

	return count > 0
}

func (h *helper) isEmailSenderUpdating(sender *email.Sender) bool {
	count, err := h.mongo.GetCount(
		setting.DB,
		setting.ReqCollection,
		bson.M{
			"type": "emailSender",
			"key":  sender.Host,
		},
	)
	if err != nil {
		log.Errorf("settings: failed to get email count: %s", err.Error())
		return false
	}

	return count > 0
}

func (h *helper) isSlackUpdating(channel *slack.ApiChannel) bool {
	count, err := h.mongo.GetCount(
		setting.DB,
		setting.ReqCollection,
		bson.M{
			"type": "slackChannel",
			"key":  channel.URL,
		},
	)
	if err != nil {
		log.Errorf("settings: failed to get slack count: %s", err.Error())
		return false
	}

	return count > 0
}

func (h *helper) updateSettingTask() error {
	return h.mongo.DeleteOne(
		setting.DB,
		setting.ReqCollection,
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
		setting.DB,
		email.SenderCollection,
		bson.M{"host": h.c.Param("senderHost")},
		bson.M{"$set": bson.M{"accessVerified": false}},
	)
}
