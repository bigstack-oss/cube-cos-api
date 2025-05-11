package cubecos

import (
	"os"
	"os/exec"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/email"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/errors"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/settings"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/slack"
	json "github.com/json-iterator/go"
	log "go-micro.dev/v5/logger"
	"gopkg.in/yaml.v3"
)

func IsAlertSetting(file string) bool {
	return file == settings.PolicyV1
}

func SyncAlertSettings() {
	srcSetting, err := GetSourceAlertSetting()
	if err != nil {
		return
	}

	settings.SetCosSchema(srcSetting)
}

func GetSourceAlertSetting() (*settings.Cos, error) {
	out, err := exec.Command("hex_sdk", "alert_get_setting").CombinedOutput()
	if err != nil {
		log.Errorf("settings: failed to get cos alert settings: %s(%s)", string(out), err.Error())
		return nil, errors.ErrSdkExecutionFailure
	}

	settings := &settings.Cos{}
	err = json.Unmarshal(out, settings)
	if err != nil {
		log.Errorf("settings: failed to unmarshal cos alert settings: %s(%s)", string(out), err.Error())
		return nil, err
	}

	return settings, nil
}

func GetAlertSetting() (*settings.Cos, error) {
	maxTries := 10
	for range maxTries {
		alert := settings.GetCosSchema()
		if alert != nil {
			return alert, nil
		}

		wait.Seconds(2)
		SyncAlertSettings()
	}

	return nil, errors.ErrAlertSettingNotInited
}

func GetEmailSenders() ([]email.Sender, error) {
	setting, err := GetAlertSetting()
	if err != nil {
		return nil, err
	}

	sender := setting.Sender.Email.ToApiSchema()
	return []email.Sender{sender}, nil
}

func GetEmailRecipients() ([]email.Recipient, error) {
	setting, err := GetAlertSetting()
	if err != nil {
		return nil, err
	}

	return setting.Emails, nil
}

func GetSlackChannel(channel string) (slack.CosChannel, error) {
	setting, err := GetAlertSetting()
	if err != nil {
		log.Errorf("settings: failed to get slack channel (%s)", err.Error())
		return slack.CosChannel{}, err
	}

	for _, slack := range setting.Slacks {
		if slack.Channel == channel {
			return slack, nil
		}
	}

	return slack.CosChannel{}, nil
}

func GetSlackChannels() ([]slack.CosChannel, error) {
	setting, err := GetAlertSetting()
	if err != nil {
		log.Errorf("settings: failed to get slack channels (%s)", err.Error())
		return nil, err
	}

	return setting.Slacks, nil
}

func ApplyTitlePrefix(titlePrefix string) error {
	payload := map[string]string{"titlePrefix": titlePrefix}
	bytes, err := json.Marshal(payload)
	if err != nil {
		log.Errorf("settings: failed to marshal title prefix (%s)", err.Error())
		return err
	}

	out, err := exec.Command("hex_sdk", "alert_set_setting_title_prefix", string(bytes)).CombinedOutput()
	if err != nil {
		log.Errorf("settings: failed to set title prefix (%s)", err.Error())
		return errors.ErrSdkExecutionFailure
	}

	err = checkSettingReturnError(err)
	if err != nil {
		log.Errorf("settings: failed to set title prefix: %s (%s)", string(out), err.Error())
		return errors.ErrSdkExecutionFailure
	}

	return nil
}

func DeleteEmailSender() error {
	out, err := exec.Command("hex_sdk", "alert_delete_setting_sender_email").CombinedOutput()
	if err != nil {
		log.Errorf("settings: failed to delete email sender (%s)", err.Error())
		return errors.ErrSdkExecutionFailure
	}

	err = checkSettingReturnError(err)
	if err != nil {
		log.Errorf("settings: failed to delete email sender: %s (%s)", string(out), err.Error())
		return errors.ErrSdkExecutionFailure
	}

	return nil
}

func ApplyEmailSender(sender email.Sender) error {
	bytes, err := json.Marshal(sender)
	if err != nil {
		log.Errorf("settings: failed to marshal email sender (%s)", err.Error())
		return err
	}

	out, err := exec.Command("hex_sdk", "alert_set_setting_sender_email", string(bytes)).CombinedOutput()
	if err != nil {
		log.Errorf("settings: failed to set email sender (%s)", err.Error())
		return errors.ErrSdkExecutionFailure
	}

	err = checkSettingReturnError(err)
	if err != nil {
		log.Errorf("settings: failed to set email sender: %s (%s)", string(out), err.Error())
		return errors.ErrSdkExecutionFailure
	}

	return nil
}

