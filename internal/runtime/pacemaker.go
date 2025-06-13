package runtime

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/pacemaker"
)

const (
	notifierCmds = `
		#!/bin/bash
		echo "Promoted host changed on $(date)" > /var/log/pacemaker/alerts/changed.txt
	`
)

func syncPacemakerAlertPreflight() error {
	err := syncPacemakerAlertDir()
	if err != nil {
		return err
	}

	err = syncPacemakerAlertNotifier()
	if err != nil {
		return err
	}

	return err
}

func syncPacemakerAlertOperation() error {
	err := syncPacemakerAlertRule()
	if err != nil {
		return err
	}

	err = syncPacemakerAlertReceiver()
	if err != nil {
		return err
	}

	return nil
}

func syncPacemakerAlertDir() error {
	for _, dir := range pacemaker.AlertDirs {
		err := os.MkdirAll(dir, 0777)
		if err == nil {
			continue
		}

		if !os.IsExist(err) {
			return err
		}
	}

	return nil
}

func syncPacemakerAlertNotifier() error {
	err := os.WriteFile(pacemaker.AlertNotifier, []byte(notifierCmds), 0777)
	if err == nil {
		return nil
	}

	if os.IsExist(err) {
		return nil
	}

	return fmt.Errorf(
		"runtime: failed to write pacemaker notifier script: %v",
		err,
	)
}

func syncPacemakerAlertRule() error {
	out, err := exec.Command("pcs", "alert", "create", fmt.Sprintf("id=%s", pacemaker.AlertId), fmt.Sprintf("path=%s", pacemaker.AlertScriptDir), "meta", "timeout=50s", `timestamp-format="%H%B%S"`).CombinedOutput()
	if err == nil {
		return nil
	}

	if isRuleExists(err) {
		return nil
	}

	return fmt.Errorf(
		"runtime: failed to create pacemaker alert rule: %s: %v",
		string(out),
		err,
	)
}

func syncPacemakerAlertReceiver() error {
	out, err := exec.Command("pcs", "alert", "recipient", "add", pacemaker.AlertId, "value=pcs-alert-receipient", "--force").CombinedOutput()
	if err == nil {
		return nil
	}

	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		return fmt.Errorf(
			"runtime: failed to get error code from adding pcs alert recipient: %s: %v",
			string(out),
			err,
		)
	}

	if exitErr.ExitCode() != 0 {
		return fmt.Errorf(
			"runtime: failed to add pacemaker alert recipient: %s: %v",
			string(out),
			err,
		)
	}

	return nil
}

func isRuleExists(err error) bool {
	return strings.Contains(err.Error(), "already exists")
}
