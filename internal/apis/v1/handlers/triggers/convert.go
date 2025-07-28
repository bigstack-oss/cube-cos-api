package triggers

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/events"
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
	emails := []Email{}
	for _, email := range trigger.Emails {
		emails = append(
			emails, Email{
				Address: email.Address,
				Note:    email.Note,
			},
		)
	}

	slacks := []Slack{}
	for _, slack := range trigger.Slacks {
		slacks = append(
			slacks, Slack{
				Name:        slack.Channel,
				Url:         slack.URL,
				Description: slack.Description,
			},
		)
	}

	script := triggers.Script{}
	for _, shell := range trigger.Execs.Shells {
		file, err := os.ReadFile(filepath.Join("/var/response", fmt.Sprintf("%s.shell", shell)))
		if err != nil {
			log.Errorf("triggers(%s): failed to read script file %s(%v)", h.reqId, shell, err)
			continue
		}

		filename := strings.ReplaceAll(shell, ".shell", "")
		script.Name = filename
		script.Content = string(file)
	}

	return Response{
		Script: script,
		Emails: emails,
		Slacks: slacks,
	}
}
