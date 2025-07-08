package triggers

import (
	"fmt"

	conf "github.com/bigstack-oss/cube-cos-api/internal/config"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/slack"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/triggers"
	bslog "github.com/bigstack-oss/cube-cos-api/internal/log"
	"github.com/fsnotify/fsnotify"
	log "go-micro.dev/v5/logger"
)

func (o *Operator) initPolicyWatcher() error {
	var err error
	o.policy, err = fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	err = o.policy.Add("/etc")
	if err != nil {
		return err
	}

	o.syncTriggers()
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
				log.Errorf("triggers: failed to fetch policy change event(%v)", err)
				continue
			}
		}
	}
}

func (o *Operator) checkTriggers(event fsnotify.Event) {
	if event.Name != conf.Opts.Spec.Identity.Policy {
		return
	}

	if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) {
		bslog.Throttle("triggers", fmt.Sprintf("%s changed, syncing triggers", event.Name))
		o.syncTriggers()
	}
}

func (o *Operator) syncTriggers() {
	list, err := cubecos.GetTriggers()
	if err != nil {
		log.Errorf("triggers: failed to sync triggers(%v)", err)
		return
	}

	apiTriggers := o.toApiSchemas(list)
	for i := range apiTriggers {
		syncTriggerDetails(&apiTriggers[i])
		syncSelectableResponseItems(&apiTriggers[i])
	}

	triggers.SyncList(apiTriggers)
}

func (o *Operator) toApiSchemas(list []triggers.CosSchema) []triggers.ApiSchema {
	apiTriggers := []triggers.ApiSchema{}

	for _, trigger := range list {
		apiTriggers = append(
			apiTriggers,
			trigger.ToApiSchema(),
		)
	}

	return apiTriggers
}

func syncTriggerDetails(apiTrigger *triggers.ApiSchema) {
	triggerMap := triggers.GetDetailsMap()
	details, found := triggerMap[apiTrigger.Name]
	if found {
		apiTrigger.Name = details.Name
		apiTrigger.IsBuiltIn = details.IsBuiltIn
	}
}

func syncSelectableResponseItems(trigger *triggers.ApiSchema) {
	trigger.Response.Types = []string{}
	setEmailRecipientsToTrigger(trigger)
	if trigger.HasEmailRecipients() {
		setEmailTypeIfAtLeastOneEmailEnabled(trigger)
	}

	setSlackChannelsToTrigger(trigger)
	if trigger.HasSlackChannels() {
		setSlackTypeIfAtLeastOneChannelEnabled(trigger)
	}
}

func setEmailTypeIfAtLeastOneEmailEnabled(trigger *triggers.ApiSchema) {
	for _, recipient := range trigger.Response.Emails {
		if recipient.Enabled {
			trigger.Response.Types = append(trigger.Response.Types, "email")
			break
		}
	}
}

func setSlackTypeIfAtLeastOneChannelEnabled(trigger *triggers.ApiSchema) {
	for _, channel := range trigger.Response.Slacks {
		if channel.Enabled {
			trigger.Response.Types = append(trigger.Response.Types, "slack")
			break
		}
	}
}

func setEmailRecipientsToTrigger(trigger *triggers.ApiSchema) {
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

func setSlackChannelsToTrigger(trigger *triggers.ApiSchema) {
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
