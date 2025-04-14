package settings

import (
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/email"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/setting"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/slack"
)

func convertEtcPolicyToApiPolicy(etcPolicy *setting.EtcPolicy) setting.ApiPolicy {
	senders := []email.Sender{}
	if etcPolicy.Sender != nil {
		senders = append(senders, *etcPolicy.Sender)
	}

	return setting.ApiPolicy{
		TitlePrefix: setting.TitlePrefix{
			Value: etcPolicy.TitlePrefix,
		},
		Email: email.Options{
			Senders:    senders,
			Recipients: etcPolicy.Receiver.Emails,
		},
		Slack: slack.Options{
			Channels: etcPolicy.Receiver.Slacks,
		},
	}
}
