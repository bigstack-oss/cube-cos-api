package settings

import (
	"context"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/email"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/settings"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/slack"
	log "go-micro.dev/v5/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func (h *helper) syncUpdatingStatus(setting *settings.Api) {
	setting.SetOkStatus()
	h.syncUpdatingTitlePrefix(&setting.TitlePrefix)
	h.syncUpdatingSender(&setting.Email.Senders)
	h.syncUpdatingRecipient(&setting.Email.Recipients)
	h.syncUpdatingSlack(&setting.Slack)
}

func (h *helper) syncUpdatingTitlePrefix(titlePrefix *settings.TitlePrefix) {
	if h.isTitlePrefixUpdating() {
		h.syncUpdatingTitlePrefixValue(titlePrefix)
		titlePrefix.SetUpdating()
	}
}

func (h *helper) syncUpdatingTitlePrefixValue(titlePrefix *settings.TitlePrefix) {
	c, err := h.mongo.GetQueryCursor(
		settings.DB,
		settings.ReqCollection,
		bson.M{"type": "titlePrefix"},
	)
	if err != nil {
		log.Errorf("settings(%s): failed to update value from title prefix cursor(%v)", h.reqId, err)
		return
	}

	ctx, cancel := context.WithTimeout(wait.CtxSeconds(5))
	defer cancel()
	defer c.Close(ctx)
	h.setUpdatingTitlePrefixValue(c, titlePrefix)
}

func (h *helper) setUpdatingTitlePrefixValue(c *mongo.Cursor, titlePrefix *settings.TitlePrefix) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(5))
	defer cancel()
	for c.Next(ctx) {
		req := &settings.Setting{}
		err := c.Decode(req)
		if err != nil {
			log.Errorf("settings(%s): failed to decode title prefix cursor (%v)", h.reqId, err)
			continue
		}

		titlePrefix.Value = req.Value.(string)
		break
	}
}

func (h *helper) syncUpdatingSender(senders *[]email.Sender) {
	if len(*senders) == 0 {
		h.syncCreatingSenderIfExists(senders)
	}

	for i, sender := range *senders {
		if h.isSenderUpdating(sender) {
			(*senders)[i] = *h.syncUpdatingSenderValue(sender)
			(*senders)[i].SetUpdating()
		}
	}
}

func (h *helper) syncCreatingSenderIfExists(senders *[]email.Sender) {
	c, err := h.mongo.GetQueryCursor(
		settings.DB,
		settings.ReqCollection,
		bson.M{"type": "emailSender"},
	)
	if err != nil {
		log.Errorf("settings: failed to update value from email sender cursor (%v)", err)
		return
	}

	ctx, cancel := context.WithTimeout(wait.CtxSeconds(5))
	defer cancel()
	defer c.Close(ctx)
	h.setUpdatingEmailSenders(c, senders)
}

func (h *helper) setUpdatingEmailSenders(c *mongo.Cursor, senders *[]email.Sender) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(5))
	defer cancel()
	for c.Next(ctx) {
		req := &settings.Setting{}
		err := c.Decode(req)
		if err != nil {
			log.Errorf("settings: failed to decode email sender cursor (%v)", err)
			continue
		}

		(*senders) = append(*senders, *req.Sender)
	}
}

func (h *helper) syncUpdatingSenderValue(sender email.Sender) *email.Sender {
	c, err := h.mongo.GetQueryCursor(
		settings.DB,
		settings.ReqCollection,
		bson.M{"type": "emailSender", "key": sender.Host},
	)
	if err != nil {
		log.Errorf("settings: failed to update value from email sender cursor (%v)", err)
		return nil
	}

	ctx, cancel := context.WithTimeout(wait.CtxSeconds(5))
	defer cancel()
	defer c.Close(ctx)
	return h.parseEmailSenderUpdateValue(c)
}

func (h *helper) parseEmailSenderUpdateValue(c *mongo.Cursor) *email.Sender {
	req := &settings.Setting{}
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(5))
	defer cancel()
	for c.Next(ctx) {
		err := c.Decode(req)
		if err != nil {
			log.Errorf("settings: failed to decode email sender cursor (%v)", err)
			return nil
		}
	}

	return req.Sender
}

func (h *helper) syncUpdatingRecipient(recipients *[]email.Recipient) {
	h.syncRecipientRecords(recipients)

	for i, recipient := range *recipients {
		if h.isRecipientUpdating(&recipient) {
			val := h.syncUpdatingRecipientValue(&recipient)
			if val == nil {
				continue
			}

			(*recipients)[i] = *val
			(*recipients)[i].SetUpdating()
		}
	}

	dedupRecipients(recipients)
}

func (h *helper) syncRecipientRecords(recipients *[]email.Recipient) {
	c, err := h.mongo.GetQueryCursor(
		settings.DB,
		settings.ReqCollection,
		bson.M{"type": "emailRecipient"},
	)
	if err != nil {
		log.Errorf("settings: failed to update value from email recipient cursor (%v)", err)
		return
	}

	ctx, cancel := context.WithTimeout(wait.CtxSeconds(5))
	defer cancel()
	defer c.Close(ctx)
	h.parseUpdatingRecipients(c, recipients)
	h.dedupUpdatingRecipients(recipients)
}

func dedupRecipients(recipients *[]email.Recipient) {
	uniqs := map[string]email.Recipient{}
	for _, recipient := range *recipients {
		prev, found := uniqs[recipient.Address]
		if !found {
			uniqs[recipient.Address] = recipient
			continue
		}

		if !prev.Status.IsUpdating {
			uniqs[recipient.Address] = recipient
		}
	}

	*recipients = []email.Recipient{}
	for _, recipient := range uniqs {
		*recipients = append(*recipients, recipient)
	}
}

