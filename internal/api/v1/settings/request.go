package settings

import "github.com/bigstack-oss/cube-cos-api/internal/definition/v1/setting"

func (h *helper) genUpdateReq(settingType string) setting.Options {
	req := setting.Options{}

	switch settingType {
	case "titlePrefix":
		req.Type = settingType
		req.TitlePrefix = &h.titlePrefix
		req.InitUpdateStatus()
	}

	return req
}
