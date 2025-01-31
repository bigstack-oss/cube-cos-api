package runtime

import (
	"github.com/bigstack-oss/cube-cos-api/internal/api"
	"github.com/bigstack-oss/cube-cos-api/internal/api/v1/datacenters"
	"github.com/bigstack-oss/cube-cos-api/internal/api/v1/events"
	"github.com/bigstack-oss/cube-cos-api/internal/api/v1/healths"
	"github.com/bigstack-oss/cube-cos-api/internal/api/v1/integrations"
	"github.com/bigstack-oss/cube-cos-api/internal/api/v1/logout"
	"github.com/bigstack-oss/cube-cos-api/internal/api/v1/metrics"
	"github.com/bigstack-oss/cube-cos-api/internal/api/v1/nodes"
	"github.com/bigstack-oss/cube-cos-api/internal/api/v1/tokens"
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
}
