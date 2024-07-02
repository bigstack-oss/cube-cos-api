package api

import (
	"net/http"

	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	ginFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

const (
	apiDoc = definition.ApiDoc
)

var (
	handlers = []Handler{
		{
			Version: V1,
			Method:  http.MethodGet,
			Path:    "/api-doc/*any",
			Func:    ginSwagger.WrapHandler(ginFiles.Handler),
		},
	}
)

func init() {
	RegisterHandlersToRoles(
		apiDoc,
		handlers,
		definition.RoleControl,
		definition.RoleCompute,
	)
}
