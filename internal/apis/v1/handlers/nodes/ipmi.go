package nodes

import (
	"fmt"
	"strings"
	"time"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/ipmi"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
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

	manufacturingDate, err := convertBmcTime(fru.ManufacturingDate)
	if err == nil {
		fru.ManufacturingDate = manufacturingDate
	}

	return fru, nil
}

// FormatBmc carries no zone token, so the BMC wall clock must be parsed as the
// node's local time — otherwise it is read as UTC and mislabeled.
func convertBmcTime(rawTime string) (string, error) {
	t, err := time.ParseInLocation(bstime.FormatBmc, rawTime, bstime.LocalFixedZone)
	if err != nil {
		return "", err
	}

	return bstime.RFC3339Z(t), nil
}

func (h *helper) checkBoardSerialConsistency(fru *ipmi.FRU) error {
	node, err := nodes.Get(h.node)
	if err != nil {
		return err
	}

	return matchBoardSerial(h.node, node.BoardSerial, h.ipmi.Ip, fru)
}

// matchBoardSerial accepts the BMC only when its FRU board serial is the
// node's board serial: either the same plain value, or the dotted
// ".<product serial>.<board serial>." layout some vendors put in DMI.
// Anything else, including an empty serial on either side, is refused.
func matchBoardSerial(host, nodeSerial, bmcIp string, fru *ipmi.FRU) error {
	nodeSerial = strings.TrimSpace(nodeSerial)
	boardSerial := strings.TrimSpace(fru.Board.Serial)
	productSerial := strings.TrimSpace(fru.Product.Serial)

	if boardSerial == "" {
		return fmt.Errorf("board serial is mismatched: BMC %s reports no board serial", bmcIp)
	}

	if nodeSerial == "" {
		return fmt.Errorf("board serial is mismatched: host %s has no board serial on record", host)
	}

	if nodeSerial == boardSerial {
		return nil
	}

	if productSerial != "" && nodeSerial == fmt.Sprintf(".%s.%s.", productSerial, boardSerial) {
		return nil
	}

	return fmt.Errorf(
		"board serial is mismatched: %s serial is '%s', but host %s board serial is '%s'",
		bmcIp,
		boardSerial,
		host,
		nodeSerial,
	)
}
