package runtime

import "fmt"

func GetAdvertiseAddress() string {
	return fmt.Sprintf(
		"%s:%d",
		conf.Spec.Listen.Advertise,
		conf.Spec.Listen.Port,
	)
}
