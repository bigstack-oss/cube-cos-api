package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func SetStatusOk(c *gin.Context, msg string, data interface{}) {
	c.JSON(
		http.StatusOK,
		gin.H{
			Code:   http.StatusOK,
			Status: "ok",
			Msg:    msg,
			Data:   data,
		},
	)
}

func SetStatusCreated(c *gin.Context, msg string, data interface{}) {
	c.JSON(
		http.StatusCreated,
		gin.H{
			Code:   http.StatusCreated,
			Status: "created",
			Msg:    msg,
			Data:   data,
		},
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
