package nodes

import (
	"fmt"
	"strings"
)

func (h *helper) validateIpmiOperation() error {
	switch strings.ToLower(h.operation) {
	case "poweron", "poweroff", "powercycle":
		return nil
	default:
		return fmt.Errorf(
			"unsupport ipmi operation(%s), should be one of [poweron, poweroff, powercycle]",
		)
	}
}
