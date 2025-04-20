package cubecos

import (
	"os"
	"os/exec"

	json "github.com/json-iterator/go"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/email"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/setting"
	cuberr "github.com/bigstack-oss/cube-cos-api/internal/errors"
	log "go-micro.dev/v5/logger"
	"gopkg.in/yaml.v3"
)

func GetAlertSetting() (*setting.CosAlert, error) {
	out, err := exec.Command("hex_sdk", "alert_get_setting").CombinedOutput()
	if err != nil {
		return nil, cuberr.SdkExecutionError
	}

	settings := &setting.CosAlert{}
	err = json.Unmarshal(out, settings)
	if err != nil {
		log.Errorf("settings: failed to unmarshal cos alert settings (%s)", err.Error())
		return nil, err
	}

	return settings, nil
}

func GetEmailSenders() ([]email.Sender, error) {
	policy, err := GetAlertSetting()
	if err != nil {
		log.Errorf("settings: failed to get email senders (%s)", err.Error())
		return nil, err
	}

	return []email.Sender{*policy.Sender}, nil
}

func ApplySettings(policy *setting.CosAlert) error {

	return nil
}

func WriteFakePolicyFile(policy *setting.CosAlert) {
	policyFile, err := os.Create(setting.PolicyV1)
	if err != nil {
		log.Errorf("failed to create fake policy file: %s", err.Error())
		return
	}

	defer policyFile.Close()
	yamlEncoder := yaml.NewEncoder(policyFile)
	yamlEncoder.SetIndent(2)
	err = yamlEncoder.Encode(policy)
	if err != nil {
		log.Errorf("failed to encode fake policy to yaml: %s", err.Error())
	}
}

func IsSettingApplied(setting setting.Options) bool {
	policy, err := GetAlertSetting()
	if err != nil {
		return false
	}

	isApplied := false
	switch setting.Type {
	case "titlePrefix":
		isApplied = policy.IsTitlePrefixEqual(setting.TitlePrefix.Value)
	case "emailSender":
		isApplied = policy.IsSenderEqual(*setting.Sender)
	case "emailRecipient":
		isApplied = policy.IsRecipientEqual(*setting.Recipient)
	case "slackChannel":
		isApplied = policy.IsSlackChannelEqual(*setting.Slack)
	}

	return isApplied
}

func IsSettingDeleted(setting setting.Options) bool {
	policy, err := GetAlertSetting()
	if err != nil {
		return false
	}

	isDeleted := false
	switch setting.Type {
	case "emailSender":
		isDeleted = policy.HasSender(setting.Sender.Host)
	case "emailRecipient":
		isDeleted = policy.HasRecipient(setting.Recipient.Address)
	case "slackChannel":
		isDeleted = policy.HasSlackChannel(setting.Slack.Name)
	}

	return isDeleted
}
