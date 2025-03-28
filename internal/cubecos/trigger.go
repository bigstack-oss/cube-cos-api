package cubecos

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"reflect"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/trigger"
	"github.com/google/uuid"
	log "go-micro.dev/v5/logger"
	"gopkg.in/yaml.v3"
)

func SyncTriggers() {}

func GetTriggerPolicy() (*trigger.Policy, error) {
	b, err := os.ReadFile(trigger.ResponsePolicyV2)
	if err != nil {
		return nil, err
	}

	policy := &trigger.Policy{}
	err = yaml.Unmarshal(b, policy)
	if err != nil {
		return nil, err
	}

	return policy, nil
}

func GetTriggerPolicyByName(name string) (*trigger.Options, error) {
	policy, err := GetTriggerPolicy()
	if err != nil {
		return nil, err
	}

	for _, trigger := range policy.Triggers {
		if trigger.Name == name {
			return &trigger, nil
		}
	}

	return nil, fmt.Errorf("trigger %s not found", name)
}

func ApplyTriggers(triggers []trigger.Options) error {
	// M1 TODO: to remove the code below once CubeCOS side finish the hex_config refactor
	trigger.WriteFakePolicyFile(&trigger.Policy{
		Name:     "alert_resp",
		Version:  2.0,
		Enable:   true,
		Triggers: triggers,
	})
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

func ApplyTrigger(isolatedDir string) error {
	out, err := exec.Command("hex_config", "apply", isolatedDir).CombinedOutput()
	if err != nil {
		log.Errorf("failed to apply trigger value: %s", string(out))
		return err
	}

	return nil
}

func IsTriggerApplied(trigger trigger.Options) error {
	policy, err := GetTriggerPolicy()
	if err != nil {
		return err
	}

	fileTrigger := policy.GetTrigger(trigger.Name)
	if !reflect.DeepEqual(fileTrigger, trigger) {
		return fmt.Errorf("trigger %s is not applied", trigger.Name)
	}

	return nil
}

func IsTriggerExist(name string) bool {
	_, err := GetTriggerPolicyByName(name)
	return err == nil
}

func genTriggersAsYaml(triggers []trigger.Options) ([]byte, error) {
	triggerTemplate := trigger.Policy{
		Name:     "alert_resp",
		Version:  2.0,
		Enable:   true,
		Triggers: triggers,
	}

	yml, err := yaml.Marshal(&triggerTemplate)
	if err != nil {
		log.Errorf("failed to marshal batch trigger info: %s", err.Error())
		return nil, err
	}

	return yml, nil
}

func writeTriggerToFile(tmpDir string, yml []byte) error {
	fullDir := fmt.Sprintf("%s/trigger", tmpDir)
	err := os.MkdirAll(fullDir, 0755)
	if err != nil {
		log.Errorf("failed to create isolated trigger directory: %s", err.Error())
		return err
	}

	file, err := os.Create(fmt.Sprintf("%s/alert_resp2_0.yml", fullDir))
	if err != nil {
		log.Errorf("failed to create isolated trigger file: %s", err.Error())
		return err
	}

	defer file.Close()
	_, err = io.Writer.Write(file, yml)
	if err != nil {
		log.Errorf("failed to write trigger info to isolated file: %s", err.Error())
		return err
	}

	return nil
}

func genTmpTriggerDir() string {
	hash := uuid.New().String()[:8]
	return fmt.Sprintf("/tmp/trigger-%s", hash)
}