func DeleteAndCreateEmailRecipient(setting settings.Setting) error {
	err := DeleteEmailRecipient(setting.Key)
	if err != nil {
		return err
	}

	return ApplyEmailRecipient(*setting.Recipient)
}

func ApplyEmailRecipient(recipient email.Recipient) error {
	bytes, err := json.Marshal(recipient)
	if err != nil {
		log.Errorf("settings: failed to marshal email recipient (%s)", err.Error())
		return err
	}

	out, err := exec.Command("hex_sdk", "alert_put_setting_receiver_email", string(bytes)).CombinedOutput()
	if err != nil {
		log.Errorf("settings: failed to set email recipient (%s)", err.Error())
		return errors.ErrSdkExecutionFailure
	}

	err = checkSettingReturnError(err)
	if err != nil {
		log.Errorf("settings: failed to set email recipient: %s (%s)", string(out), err.Error())
		return errors.ErrSdkExecutionFailure
	}

	return nil
}

func DeleteEmailRecipient(address string) error {
	payload := map[string]string{"address": address}
	bytes, err := json.Marshal(payload)
	if err != nil {
		log.Errorf("settings: failed to marshal email recipient (%s)", err.Error())
		return err
	}

	out, err := exec.Command("hex_sdk", "alert_delete_setting_receiver_email", string(bytes)).CombinedOutput()
	if err != nil {
		log.Errorf("settings: failed to delete email recipient (%s)", err.Error())
		return errors.ErrSdkExecutionFailure
	}

	err = checkSettingReturnError(err)
	if err != nil {
		log.Errorf("settings: failed to delete email recipient: %s (%s)", string(out), err.Error())
		return errors.ErrSdkExecutionFailure
	}

	return nil
}

func DeleteAndCreateSlackChannel(setting settings.Setting) error {
	err := DeleteSlackChannel(setting.Key)
	if err != nil {
		return err
	}

	return ApplySlackChannel(setting.Slack.ToCosSchema())
}

func ApplySlackChannel(channel slack.CosChannel) error {
	bytes, err := json.Marshal(channel)
	if err != nil {
		log.Errorf("settings: failed to marshal slack channel (%s)", err.Error())
		return err
	}

	out, err := exec.Command("hex_sdk", "alert_put_setting_receiver_slack", string(bytes)).CombinedOutput()
	if err != nil {
		log.Errorf("settings: failed to set slack channel (%s)", err.Error())
		return errors.ErrSdkExecutionFailure
	}

	err = checkSettingReturnError(err)
	if err != nil {
		log.Errorf("settings: failed to set slack channel: %s (%s)", string(out), err.Error())
		return errors.ErrSdkExecutionFailure
	}

	return nil
}

func DeleteSlackChannel(url string) error {
	payload := map[string]string{"url": url}
	bytes, err := json.Marshal(payload)
	if err != nil {
		log.Errorf("settings: failed to marshal slack channel (%s)", err.Error())
		return err
	}

	out, err := exec.Command("hex_sdk", "alert_delete_setting_receiver_slack", string(bytes)).CombinedOutput()
	if err != nil {
		log.Errorf("settings: failed to delete slack channel (%s)", err.Error())
		return errors.ErrSdkExecutionFailure
	}

	err = checkSettingReturnError(err)
	if err != nil {
		log.Errorf("settings: failed to delete slack channel: %s (%s)", string(out), err.Error())
		return errors.ErrSdkExecutionFailure
	}

	return nil
}

func WriteFakePolicyFile(policy *settings.Cos) {
	policyFile, err := os.Create(settings.PolicyV1)
	if err != nil {
		log.Errorf("settings: failed to create fake policy file: %s", err.Error())
		return
	}

	defer policyFile.Close()
	yamlEncoder := yaml.NewEncoder(policyFile)
	yamlEncoder.SetIndent(2)
	err = yamlEncoder.Encode(policy)
	if err != nil {
		log.Errorf("settings: failed to encode fake policy to yaml: %s", err.Error())
	}
}

func IsSettingDeleted(setting settings.Setting) bool {
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
		channel := setting.Slack.ToCosSchema()
		isDeleted = policy.HasSlack(channel)
	}

	return isDeleted
}

func checkSettingReturnError(err error) error {
	if err == nil {
		return nil
	}

	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		log.Errorf("settings: failed to get setting exit error (%s)", err.Error())
		return errors.ErrSdkExecutionFailure
	}

	if exitErr.ExitCode() != 0 {
		log.Errorf("settings: failed to get setting exit code (%d)", exitErr.ExitCode())
		return errors.ErrSdkExecutionFailure
	}

	return nil
}
