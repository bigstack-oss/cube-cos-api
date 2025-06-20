package nodes

import (
	"encoding/base64"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/ipmi"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	log "go-micro.dev/v5/logger"
)

func (h *helper) verifyNodeIpmi() (*nodes.ImpiValidation, error) {
	helper, err := ipmi.NewHelper(
		ipmi.Host(h.ipmi.Host.Ip),
		ipmi.Port(h.ipmi.Port),
		ipmi.Username(h.ipmi.Username),
		ipmi.Password(h.ipmi.Password),
	)
	if err != nil {
		return nil, err
	}

	validation, err := helper.GetFRU(nodes.DefaultIpmiDeviceId)
	if err != nil {
		return nil, err
	}

	manufacturer, err := base64.StdEncoding.DecodeString(string(validation.BoardInfoArea.Manufacturer))
	if err != nil {
		log.Warnf("nodes: failed to decode ipmi manufacturer(%v)", err)
	}

	productName, err := base64.StdEncoding.DecodeString(string(validation.BoardInfoArea.ProductName))
	if err != nil {
		log.Warnf("nodes: failed to decode ipmi product name(%v)", err)
	}

	ManufacturingSerial, err := base64.StdEncoding.DecodeString(string(validation.BoardInfoArea.SerialNumber))
	if err != nil {
		log.Warnf("nodes: failed to decode ipmi serial(%v)", err)
	}

	partNumber, err := base64.StdEncoding.DecodeString(string(validation.BoardInfoArea.PartNumber))
	if err != nil {
		log.Warnf("nodes: failed to decode ipmi part number(%v)", err)
	}

	productVersion, err := base64.StdEncoding.DecodeString(string(validation.ProductInfoArea.Version))
	if err != nil {
		log.Warnf("nodes: failed to decode ipmi product version(%v)", err)
	}

	productSerial, err := base64.StdEncoding.DecodeString(string(validation.ProductInfoArea.SerialNumber))
	if err != nil {
		log.Warnf("nodes: failed to decode ipmi product serial(%v)", err)
	}

	return &nodes.ImpiValidation{
		Board: nodes.Board{
			ManufacturingDate: validation.BoardInfoArea.MfgDateTime.String(),
			Manufacturer:      string(manufacturer),
			Product:           string(productName),
			Serial:            string(ManufacturingSerial),
			PartNumber:        string(partNumber),
		},
		Product: nodes.Product{
			Manufacturer: string(manufacturer),
			Name:         string(productName),
			Version:      string(productVersion),
			Serial:       string(productSerial),
		},
	}, nil
}
