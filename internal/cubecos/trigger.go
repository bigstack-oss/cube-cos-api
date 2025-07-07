package cubecos

import (
	"fmt"
	"os/exec"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/events"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/triggers"
	log "go-micro.dev/v5/logger"
	"gopkg.in/yaml.v3"
)

func IsTriggerExist(name string) bool {
	list, err := GetTriggers()
	if err != nil {
		log.Errorf("triggers: failed to get trigger value(%v)", err)
		return false
	}

	cosName := triggers.GetBuiltInNameMap()
	for _, trigger := range list {
		if trigger.Name == cosName[name] {
			return true
		}
	}

	return false
}

func GetTriggers() ([]triggers.CosSchema, error) {
	out, err := exec.Command("hex_sdk", "alert_get_trigger").CombinedOutput()
	if err != nil {
		log.Errorf("triggers: failed to get trigger value: %s", string(out))
		return nil, err
	}

	triggers := []triggers.CosSchema{}
	err = yaml.Unmarshal(out, &triggers)
	if err != nil {
		log.Errorf("triggers: failed to unmarshal trigger value(%v)", err)
		return nil, err
	}

	if !IsHexSdkSuccess(err) {
		err := fmt.Errorf("triggers: failed to apply trigger: %s", string(out))
		return nil, err
	}

	return triggers, nil
}

func ApplyTrigger(trigger triggers.CosSchema) error {
	b, err := trigger.Bytes()
	if err != nil {
		log.Errorf("triggers: failed to get trigger bytes(%v)", err)
		return err
	}

	out, err := exec.Command("hex_sdk", "alert_put_trigger", string(b)).CombinedOutput()
	if err != nil {
		log.Errorf("triggers: failed to apply trigger value: %v(%s)", err, string(out))
		return err
	}

	if !IsHexSdkSuccess(err) {
		err := fmt.Errorf("triggers: failed to apply trigger: %s", string(out))
		return err
	}

	return nil
}

func GetPredefinedEvents() ([]events.Event, error) {
	return []events.Event{}, nil
}
