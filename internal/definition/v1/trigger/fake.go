package trigger

import (
	"os"

	log "go-micro.dev/v5/logger"
	"gopkg.in/yaml.v3"
)

func init() {
	policy := genFakePolicy()
	initFakePolicyFile(policy)
}

func initFakePolicyFile(policy *Policy) {
	_, err := os.Stat(ResponsePolicyV2)
	if err == nil {
		return
	}

	WriteFakePolicyFile(policy)
}

func genFakePolicy() *Policy {
	policy := &Policy{
		Name:     "alert_resp",
		Version:  2.0,
		Enabled:  true,
		Triggers: DefaultOptions,
	}

	for i, trigger := range policy.Triggers {
		policy.Triggers[i].Match = trigger.GenMatchRule()
	}

	return policy
}

func WriteFakePolicyFile(policy *Policy) {
	policyFile, err := os.Create(ResponsePolicyV2)
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
