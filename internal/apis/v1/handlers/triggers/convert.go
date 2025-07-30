package triggers

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/email"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/events"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/settings"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/slack"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/triggers"
	log "go-micro.dev/v5/logger"
)

func (h *helper) GetAlertTypes(events []events.Event) ([]string, error) {
	alertTypes := []string{}
	dedup := map[string]struct{}{}

	for _, event := range events {
		if event.Type == "" {
			continue
		}

		_, exists := dedup[event.Type]
		if exists {
			continue
		}

		dedup[event.Type] = struct{}{}
		alertTypes = append(alertTypes, event.Type)
	}

	return alertTypes, nil
}

func (h *helper) GetSeverities(events []events.Event) ([]string, error) {
	severities := []string{}
	dedup := map[string]struct{}{}

	for _, event := range events {
		if event.Severity == "" {
			continue
		}

		_, exists := dedup[event.Severity]
		if exists {
			continue
		}

		dedup[event.Severity] = struct{}{}
		severities = append(severities, event.Severity)
	}

	return severities, nil
}

func (h *helper) GetCategories(events []events.Event) ([]string, error) {
	categories := []string{}
	dedup := map[string]struct{}{}

	for _, event := range events {
		if event.Category == "" {
			continue
		}

		_, exists := dedup[event.Category]
		if exists {
			continue
		}

		dedup[event.Category] = struct{}{}
		categories = append(categories, event.Category)
	}

	return categories, nil
}

func (h *helper) GetEventIds(events []events.Event) ([]string, error) {
	eventIds := []string{}
	dedup := map[string]struct{}{}

	for _, event := range events {
		if event.Id == "" {
			continue
		}

		_, exists := dedup[event.Id]
		if exists {
			continue
		}

		dedup[event.Id] = struct{}{}
		eventIds = append(eventIds, event.Id)
	}

	return eventIds, nil
}

func (h *helper) convertTrigger(trigger triggers.Trigger) triggerResp {
	return triggerResp{
		Name:        trigger.Name,
		Description: trigger.Description,
		Attribute:   trigger.ParseAttributes(),
		Response:    h.convertResponse(trigger),
		Enabled:     trigger.Enabled,
		Status:      &status.Trigger{},
		Topic:       trigger.Topic,
	}
}

func (h *helper) convertResponse(trigger triggers.Trigger) Response {
	settings, err := cubecos.GetAlertSetting()
	if err != nil {
		log.Warnf("triggers(%s): failed to get alert settings(%v)", h.reqId, err)
	}

	return Response{
		Script: h.parseScriptDetails(trigger),
		Emails: h.parseEmailDetails(settings, trigger.Emails),
		Slacks: h.parseSlackDetails(settings, trigger.Slacks),
	}
}

func (h *helper) convertToEmailRecipients(emails []string) []email.Recipient {
	recipients := []email.Recipient{}
	for _, e := range emails {
		recipients = append(recipients, email.Recipient{Address: e})
	}

	return recipients
}

func (h *helper) convertToSlackChannels(slacks []string) []slack.CosChannel {
	channels := []slack.CosChannel{}
	for _, s := range slacks {
		channels = append(channels, slack.CosChannel{URL: s})
	}

	return channels
}

func (h *helper) parseEmailDetails(settings *settings.Cos, emails []email.Recipient) []Email {
	list := []Email{}
	for _, email := range emails {
		e := Email{Address: email.Address}
		if settings == nil {
			list = append(list, e)
			continue
		}

		email, found := settings.GetEmail(email.Address)
		if found {
			e.Note = email.Note
			list = append(list, e)
		}
	}

	return list
}

func (h *helper) parseSlackDetails(settings *settings.Cos, slacks []slack.CosChannel) []Slack {
	list := []Slack{}
	for _, slack := range slacks {
		s := Slack{Url: slack.URL}
		if settings == nil {
			list = append(list, s)
			continue
		}

		slack, found := settings.GetSlack(s.Url)
		if found {
			s.Name = slack.Channel
			s.Description = slack.Description
			list = append(list, s)
		}
	}

	return list
}

func (h *helper) parseScriptDetails(trigger triggers.Trigger) triggers.Script {
	script := triggers.Script{}
	for _, shell := range trigger.Execs.Shells {
		path := filepath.Join(settings.ScriptDir, fmt.Sprintf("%s.shell", shell.Name))
		file, err := os.ReadFile(path)
		if err != nil {
			log.Errorf("triggers(%s): failed to read script file %s(%v)", h.reqId, shell.Name, err)
			continue
		}

		script.Name = shell.Name
		script.Content = string(file)
	}

	return script
}
