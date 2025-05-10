package apis

import (
	"fmt"

	"maps"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	"github.com/gin-gonic/gin"
)

const (
	API = "api"
)

var (
	Role string
	V1   = fmt.Sprintf("/%s/%s", API, "v1")

	ControlHandlers   = map[string][]Handler{}
	ComputeHandlers   = map[string][]Handler{}
	StorageHandlers   = map[string][]Handler{}
	ModeratorHandlers = map[string][]Handler{}
	EdgeCoreHandlers  = map[string][]Handler{}
)

type Handler struct {
	Version              string
	Method               string
	Path                 string
	Func                 gin.HandlerFunc
	IsNotUnderDataCenter bool
}

func (h Handler) IsUnderDataCenter() bool {
	return h.Path != ""
}

func RegisterHandlersToRoles(module string, handlers []Handler, rolesToRegister ...string) {
	for _, role := range rolesToRegister {
		switch role {
		case nodes.RoleControl:
			ControlHandlers[module] = handlers
		case nodes.RoleCompute:
			ComputeHandlers[module] = handlers
		case nodes.RoleStorage:
			StorageHandlers[module] = handlers
		case nodes.RoleModerator:
			ModeratorHandlers[module] = handlers
		case nodes.RoleEdgeCore:
			EdgeCoreHandlers[module] = handlers
		}
	}
}

func GenControlConvergedHandlers() map[string][]Handler {
	controlConvergedHandlers := map[string][]Handler{}

	appendGroupHandlers(controlConvergedHandlers, ControlHandlers)
	appendGroupHandlers(controlConvergedHandlers, ComputeHandlers)
	appendGroupHandlers(controlConvergedHandlers, StorageHandlers)

	return controlConvergedHandlers
}

func GetRoleHandlers(role string) map[string][]Handler {
	switch role {
	case nodes.RoleControl:
		return ControlHandlers
	case nodes.RoleCompute:
		return ComputeHandlers
	case nodes.RoleStorage:
		return StorageHandlers
	case nodes.RoleControlConverged:
		return GenControlConvergedHandlers()
	case nodes.RoleModerator:
		return ModeratorHandlers
	case nodes.RoleEdgeCore:
		return EdgeCoreHandlers
	default:
		return nil
	}
}

func appendGroupHandlers(dstGroupHandlers, srcGroupHandlers map[string][]Handler) {
	maps.Copy(dstGroupHandlers, srcGroupHandlers)
}
