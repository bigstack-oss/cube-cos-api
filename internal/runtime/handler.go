package runtime

import (
	"github.com/bigstack-oss/cube-cos-api/internal/api"
	"github.com/bigstack-oss/cube-cos-api/internal/api/v1/datacenters"
	"github.com/bigstack-oss/cube-cos-api/internal/api/v1/events"
	"github.com/bigstack-oss/cube-cos-api/internal/api/v1/healths"
	"github.com/bigstack-oss/cube-cos-api/internal/api/v1/integrations"
	"github.com/bigstack-oss/cube-cos-api/internal/api/v1/licenses"
	"github.com/bigstack-oss/cube-cos-api/internal/api/v1/logout"
	"github.com/bigstack-oss/cube-cos-api/internal/api/v1/me"
	"github.com/bigstack-oss/cube-cos-api/internal/api/v1/metrics"
	"github.com/bigstack-oss/cube-cos-api/internal/api/v1/nodes"
	"github.com/bigstack-oss/cube-cos-api/internal/api/v1/services"
	"github.com/bigstack-oss/cube-cos-api/internal/api/v1/settings"
	"github.com/bigstack-oss/cube-cos-api/internal/api/v1/supportfiles"
	"github.com/bigstack-oss/cube-cos-api/internal/api/v1/tokens"
	"github.com/bigstack-oss/cube-cos-api/internal/api/v1/triggers"
	apitunings "github.com/bigstack-oss/cube-cos-api/internal/api/v1/tunings"
	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
)

func initNodeApiHandler() {
	api.RegisterHandlersToRoles(
		definition.DataCenters,
		datacenters.Handlers,
		definition.RoleControl,
	)

	api.RegisterHandlersToRoles(
		definition.Services,
		services.Handlers,
		definition.RoleControl,
	)

	api.RegisterHandlersToRoles(
		definition.Me,
		me.Handlers,
		definition.RoleControl,
	)

	api.RegisterHandlersToRoles(
		definition.Integrations,
		integrations.Handlers,
		definition.RoleControl,
	)

	api.RegisterHandlersToRoles(
		definition.Healths,
		healths.Handlers,
		definition.RoleControl,
	)

	api.RegisterHandlersToRoles(
		definition.Events,
		events.Handlers,
		definition.RoleControl,
	)

	api.RegisterHandlersToRoles(
		definition.Nodes,
		nodes.Handlers,
		definition.RoleControl,
	)

	api.RegisterHandlersToRoles(
		definition.Tunings,
		apitunings.Handlers,
		definition.RoleControl,
		definition.RoleCompute,
	)

	api.RegisterHandlersToRoles(
		definition.Metrics,
		metrics.Handlers,
		definition.RoleControl,
	)

	api.RegisterHandlersToRoles(
		definition.Tokens,
		tokens.Handlers,
		definition.RoleControl,
	)

	api.RegisterHandlersToRoles(
		definition.Logout,
		logout.Handlers,
		definition.RoleControl,
	)

	api.RegisterHandlersToRoles(
		definition.Licenses,
		licenses.Handlers,
		definition.RoleControl,
	)

	api.RegisterHandlersToRoles(
		definition.Triggers,
		triggers.Handlers,
		definition.RoleControl,
	)

	api.RegisterHandlersToRoles(
		definition.SupportFiles,
		supportfiles.Handlers,
		definition.RoleControlConverged,
		definition.RoleControl,
		definition.RoleCompute,
		definition.RoleStorage,
		definition.RoleEdgeCore,
		definition.RoleModerator,
	)

	api.RegisterHandlersToRoles(
		definition.Settings,
		settings.Handlers,
		definition.RoleControl,
	)
}
