package setting

import (
	"os"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/email"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/slack"
	log "go-micro.dev/v5/logger"
	"gopkg.in/yaml.v3"
)

func init() {
	policy := genFakePolicy()
	initFakePolicyFile(policy)
}

func genFakePolicy() *EtcPolicy {
	policy := &EtcPolicy{
		Name:        "alert_setting",
		Version:     1.0,
		Enabled:     true,
		TitlePrefix: "",
		Receiver: Receiver{
			Emails: []email.Recipient{},
			Slacks: []slack.Channel{},
		},
	}

	return policy
}

func initFakePolicyFile(policy *EtcPolicy) {
	_, err := os.Stat(PolicyV1)
	if err == nil {
		return
	}

	WriteFakePolicyFile(policy)
}

func WriteFakePolicyFile(policy *EtcPolicy) {
	os.MkdirAll("/etc/policies/alert_setting", 0755)
	policyFile, err := os.Create(PolicyV1)
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
