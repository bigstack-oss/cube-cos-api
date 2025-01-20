package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func SetStatusOkResp(c *gin.Context, msg string, data interface{}) {
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

func SetRedirectResp(c *gin.Context, redirectUrl string) {
	c.Redirect(
		http.StatusFound,
		redirectUrl,
	)
}

func SetErrBadRequestResp(c *gin.Context, err error) {
	c.JSON(
		http.StatusBadRequest,
		gin.H{
			Code:   http.StatusBadRequest,
			Status: "bad request",
			Msg:    err.Error(),
		},
	)
}

func SetErrInternalServerErrorResp(c *gin.Context, err error) {
	c.JSON(
		http.StatusInternalServerError,
		gin.H{
			Code:   http.StatusInternalServerError,
			Status: "internal server error",
			Msg:    err.Error(),
		},
	)
}
