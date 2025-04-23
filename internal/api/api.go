package api

import (
	"errors"
	"fmt"
	"strconv"

	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/gin-gonic/gin"
)

const (
	API = "api"

	Code   = "code"
	Status = "status"
	Msg    = "msg"
	Data   = "data"
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
		case v1.RoleControl:
			ControlHandlers[module] = handlers
		case v1.RoleCompute:
			ComputeHandlers[module] = handlers
		case v1.RoleStorage:
			StorageHandlers[module] = handlers
		case v1.RoleModerator:
			ModeratorHandlers[module] = handlers
		case v1.RoleEdgeCore:
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

func GetRoleHandlers(role string) map[string][]Handler {
	switch role {
	case v1.RoleControl:
		return ControlHandlers
	case v1.RoleCompute:
		return ComputeHandlers
	case v1.RoleStorage:
		return StorageHandlers
	case v1.RoleControlConverged:
		return GenControlConvergedHandlers()
	case v1.RoleModerator:
		return ModeratorHandlers
	case v1.RoleEdgeCore:
		return EdgeCoreHandlers
	default:
		return nil
	}
}

func GetReqId(c *gin.Context) string {
	id, found := c.Get("reqId")
	if !found {
		return ""
	}

	return id.(string)
}

func ParseWatch(c *gin.Context) (bool, error) {
	rawParam := c.DefaultQuery("watch", "false")
	watch, err := strconv.ParseBool(rawParam)
	if err != nil {
		return false, errors.New("watch parameter is invalid, it should be true or false if provided")
	}

	return watch, nil
}
