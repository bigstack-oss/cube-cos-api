package apis

import (
	"net/http"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	ginFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

const (
	apiDoc = "api-doc"
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
		nodes.RoleControl,
		nodes.RoleModerator,
	)
}
