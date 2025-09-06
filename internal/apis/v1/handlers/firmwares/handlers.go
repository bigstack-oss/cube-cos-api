package firmwares

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/bigstack-oss/cube-cos-api/internal/apis"
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/bodies"
	_ "github.com/bigstack-oss/cube-cos-api/internal/operators/v1/nodes"
	"github.com/gin-gonic/gin"
	log "go-micro.dev/v5/logger"
)

var (
	Handlers = []apis.Handler{
		{
			Version: apis.V1,
			Method:  http.MethodGet,
			Path:    "/firmwares",
			Func:    listFirmwares,
		},
		{
			Version: apis.V1,
			Method:  http.MethodPost,
			Path:    "/firmwares",
			Func:    uploadFirmware,
		},
		{
			Version: apis.V1,
			Method:  http.MethodPost,
			Path:    "/firmwares/md5sum",
			Func:    uploadFirmwareMd5Sum,
		},
		{
			Version: apis.V1,
			Method:  http.MethodPatch,
			Path:    "/firmwares",
			Func:    updateFirmware,
		},
		{
			Version: apis.V1,
			Method:  http.MethodGet,
			Path:    "/firmwares/upgradeProgress",
			Func:    getFirmwareUpgradeProgress,
		},
		{
			Version: apis.V1,
			Method:  http.MethodPost,
			Path:    "/firmwares/md5sum/verify",
			Func:    verfiyFirmwareAndMd5Sum,
		},
		{
			Version: apis.V1,
			Method:  http.MethodGet,
			Path:    "/firmwares/:version/updatableNodes",
			Func:    listUpdatableNodes,
		},
		{
			Version: apis.V1,
			Method:  http.MethodPost,
			Path:    "/firmwares/continueAnyway",
			Func:    continueInterruptedFirmwareUpdate,
		},
		{
			Version: apis.V1,
			Method:  http.MethodDelete,
			Path:    "/firmwares/:version",
			Func:    deleteFirmware,
		},
	}
)

func listFirmwares(c *gin.Context) {
	h, err := initHelper(c, "listFirmwares")
	if err != nil {
		log.Errorf("firmwares(%s): failed to init helper(%v)", h.reqId, err)
		bodies.SetBadRequest(c, err, nil)
		return
	}

	firmwares, err := h.listFirmwares()
	if err != nil {
		bodies.SetInternalServerError(c, err)
		return
	}

	bodies.SetOk(
		c,
		"List of firmwares",
		firmwares,
	)
}

func uploadFirmware(c *gin.Context) {
	h, err := initHelper(c, "uploadFirmware")
	if err != nil {
		log.Errorf("firmwares(%s): failed to init helper(%v)", h.reqId, err)
		bodies.SetBadRequest(c, err, nil)
		return
	}

	err = h.resetTmpFirmwareArtifacts()
	if err != nil {
		bodies.SetInternalServerError(c, err)
		return
	}

	duplicated, err := h.checkFirmwareDuplication()
	if err != nil {
		bodies.SetInternalServerError(c, err)
		return
	}
	if duplicated {
		bodies.SetConflict(c, fmt.Errorf("file %s already exists", h.file))
		return
	}

	err = h.saveUploadFile()
	if err != nil {
		bodies.SetBadRequest(c, err, nil)
		return
	}

	err = h.syncFirmwareMd5()
	if err != nil {
		bodies.SetInternalServerError(c, err)
		return
	}

	bodies.SetOk(
		c,
		"Firmware uploaded successfully",
		nil,
	)
}

func listUpdatableNodes(c *gin.Context) {
	h, err := initHelper(c, "listUpdatableNodes")
	if err != nil {
		log.Errorf("firmwares(%s): failed to init helper(%v)", h.reqId, err)
		bodies.SetBadRequest(c, err, nil)
		return
	}

	nodes, err := h.listUpdatableNodes()
	if err != nil {
		bodies.SetInternalServerError(c, err)
		return
	}

	bodies.SetOk(
		c,
		"List of updatable nodes",
		nodes,
	)
}

