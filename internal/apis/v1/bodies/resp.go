package bodies

import (
	"net/http"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/metric"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/support"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/tunings"
	"github.com/gin-gonic/gin"
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

func SetOk(c *gin.Context, msg string, data any) {
	resp := gin.H{Code: http.StatusOK, Status: "ok", Msg: msg}
	if data != nil {
		resp[Data] = data
	}

	c.JSON(
		http.StatusOK,
		resp,
	)
}

func SetCreated(c *gin.Context, msg string, data any) {
	resp := gin.H{Code: http.StatusOK, Status: "ok", Msg: msg}
	if data != nil {
		resp[Data] = data
	}

	c.JSON(
		http.StatusCreated,
		resp,
	)
}

func SetAccepted(c *gin.Context, msg string) {
	c.JSON(
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

func SetBadRequest(c *gin.Context, err error) {
	c.JSON(
		http.StatusBadRequest,
		gin.H{
			Code:   http.StatusBadRequest,
			Status: "bad request",
			Msg:    err.Error(),
		},
	)
}

func SetUnauthorized(c *gin.Context, err error) {
	c.JSON(
		http.StatusUnauthorized,
		gin.H{
			Code:   http.StatusUnauthorized,
			Status: "unauthorized",
			Msg:    err.Error(),
		},
	)
}

func SetNotFound(c *gin.Context, err error) {
	c.JSON(
		http.StatusNotFound,
		gin.H{
			Code:   http.StatusNotFound,
			Status: "not found",
			Msg:    err.Error(),
		},
	)
}

func SetConflict(c *gin.Context, err error) {
	c.JSON(
		http.StatusConflict,
		gin.H{
			Code:   http.StatusConflict,
			Status: "status conflict",
			Msg:    err.Error(),
		},
	)
}

func SetTooManyRequests(c *gin.Context, err error) {
	c.JSON(
		http.StatusTooManyRequests,
		gin.H{
			Code:   http.StatusTooManyRequests,
			Status: "too many requests",
			Msg:    err.Error(),
		},
	)
}

func SetInternalServerError(c *gin.Context, err error) {
	c.JSON(
		http.StatusInternalServerError,
		gin.H{
			Code:   http.StatusInternalServerError,
			Status: "internal server error",
			Msg:    err.Error(),
		},
	)
}
