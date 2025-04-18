package runtime

import (
	"github.com/bigstack-oss/cube-cos-api/internal/api"
	"github.com/bigstack-oss/cube-cos-api/internal/api/v1/datacenters"
	"github.com/bigstack-oss/cube-cos-api/internal/api/v1/events"
	"github.com/bigstack-oss/cube-cos-api/internal/api/v1/grafana"
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
	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/support"
)

// M1 TODO:
// recheck the role assignments for the handlers
func registerNodeApiHandler() {
	api.RegisterHandlersToRoles(
		v1.DataCenters,
		datacenters.Handlers,
		v1.RoleControl,
		v1.RoleControlConverged,
	)

	api.RegisterHandlersToRoles(
		v1.Services,
		services.Handlers,
		v1.RoleControl,
		v1.RoleControlConverged,
	)

	api.RegisterHandlersToRoles(
		v1.Me,
		me.Handlers,
		v1.RoleControl,
		v1.RoleControlConverged,
	)

	api.RegisterHandlersToRoles(
		v1.Integrations,
		integrations.Handlers,
		v1.RoleControl,
		v1.RoleControlConverged,
	)

	api.RegisterHandlersToRoles(
		v1.Healths,
		healths.Handlers,
		v1.RoleControl,
	)

	api.RegisterHandlersToRoles(
		v1.Events,
		events.Handlers,
		v1.RoleControl,
	)

	api.RegisterHandlersToRoles(
		v1.Nodes,
		nodes.Handlers,
		v1.RoleControl,
	)

	api.RegisterHandlersToRoles(
		v1.Tunings,
		apitunings.Handlers,
		v1.RoleControl,
		v1.RoleCompute,
	)

	api.RegisterHandlersToRoles(
		v1.Metrics,
		metrics.Handlers,
		v1.RoleControl,
	)

	api.RegisterHandlersToRoles(
		v1.Tokens,
		tokens.Handlers,
		v1.RoleControl,
	)

	api.RegisterHandlersToRoles(
		v1.Logout,
		logout.Handlers,
		v1.RoleControl,
	)

	api.RegisterHandlersToRoles(
		v1.Licenses,
		licenses.Handlers,
		v1.RoleControl,
	)

	api.RegisterHandlersToRoles(
		v1.Triggers,
		triggers.Handlers,
		v1.RoleControl,
	)

	api.RegisterHandlersToRoles(
		support.Files,
		supportfiles.Handlers,
		v1.RoleControlConverged,
		v1.RoleControl,
		v1.RoleCompute,
		v1.RoleStorage,
		v1.RoleEdgeCore,
		v1.RoleModerator,
	)

	api.RegisterHandlersToRoles(
		v1.Grafana,
		grafana.Handlers,
		v1.RoleControl,
		v1.RoleControlConverged,
	)

	api.RegisterHandlersToRoles(
		v1.Settings,
		settings.Handlers,
		v1.RoleControl,
	)
}