func uploadFirmwareMd5Sum(c *gin.Context) {
	h, err := initHelper(c, "uploadFirmwareMd5Sum")
	if err != nil {
		log.Errorf("firmwares(%s): failed to init helper(%v)", h.reqId, err)
		bodies.SetBadRequest(c, err, nil)
		return
	}

	err = h.resetTmpFirmwareMd5()
	if err != nil {
		bodies.SetInternalServerError(c, err)
		return
	}

	err = h.saveUploadFile()
	if err != nil {
		bodies.SetBadRequest(c, err, nil)
		return
	}

	bodies.SetOk(
		c,
		"Firmware MD5 sum uploaded successfully",
		nil,
	)
}

func verfiyFirmwareAndMd5Sum(c *gin.Context) {
	h, err := initHelper(c, "verfiyFirmwareAndMd5Sum")
	if err != nil {
		log.Errorf("firmwares(%s): failed to init helper(%v)", h.reqId, err)
		bodies.SetBadRequest(c, err, nil)
		return
	}

	result, err := h.verifyFirmwareAndMd5()
	if err != nil {
		bodies.SetBadRequest(c, err, result)
		return
	}

	err = h.setValidFirmware()
	if err != nil {
		bodies.SetInternalServerError(c, err)
		return
	}

	bodies.SetOk(
		c,
		"Firmware and MD5 sum verified successfully",
		result,
	)
}

func continueInterruptedFirmwareUpdate(c *gin.Context) {
	h, err := initHelper(c, "continueInterruptedFirmwareUpdate")
	if err != nil {
		log.Errorf("firmwares(%s): failed to init helper(%v)", h.reqId, err)
		bodies.SetBadRequest(c, err, nil)
		return
	}

	err = h.checkConditionForContinue()
	if err != nil {
		bodies.SetBadRequest(c, err, nil)
		return
	}

	err = h.continueInterruptedFirmwareUpdate()
	if err != nil {
		bodies.SetInternalServerError(c, err)
		return
	}

	bodies.SetOk(
		c,
		"Firmware update continued successfully",
		nil,
	)
}

func updateFirmware(c *gin.Context) {
	h, err := initHelper(c, "updateFirmware")
	if err != nil {
		log.Errorf("firmwares(%s): failed to init helper(%v)", h.reqId, err)
		bodies.SetBadRequest(c, err, nil)
		return
	}

	err = h.updateFirmware()
	if err != nil {
		bodies.SetInternalServerError(c, err)
		return
	}

	bodies.SetOk(
		c,
		"Firmware updated successfully",
		nil,
	)
}

func getFirmwareUpgradeProgress(c *gin.Context) {
	h, err := initHelper(c, "getFirmwareUpgradeProgress")
	if err != nil {
		log.Errorf("firmwares(%s): failed to init helper(%v)", h.reqId, err)
		bodies.SetBadRequest(c, err, nil)
		return
	}

	progresses, err := h.getFirmwareUpgradeProgress()
	if err != nil {
		bodies.SetInternalServerError(c, err)
		return
	}

	bodies.SetOk(
		c,
		"List of firmware upgrade progress",
		progresses,
	)
}

func deleteFirmware(c *gin.Context) {
	h, err := initHelper(c, "deleteFirmware")
	if err != nil {
		log.Errorf("firmwares(%s): failed to init helper(%v)", h.reqId, err)
		bodies.SetBadRequest(c, err, nil)
		return
	}

	if !h.isFirmwareExists() {
		bodies.SetNotFound(c, errors.New("firmware not found"))
		return
	}

	if h.isFirmwareInstalled() {
		bodies.SetConflict(c, errors.New("cannot delete an installed firmware"))
		return
	}

	err = h.deleteFirmware()
	if err != nil {
		bodies.SetInternalServerError(c, err)
		return
	}

	bodies.SetOk(
		c,
		"Firmware deleted successfully",
		nil,
	)
}
