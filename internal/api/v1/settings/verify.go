package settings

import (
	"errors"

	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/email"
)

func (h *helper) checkRecipientUpdate() error {
	if !h.isRecipientExist() {
		return errors.New("recipient not found")
	}

	err := email.CheckFormat(h.c.Param("recipientEmail"))
	if err != nil {
		return errors.New("recipient email format is invalid")
	}

	return nil
}

func (h *helper) isRecipientExist() bool {
	policy, err := cubecos.GetEtcSettingPolicy()
	if err != nil {
		return false
	}

	return policy.HasRecipient(h.c.Param("recipientEmail"))
}

func (h *helper) isSlackChannlExist() bool {
	policy, err := cubecos.GetEtcSettingPolicy()
	if err != nil {
		return false
	}

	return policy.HasSlackChannel(h.c.Param("channelName"))
}
