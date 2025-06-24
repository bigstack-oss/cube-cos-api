package nodes

import (
	"fmt"
	"strings"
	"time"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/ipmi"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	bstime "github.com/bigstack-oss/cube-cos-api/internal/definition/v1/time"
	log "go-micro.dev/v5/logger"
)

func (h *helper) verifyNodeIpmi() (*nodes.ImpiValidation, error) {
	helper, err := ipmi.NewHelper(
		ipmi.Host(h.ipmi.Ip),
		ipmi.Port(h.ipmi.Port),
		ipmi.Username(h.ipmi.Username),
		ipmi.Password(h.ipmi.Password),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to IPMI %s(%v)", h.ipmi.Ip, err)
	}

	defer helper.Close()
	validation, err := helper.GetFRU(nodes.DefaultIpmiDeviceId)
	if err != nil {
		return nil, fmt.Errorf("unable to get IPMI board info %s(%v)", h.ipmi.Ip, err)
	}

	manufacturerDate, err := time.Parse(bstime.FormatBmc, validation.BoardInfoArea.MfgDateTime.String())
	if err != nil {
		log.Warnf("nodes(%s): failed to parse IPMI manufacturer date(%v)", h.reqId, err)
	}

	return &nodes.ImpiValidation{
		Board: nodes.Board{
			ManufacturingDate: bstime.RFC3339Z(manufacturerDate),
			Manufacturer:      string(validation.BoardInfoArea.Manufacturer),
			Product:           strings.TrimRight(string(validation.BoardInfoArea.ProductName), " "),
			Serial:            string(validation.BoardInfoArea.SerialNumber),
			PartNumber:        string(validation.BoardInfoArea.PartNumber),
		},
		Product: nodes.Product{
			Manufacturer: string(validation.BoardInfoArea.Manufacturer),
			Name:         strings.TrimRight(string(validation.BoardInfoArea.ProductName), " "),
			Version:      string(validation.ProductInfoArea.Version),
			Serial:       strings.TrimRight(string(validation.ProductInfoArea.SerialNumber), "\x00"),
		},
	}, nil
}
