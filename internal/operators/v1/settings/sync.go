package settings

import (
	"fmt"

	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/setting"
	"github.com/bigstack-oss/cube-cos-api/internal/status"
	log "go-micro.dev/v5/logger"
)

func (o *Operator) operateReq(setting setting.Options) error {
	switch setting.Status.Desired {
	case status.Updated:
		return o.updateSetting(setting)
	case status.Deleted:
		return o.deleteSetting(setting)
	}

	return fmt.Errorf(
		"unknown desired action(%s) for setting(%s)",
		setting.Status.Desired,
		setting.Type,
	)
}

func (o *Operator) updateSetting(setting setting.Options) error {
	policy, err := cubecos.GetAlertSetting()
	if err != nil {
		log.Infof("settings: %v", err)
		return err
	}

	policy.UpdateOrAppendSetting(setting)
	err = cubecos.ApplySettings(policy)
	if err != nil {
		return err
	}

	if !cubecos.IsSettingApplied(setting) {
		return fmt.Errorf("settings: %s(%s) is not applied", setting.Type, setting.GetKey())
	}

	return nil
}

func (o *Operator) deleteSetting(setting setting.Options) error {
	policy, err := cubecos.GetAlertSetting()
	if err != nil {
		log.Infof("settings: %v", err)
		return err
	}

	policy.DeleteSetting(setting)
	err = cubecos.ApplySettings(policy)
	if err != nil {
		return err
	}

	if cubecos.IsSettingDeleted(setting) {
		return fmt.Errorf("settings: %s(%s) is not deleted", setting.Type, setting.GetKey())
	}

	return nil
}
