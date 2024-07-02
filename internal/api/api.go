package api

import (
	"fmt"

	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
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
	NetworkHandlers   = map[string][]Handler{}
	ModeratorHandlers = map[string][]Handler{}
	EdgeCoreHandlers  = map[string][]Handler{}
)

type Handler struct {
	Version string
	Method  string
	Path    string
	Func    gin.HandlerFunc
}

func RegisterHandlersToRoles(module string, handlers []Handler, rolesToRegister ...string) {
	for _, role := range rolesToRegister {
		switch role {
		case definition.RoleControl:
			ControlHandlers[module] = handlers
		case definition.RoleCompute:
			ComputeHandlers[module] = handlers
		case definition.RoleStorage:
			StorageHandlers[module] = handlers
		case definition.RoleNetwork:
			NetworkHandlers[module] = handlers
		case definition.RoleModerator:
			ModeratorHandlers[module] = handlers
		case definition.RoleEdgeCore:
			EdgeCoreHandlers[module] = handlers
		}
	}
}

func appendGroupHandlers(dstGroupHandlers, srcGroupHandlers map[string][]Handler) {
	for name, handlers := range srcGroupHandlers {
		dstGroupHandlers[name] = handlers
	}
}

func GenControlConvergedHandlers() map[string][]Handler {
	controlConvergedHandlers := map[string][]Handler{}

	appendGroupHandlers(controlConvergedHandlers, ControlHandlers)
	appendGroupHandlers(controlConvergedHandlers, ComputeHandlers)
	appendGroupHandlers(controlConvergedHandlers, StorageHandlers)
	appendGroupHandlers(controlConvergedHandlers, NetworkHandlers)

	return controlConvergedHandlers
}

func GetGroupHandlersByRole(role string) map[string][]Handler {
	switch role {
	case definition.RoleControl:
		return ControlHandlers
	case definition.RoleCompute:
		return ComputeHandlers
	case definition.RoleStorage:
		return StorageHandlers
	case definition.RoleNetwork:
		return NetworkHandlers
	case definition.RoleControlConverged:
		return GenControlConvergedHandlers()
	case definition.RoleModerator:
		return ModeratorHandlers
	case definition.RoleEdgeCore:
		return EdgeCoreHandlers
	default:
		return nil
	}
}
