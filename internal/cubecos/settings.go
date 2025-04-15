package cubecos

import (
	"os"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/email"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/setting"
	log "go-micro.dev/v5/logger"
	"gopkg.in/yaml.v3"
)

func GetEtcSettingPolicy() (*setting.EtcPolicy, error) {
	b, err := os.ReadFile(setting.PolicyV1)
	if err != nil {
		return nil, err
	}

	settings := &setting.EtcPolicy{}
	err = yaml.Unmarshal(b, settings)
	if err != nil {
		return nil, err
	}

	return settings, nil
}

func GetEmailSenders() ([]email.Sender, error) {
	policy, err := GetEtcSettingPolicy()
	if err != nil {
		log.Errorf("settings: failed to get email senders (%s)", err.Error())
		return nil, err
	}

	return []email.Sender{*policy.Sender}, nil
}

func ApplySettings(policy *setting.EtcPolicy) error {
	// M1 TODO: to remove the code below once CubeCOS side finish the hex_config refactor
	setting.WriteFakePolicyFile(policy)
	wait.Seconds(3)
	return nil

	// M1 TODO: to recover the code below once CubeCOS side finish the hex_config refactor
	// newTriggers, err := genTriggersAsYaml(triggers)
	// if err != nil {
	// 	return err
	// }

	// tmpTriggerDir := genTmpTriggerDir()
	// err = writeTriggerToFile(tmpTriggerDir, newTriggers)
	// if err != nil {
	// 	return err
	// }

	// err = ApplyTrigger(tmpTriggerDir)
	// if err != nil {
	// 	return err
	// }

	// return nil
}

func WriteFakePolicyFile(policy *setting.EtcPolicy) {
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
	policy, err := GetEtcSettingPolicy()
	if err != nil {
		return false
	}

	isApplied := false
	switch setting.Type {
	case "titlePrefix":
		isApplied = policy.IsTitlePrefixEqual(setting.TitlePrefix.Value)
	case "emailSender":
		isApplied = policy.IsSenderEqual(*setting.Sender)
	}

	return isApplied
}

func IsSettingDeleted(setting setting.Options) bool {
	policy, err := GetEtcSettingPolicy()
	if err != nil {
		return false
	}

	isDeleted := false
	switch setting.Type {
	case "emailSender":
		isDeleted = policy.HasSender(setting.Sender.Host)
	}

	return isDeleted
}
