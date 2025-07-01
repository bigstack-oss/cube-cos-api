package nodes

import (
	"fmt"
	"time"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/ipmi"
	bstime "github.com/bigstack-oss/cube-cos-api/internal/definition/v1/time"
)

func (h *helper) verifyNodeIpmi() (*ipmi.FRU, error) {
	helper, err := ipmi.NewHelper(
		ipmi.Host(h.ipmi.Ip),
		ipmi.Port(h.ipmi.Port),
		ipmi.Username(h.ipmi.Username),
		ipmi.Password(h.ipmi.Password),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to IPMI %s(%v)", h.ipmi.Ip, err)
	}

	fru, err := helper.GetFRU()
	if err != nil {
		return nil, fmt.Errorf("unable to connect with IPMI, please check your IPMI settings")
	}

	dateTime, err := time.Parse(bstime.FormatBmc, fru.ManufacturingDate)
	if err == nil {
		fru.ManufacturingDate = bstime.ISO8601Z(dateTime)
	}

	return fru, nil
}
