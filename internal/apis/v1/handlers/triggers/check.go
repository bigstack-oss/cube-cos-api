package triggers

import (
	"encoding/base64"
	"fmt"
	"slices"
	"strings"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/triggers"
)

func (h *helper) isTriggerExist(trigger string) bool {
	for _, t := range triggers.List() {
		if t.Name == trigger {
			return true
		}
	}

	return false
}

func (h *helper) checkMaterials() error {
	err := h.checkAttributes()
	if err != nil {
		return err
	}

	err = h.checkResponse()
	if err != nil {
		return err
	}

	return nil
}

func (h *helper) checkAttributes() error {
	if h.allAttributesAreEmpty() {
		return fmt.Errorf("trigger %s has no attributes", h.applyOpts.Name)
	}

	var err error
	h.materials, err = h.listMaterials()
	if err != nil {
		return err
	}

	err = h.hasInvalidEventIds()
	if err != nil {
		return nil
	}

	err = h.hasInvaliadAlertTypes()
	if err != nil {
		return err
	}

	err = h.hasInvalidSeverities()
	if err != nil {
		return err
	}

	err = h.hasInvalidCategories()
	if err != nil {
		return err
	}

	return nil
}

func (h *helper) checkResponse() error {
	err := h.checkScript()
	if err != nil {
		return err
	}

	err = h.checkEmails()
	if err != nil {
		return err
	}

	err = h.checkSlackChannels()
	if err != nil {
		return err
	}

	return nil
}

func (h *helper) allAttributesAreEmpty() bool {
	return len(h.applyOpts.Attributes.AlertTypes) == 0 &&
		len(h.applyOpts.Attributes.Severities) == 0 &&
		len(h.applyOpts.Attributes.Categories) == 0 &&
		len(h.applyOpts.Attributes.EventIds) == 0
}

func (h *helper) hasInvalidEventIds() error {
	for _, eventId := range h.applyOpts.Attributes.EventIds {
		if !slices.Contains(h.materials.Attribute.EventIds, eventId) {
			return fmt.Errorf(
				"invalid event id %s for trigger %s",
				eventId,
				h.applyOpts.Name,
			)
		}
	}

	return nil
}

func (h *helper) hasInvalidSeverities() error {
	for _, severity := range h.applyOpts.Attributes.Severities {
		if !slices.Contains(h.materials.Attribute.Severities, severity) {
			return fmt.Errorf(
				"invalid severity %s for trigger %s",
				severity,
				h.applyOpts.Name,
			)
		}
	}

	return nil
}

func (h *helper) hasInvaliadAlertTypes() error {
	for _, alertType := range h.applyOpts.Attributes.AlertTypes {
		if !slices.Contains(h.materials.Attribute.AlertTypes, alertType) {
			return fmt.Errorf(
				"invalid alert type %s for trigger %s",
				alertType,
				h.applyOpts.Name,
			)
		}
	}

	return nil
}

func (h *helper) hasInvalidCategories() error {
	for _, category := range h.applyOpts.Attributes.Categories {
		if !slices.Contains(h.materials.Attribute.Categories, category) {
			return fmt.Errorf(
				"invalid category %s for trigger %s",
				category,
				h.applyOpts.Name,
			)
		}
	}

	return nil
}

func (h *helper) checkScript() error {
	if h.applyOpts.ApplyResponse.Script.Content == "" {
		return nil
	}

	if h.applyOpts.ApplyResponse.Script.FilePath == "" {
		return fmt.Errorf("script file path is required for trigger %s", h.applyOpts.Name)
	}

	err := h.checkBashScript(h.applyOpts.ApplyResponse.Script.Content)
	if err != nil {
		return fmt.Errorf("failed to verify script for trigger %s: %w", h.applyOpts.Name, err)
	}

	return nil
}

func (h *helper) checkEmails() error {
	if len(h.applyOpts.ApplyResponse.Emails) == 0 {
		return nil
	}

	for _, email := range h.applyOpts.ApplyResponse.Emails {
		if !h.hasFoundEmailInMatirial(email) {
			return fmt.Errorf(
				"email %s for trigger %s is not found in settings",
				email,
				h.applyOpts.Name,
			)
		}
	}

	return nil
}

func (h *helper) checkSlackChannels() error {
	if len(h.applyOpts.ApplyResponse.Slacks) == 0 {
		return nil
	}

	for _, slack := range h.applyOpts.ApplyResponse.Slacks {
		if !h.hasFoundSlackInMaterial(slack) {
			return fmt.Errorf(
				"slack url %s for trigger %s is not found in settings",
				slack,
				h.applyOpts.Name,
			)
		}
	}

	return nil
}

func (h *helper) checkBashScript(script string) error {
	decodedBytes, err := base64.StdEncoding.DecodeString(script)
	if err != nil {
		return err
	}

	script = string(decodedBytes)
	lines := strings.Split(script, "\n")
	if len(lines) == 0 {
		return fmt.Errorf("script for trigger %s is empty", h.applyOpts.Name)
	}

	if !strings.HasPrefix(lines[0], "#!/bin/bash") && !strings.HasPrefix(lines[0], "#!/usr/bin/env bash") {
		return fmt.Errorf(
			"script for trigger %s must start with a shebang line",
			h.applyOpts.Name,
		)
	}

	return nil
}

func (h *helper) hasFoundEmailInMatirial(email string) bool {
	for _, e := range h.materials.Response.Emails {
		if e.Address == email {
			return true
		}
	}

	return false
}

func (h *helper) hasFoundSlackInMaterial(slackUrl string) bool {
	for _, s := range h.materials.Response.Slacks {
		if s.Url == slackUrl {
			return true
		}
	}

	return false
}

func (h *helper) decodeScript(content string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(content)
	if err != nil {
		return "", fmt.Errorf("failed to decode script(%v)", err)
	}

	return string(decoded), nil
}
