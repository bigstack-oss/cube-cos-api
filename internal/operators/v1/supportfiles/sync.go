package supportfiles

import (
	"fmt"

	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/support"
	log "go-micro.dev/v5/logger"
)

func (o *Operator) operateReq(file *support.File) error {
	switch file.Status.Desired {
	case status.Create:
		return o.createSupportFile(file)
	}

	return fmt.Errorf(
		"unknown desired action(%s) for support file(%s)",
		file.Status.Desired,
		file.Name,
	)
}

func (o *Operator) createSupportFile(file *support.File) error {
	err := cubecos.CreateSupportFile(*file)
	if err != nil {
		log.Errorf("supportfiles: failed to create support file(%v)", err)
		return err
	}

	file.Name, err = cubecos.GetSupportFile(*file)
	if err != nil {
		log.Errorf("supportfiles: failed to get new support file(%v)", err)
		return err
	}

	file.Url = cubecos.GetSupportFileUrl(*file)
	return nil
}
