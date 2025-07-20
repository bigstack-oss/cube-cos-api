package nodes

import "github.com/bigstack-oss/cube-cos-api/internal/definition/v1/notifications"

type DeviceListOpts struct {
	UseCache bool
	Notify
}

type Notify struct {
	Changes bool
	Payload notifications.Notification
}
