package bodies

import (
	"encoding/json"
	"net/http"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/firmwares"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/fixpacks"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/metric"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/support"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/tunings"
	"github.com/gin-gonic/gin"
	log "go-micro.dev/v5/logger"
)

const (
	Code   = "code"
	Status = "status"
	Msg    = "msg"
	Data   = "data"
)

type Node struct {
	Code   int        `json:"code"`
	Status string     `json:"status"`
	Msg    string     `json:"msg"`
	Data   nodes.Node `json:"data"`
}

type TuningList struct {
	Code   int    `json:"code"`
	Status string `json:"status"`
	Msg    string `json:"msg"`
	Data   tuning `json:"data"`
}

type tuning struct {
	Tunings []tunings.Tuning `json:"tunings"`
}

type ComputeStatistic struct {
	Code   int            `json:"code"`
	Status string         `json:"status"`
	Msg    string         `json:"msg"`
	Data   metric.Compute `json:"data"`
}

type SpaceStatistic struct {
	Code   int          `json:"code"`
	Status string       `json:"status"`
	Msg    string       `json:"msg"`
	Data   metric.Space `json:"data"`
}

type SupportFileList struct {
	Code   int            `json:"code"`
	Status string         `json:"status"`
	Msg    string         `json:"msg"`
	Data   []support.File `json:"data"`
}

type FirmwareUpgradeProgress struct {
	Code   int               `json:"code"`
	Status string            `json:"status"`
	Msg    string            `json:"msg"`
	Data   firmwares.Upgrade `json:"data"`
}

type Fixpack struct {
	Code   int              `json:"code"`
	Status string           `json:"status"`
	Msg    string           `json:"msg"`
	Data   fixpacks.Fixpack `json:"data"`
}

const jsonContentType = "application/json; charset=utf-8"

// The fallback body is encoded here, so the failure path cannot fail again.
var encodeFailureBody = []byte(
	`{"code":500,"status":"internal server error","msg":"failed to encode the response body"}`,
)

// writeJson encodes the body before it touches the wire. c.JSON writes the
// content type first, and gin only records an encoder failure in c.Errors, so a
// payload gin cannot encode leaves the client with a 200 and an empty body.
// Peers read that empty body as a broken JSON document. Encode first, and answer
// 500 when the payload is not encodable.
func writeJson(c *gin.Context, code int, resp gin.H) {
	body, err := json.Marshal(resp)
	if err != nil {
		log.Errorf("bodies: failed to encode the %d response body(%v)", code, err)
		c.Data(http.StatusInternalServerError, jsonContentType, encodeFailureBody)
		c.Abort()
		return
	}

	c.Data(code, jsonContentType, body)
}

func SetOk(c *gin.Context, msg string, data any) {
	resp := gin.H{Code: http.StatusOK, Status: "ok", Msg: msg}
	if data != nil {
		resp[Data] = data
	}

	writeJson(
		c,
		http.StatusOK,
		resp,
	)
}

func SetCreated(c *gin.Context, msg string, data any) {
	resp := gin.H{Code: http.StatusOK, Status: "ok", Msg: msg}
	if data != nil {
		resp[Data] = data
	}

	writeJson(
		c,
		http.StatusCreated,
		resp,
	)
}

func SetAccepted(c *gin.Context, msg string) {
	writeJson(
		c,
		http.StatusAccepted,
		gin.H{
			Code:   http.StatusAccepted,
			Status: "accepted",
			Msg:    msg,
		},
	)
}

func SetRedirect(c *gin.Context, redirectUrl string) {
	c.Redirect(
		http.StatusFound,
		redirectUrl,
	)
}

func SetBadRequest(c *gin.Context, err error, data any) {
	resp := gin.H{Code: http.StatusBadRequest, Status: "bad request", Msg: err.Error()}
	if data != nil {
		resp[Data] = data
	}

	writeJson(
		c,
		http.StatusBadRequest,
		resp,
	)
}

func SetUnauthorized(c *gin.Context, err error) {
	writeJson(
		c,
		http.StatusUnauthorized,
		gin.H{
			Code:   http.StatusUnauthorized,
			Status: "unauthorized",
			Msg:    err.Error(),
		},
	)
}

func SetNotFound(c *gin.Context, err error) {
	writeJson(
		c,
		http.StatusNotFound,
		gin.H{
			Code:   http.StatusNotFound,
			Status: "not found",
			Msg:    err.Error(),
		},
	)
}

func SetConflict(c *gin.Context, err error) {
	writeJson(
		c,
		http.StatusConflict,
		gin.H{
			Code:   http.StatusConflict,
			Status: "status conflict",
			Msg:    err.Error(),
		},
	)
}

func SetTooManyRequests(c *gin.Context, err error) {
	writeJson(
		c,
		http.StatusTooManyRequests,
		gin.H{
			Code:   http.StatusTooManyRequests,
			Status: "too many requests",
			Msg:    err.Error(),
		},
	)
}

func SetInternalServerError(c *gin.Context, err error) {
	writeJson(
		c,
		http.StatusInternalServerError,
		gin.H{
			Code:   http.StatusInternalServerError,
			Status: "internal server error",
			Msg:    err.Error(),
		},
	)
}
