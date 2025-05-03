package cubecos

import (
	"os/exec"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/trigger"
	log "go-micro.dev/v5/logger"
	"gopkg.in/yaml.v3"
)

func IsTriggerExist(name string) bool {
	triggers, err := GetTriggers()
	if err != nil {
		log.Errorf("triggers: failed to get trigger value: %s", err.Error())
		return false
	}

	for _, t := range triggers {
		if t.Name == name {
			return true
		}
	}

	return false
}

func GetTriggers() ([]trigger.CosOptions, error) {
	out, err := exec.Command("hex_config", "alert_get_trigger").CombinedOutput()
	if err != nil {
		log.Errorf("triggers: failed to get trigger value: %s", string(out))
		return nil, err
	}

	triggers := []trigger.CosOptions{}
	err = yaml.Unmarshal(out, &triggers)
	if err != nil {
		log.Errorf("triggers: failed to unmarshal trigger value: %s", err.Error())
		return nil, err
	}

	return triggers, nil
}

func ApplyTrigger(trigger trigger.CosOptions) error {
	b, err := trigger.Bytes()
	if err != nil {
		log.Errorf("triggers: failed to get trigger bytes: %s", err.Error())
		return err
	}

	out, err := exec.Command("hex_config", "alert_put_trigger", string(b)).CombinedOutput()
	if err != nil {
		log.Errorf("triggers: failed to apply trigger value: %v(%s)", err, string(out))
		return err
	}

	return nil
}
