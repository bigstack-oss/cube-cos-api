package api

import (
	"net/http"

	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	ginFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

const (
	apiDoc = v1.ApiDoc
)

var (
	handlers = []Handler{
		{
			Version: V1,
			Method:  http.MethodGet,
			Path:    "/apidocs/*any",
			Func:    ginSwagger.WrapHandler(ginFiles.Handler),
		},
	}
)

func init() {
	RegisterHandlersToRoles(
		apiDoc,
		handlers,
		v1.RoleControl,
		v1.RoleCompute,
	)
}
