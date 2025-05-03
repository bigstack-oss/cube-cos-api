package triggers

import (
	"fmt"
	"strings"

	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/slack"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/trigger"
	cubelog "github.com/bigstack-oss/cube-cos-api/internal/log"
	"github.com/fsnotify/fsnotify"
	log "go-micro.dev/v5/logger"
)

func (o *Operator) initPolicyWatcher() error {
	var err error
	o.policy, err = fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	err = o.policy.Add("/etc/policies/alert_resp")
	if err != nil {
		return err
	}

	go o.watchChanges()
	return nil
}

func (o *Operator) watchChanges() {
	for {
		select {
		case event, ok := <-o.policy.Events:
			if ok {
				o.checkTriggers(event)
			}
		case err, ok := <-o.policy.Errors:
			if !ok {
				continue
			}
			if err != nil {
				log.Errorf("triggers: failed to fetch policy change event: %s", err.Error())
				continue
			}
		}
	}
}

func (o *Operator) checkTriggers(event fsnotify.Event) {
	if event.Name != trigger.ResponsePolicyV2 {
		return
	}

	if event.Has(fsnotify.Write) {
		cubelog.Throttle("triggers", fmt.Sprintf("%s changed, syncing triggers", event.Name))
		o.syncTriggers()
	}
}

func (o *Operator) syncTriggers() {
	triggers, err := cubecos.GetTriggers()
	if err != nil {
		log.Errorf("triggers: failed to sync triggers: %s", err.Error())
		return
	}

	apiTriggers := o.convertToApiTriggers(triggers)
	for i := range apiTriggers {
		syncSelectableResponseItems(&apiTriggers[i])
		syncAttrEnablement(&apiTriggers[i])
	}

	trigger.SyncList(apiTriggers)
}

func (o *Operator) convertToApiTriggers(triggers []trigger.CosOptions) []trigger.ApiOptions {
	apiTriggers := []trigger.ApiOptions{}

	for _, trigger := range triggers {
		apiTriggers = append(
			apiTriggers,
			trigger.ConvertToApiOptions(),
		)
	}

	return apiTriggers
}

func syncSelectableResponseItems(trigger *trigger.ApiOptions) {
	setEmailRecipientsToTrigger(trigger)
	if trigger.HasEmailRecipients() {
		trigger.Response.Types = append(
			trigger.Response.Types,
			"email",
		)
	}

	setSlackChannelsToTrigger(trigger)
	if trigger.HasSlackChannels() {
		trigger.Response.Types = append(
			trigger.Response.Types,
			"slack",
		)
	}
}

func setEmailRecipientsToTrigger(trigger *trigger.ApiOptions) {
	recipients, err := cubecos.GetEmailRecipients()
	if err != nil {
		return
	}

	for _, recipient := range recipients {
		if trigger.HasEmail(recipient.Address) {
			recipient.Enabled = true
			trigger.SetEmailDetails(recipient)
			continue
		}

		trigger.AppendEmail(recipient)
	}
}

func setSlackChannelsToTrigger(trigger *trigger.ApiOptions) {
	slacks, err := cubecos.GetSlackChannels()
	if err != nil {
		return
	}

	apiChannels := convertToApiChannels(slacks)
	for _, apiChannel := range apiChannels {
		if trigger.HasSlack(apiChannel.URL) {
			apiChannel.Enabled = true
			trigger.SetSlackDetails(apiChannel)
			continue
		}

		trigger.AppendSlack(apiChannel)
	}
}

func convertToApiChannels(channels []slack.CosChannel) []slack.ApiChannel {
	apiChannels := []slack.ApiChannel{}

	for _, channel := range channels {
		apiChannels = append(
			apiChannels,
			slack.ApiChannel{
				Name:        channel.Channel,
				URL:         channel.URL,
				Description: channel.Description,
			},
		)
	}

	return apiChannels
}

func syncAttrEnablement(options *trigger.ApiOptions) []trigger.Attribute {
	enabledAttrs := getEnabledAttrs(options)
	attributes := []trigger.Attribute{}
	for i, attr := range options.Attributes {
		for _, enabledAttr := range enabledAttrs {
			if attr.Name != enabledAttr.Name {
				continue
			}

			if attr.Value != enabledAttr.Value {
				continue
			}

			options.Attributes[i].Enabled = true
			break
		}
	}

	return attributes
}

func getEnabledAttrs(policyTrigger *trigger.ApiOptions) []trigger.Attribute {
	enabledAttrs := []trigger.Attribute{}
	matchRule := strings.ReplaceAll(policyTrigger.Match, `"`, ``)
	parts := strings.SplitSeq(matchRule, " OR ")
	for part := range parts {
		attrPair := strings.Split(part, " == ")
		if !isValidAttrPair(attrPair) {
			continue
		}

		enabledAttrs = append(
			enabledAttrs,
			trigger.Attribute{
				Name:  strings.TrimSpace(attrPair[0]),
				Value: strings.TrimSpace(attrPair[1]),
			},
		)
	}

	return enabledAttrs
}

func isValidAttrPair(attrPair []string) bool {
	return len(attrPair) == 2
}
