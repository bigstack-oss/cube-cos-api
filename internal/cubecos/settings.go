package cubecos

import (
	"os"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/setting"
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
