package api

import (
	"net/http"

	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/gin-gonic/gin"
)

type NodeData struct {
	Code   int     `json:"code"`
	Status string  `json:"status"`
	Msg    string  `json:"msg"`
	Data   v1.Node `json:"data"`
}

type TuningListData struct {
	Code   int         `json:"code"`
	Status string      `json:"status"`
	Msg    string      `json:"msg"`
	Data   []v1.Tuning `json:"data"`
}

type ComputeStatisticData struct {
	Code   int                 `json:"code"`
	Status string              `json:"status"`
	Msg    string              `json:"msg"`
	Data   v1.ComputeStatistic `json:"data"`
}

type SupportFileListData struct {
	Code   int              `json:"code"`
	Status string           `json:"status"`
	Msg    string           `json:"msg"`
	Data   []v1.SupportFile `json:"data"`
}

func SetStatusOk(c *gin.Context, msg string, data interface{}) {
	resp := gin.H{Code: http.StatusOK, Status: "ok", Msg: msg}
	if data != nil {
		resp[Data] = data
	}

	c.JSON(
		http.StatusOK,
		resp,
	)
}

func SetStatusCreated(c *gin.Context, msg string, data interface{}) {
	resp := gin.H{Code: http.StatusOK, Status: "ok", Msg: msg}
	if data != nil {
		resp[Data] = data
	}

	c.JSON(
		http.StatusCreated,
		resp,
	)
}

func SetStatusAccepted(c *gin.Context, msg string) {
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

func SetStatusNotFound(c *gin.Context, err error) {
	c.JSON(
		http.StatusNotFound,
		gin.H{
			Code:   http.StatusNotFound,
			Status: "not found",
			Msg:    err.Error(),
		},
	)
}

func SetStatusConflict(c *gin.Context, err error) {
	c.JSON(
		http.StatusConflict,
		gin.H{
			Code:   http.StatusConflict,
			Status: "status conflict",
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

func SetErrVariantAlsoNegotiates(c *gin.Context, err error) {
	c.JSON(
		http.StatusVariantAlsoNegotiates,
		gin.H{
			Code:   http.StatusVariantAlsoNegotiates,
			Status: "variant also negotiates",
			Msg:    err.Error(),
		},
	)
}
