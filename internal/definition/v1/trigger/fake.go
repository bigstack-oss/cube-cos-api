package trigger

import (
	"os"

	log "go-micro.dev/v5/logger"
	"gopkg.in/yaml.v3"
)

func init() {
	policy := genFakePolicy()
	WriteFakePolicyFile(policy)
}

func WriteFakePolicyFile(policy *Policy) {
	_, err := os.Stat(ResponsePolicyV2)
	if err == nil {
		return
	}

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

func genFakePolicy() *Policy {
	policy := &Policy{
		Name:     "alert_resp",
		Version:  2.0,
		Enable:   true,
		Triggers: DefaultOptions,
	}

	for i, trigger := range policy.Triggers {
		policy.Triggers[i].Match = trigger.GenMatchRule()
	}

	return policy
}
