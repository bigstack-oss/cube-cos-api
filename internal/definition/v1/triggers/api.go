package triggers

const (
	Triggers         = "triggers"
	DB               = "triggers"
	Collection       = "triggers"
	ReqCollection    = "requests"
	ResponsePolicyV2 = "/etc/policies/alert_resp/alert_resp2_0.yml"
	ISO8601Z         = "2006-01-02T15:04:05+00:00"
)

var (
	builtInNameMap = map[string]string{
		"Administrative Level Notification": "admin-notify",
		"Instance Level Notification":       "instance-notify",
	}
)

func Get(name string) (*Trigger, bool) {
	for _, trigger := range List() {
		if trigger.Name == name {
			return &trigger, true
		}
	}

	return nil, false
}

func GetBuiltInNameMap() map[string]string {
	return builtInNameMap
}
