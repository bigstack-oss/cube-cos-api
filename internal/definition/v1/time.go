package v1

import "time"

func TimeNowRFC3339() string {
	return time.Now().UTC().Format(time.RFC3339)
}