func (h *helper) parseUpdatingRecipients(c *mongo.Cursor, recipients *[]email.Recipient) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(5))
	defer cancel()
	for c.Next(ctx) {
		req := &settings.Setting{}
		err := c.Decode(req)
		if err != nil {
			log.Errorf("settings(%s): failed to decode recipient cursor (%v)", h.reqId, err)
			continue
		}

		*recipients = append(*recipients, *req.Recipient)
	}
}

func (h *helper) syncUpdatingRecipientValue(recipient *email.Recipient) *email.Recipient {
	c, err := h.mongo.GetQueryCursor(
		settings.DB,
		settings.ReqCollection,
		bson.M{"type": "emailRecipient", "key": recipient.Address},
	)
	if err != nil {
		log.Errorf("settings(%s): failed to get update value from recipient cursor (%v)", h.reqId, err)
		return nil
	}

	ctx, cancel := context.WithTimeout(wait.CtxSeconds(5))
	defer cancel()
	defer c.Close(ctx)
	return h.parseUpdatingRecipientValue(c)
}

func (h *helper) parseUpdatingRecipientValue(c *mongo.Cursor) *email.Recipient {
	req := &settings.Setting{}
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(5))
	defer cancel()
	for c.Next(ctx) {
		err := c.Decode(req)
		if err != nil {
			log.Errorf("settings(%s): failed to decode recipient cursor (%v)", h.reqId, err)
			return nil
		}
	}

	return req.Recipient
}

func (h *helper) syncUpdatingSlack(opt *slack.Options) {
	if len(opt.Channels) == 0 {
		opt.Channels = []slack.ApiChannel{}
	}

	h.syncUpdatingSlacks(&opt.Channels)
	for i, channel := range opt.Channels {
		if !h.isSlackUpdating(&channel) {
			continue
		}

		val := h.syncUpdatingSlackValue(&channel)
		if val == nil {
			continue
		}

		opt.Channels[i] = *val
		opt.Channels[i].SetUpdating()
	}

	// note:
	// in the M2, we might consider to let url or name either one to be uneditable
	// otherwise, the tracking logic will be complicated
	dedupNameOrUrlUpdatingSlacks(opt)
}

func (h *helper) syncUpdatingSlacks(channels *[]slack.ApiChannel) {
	c, err := h.mongo.GetQueryCursor(
		settings.DB,
		settings.ReqCollection,
		bson.M{"type": "slackChannel"},
	)
	if err != nil {
		log.Errorf("settings(%s): failed to get update value from slack cursor(%v)", h.reqId, err)
		return
	}

	ctx, cancel := context.WithTimeout(wait.CtxSeconds(5))
	defer cancel()
	defer c.Close(ctx)
	h.parseUpdatingSlacks(c, channels)
	h.dedupUpdatingSlacks(channels)
}

func (h *helper) parseUpdatingSlacks(c *mongo.Cursor, slack *[]slack.ApiChannel) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(5))
	defer cancel()
	for c.Next(ctx) {
		req := &settings.Setting{}
		err := c.Decode(req)
		if err != nil {
			log.Errorf("settings: failed to decode slack channel cursor (%v)", err)
			continue
		}

		*slack = append(*slack, *req.Slack)
	}
}

func (h *helper) dedupUpdatingSlacks(channels *[]slack.ApiChannel) {
	uniqs := map[string]slack.ApiChannel{}
	for _, channel := range *channels {
		uniqs[channel.URL] = channel
	}

	*channels = []slack.ApiChannel{}
	for _, channel := range uniqs {
		*channels = append(*channels, channel)
	}
}

func dedupNameOrUrlUpdatingSlacks(opts *slack.Options) {
	uniqs := map[string]slack.ApiChannel{}
	for _, channel := range opts.Channels {
		prev, found := uniqs[channel.URL]
		if !found {
			uniqs[channel.URL] = channel
			continue
		}

		if !prev.Status.IsUpdating {
			uniqs[channel.URL] = channel
		}
	}

	opts.Channels = []slack.ApiChannel{}
	for _, channel := range uniqs {
		opts.Channels = append(opts.Channels, channel)
	}
}

func (h *helper) syncUpdatingSlackValue(channel *slack.ApiChannel) *slack.ApiChannel {
	c, err := h.mongo.GetQueryCursor(
		settings.DB,
		settings.ReqCollection,
		bson.M{"type": "slackChannel", "key": channel.URL},
	)
	if err != nil {
		log.Errorf("settings: failed to update value from slack channel cursor (%v)", err)
		return nil
	}

	ctx, cancel := context.WithTimeout(wait.CtxSeconds(5))
	defer cancel()
	defer c.Close(ctx)
	return h.parseUpdatingSlackValue(c)
}

func (h *helper) parseUpdatingSlackValue(c *mongo.Cursor) *slack.ApiChannel {
	req := &settings.Setting{}
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(5))
	defer cancel()
	for c.Next(ctx) {
		err := c.Decode(req)
		if err != nil {
			log.Errorf("settings: failed to decode slack channel cursor (%v)", err)
			return nil
		}
	}

	return req.Slack
}

func (h *helper) dedupUpdatingRecipients(recipients *[]email.Recipient) {
	uniqs := map[string]email.Recipient{}
	for _, recipient := range *recipients {
		uniqs[recipient.Address] = recipient
	}

	*recipients = []email.Recipient{}
	for _, recipient := range uniqs {
		*recipients = append(*recipients, recipient)
	}
}
