package supportfiles

import (
	"fmt"

	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/bigstack-oss/cube-cos-api/internal/status"
	log "go-micro.dev/v5/logger"
)

func (o *Operator) operateReq(supportFile v1.SupportFile) error {
	switch supportFile.Status.Desired {
	case status.Create:
		return o.createSupportFile(&supportFile)
	}

	return fmt.Errorf(
		"unknown desired action(%s) for support file(%s)",
		supportFile.Status.Desired,
		supportFile.Name,
	)
}

func (o *Operator) createSupportFile(supportFile *v1.SupportFile) error {
	err := cubecos.CreateSupportFile()
	if err != nil {
		log.Errorf("supportfiles: failed to create support file: %s", err.Error())
		return err
	}

	supportFile.Name, err = cubecos.GetNewSupportFile()
	if err != nil {
		log.Errorf("supportfiles: failed to get new support file: %s", err.Error())
		return err
	}

	return nil
}
