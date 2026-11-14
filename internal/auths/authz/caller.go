package authz

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/openstack/v2"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/queries"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/auths"
	"github.com/crewjam/saml/samlsp"
	"github.com/gin-gonic/gin"
	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/openstack/identity/v3/tokens"
	log "go-micro.dev/v5/logger"
)

const callerKey = "caller"

var (
	adminRoles = []string{"admin", "service"}

	// ErrUnavailable means the caller's roles could not be looked up, so
	// whether the caller may proceed is unknown.
	ErrUnavailable = errors.New("cannot look up the caller's roles")
)

// Caller is who sent the request and what the caller may administer.
type Caller struct {
	User        string
	IsNode      bool
	IsAdmin     bool
	Roles       []string
	ProjectId   string
	ProjectName string
	DomainId    string
	DomainName  string
}

// Resolve identifies the caller of an authenticated request. The result is
// kept in the request context, so it is worked out once per request.
func Resolve(c *gin.Context) (Caller, error) {
	value, found := c.Get(callerKey)
	if found {
		return value.(Caller), nil
	}

	caller, err := resolve(c)
	if err != nil {
		return caller, err
	}

	c.Set(callerKey, caller)
	return caller, nil
}

func resolve(c *gin.Context) (Caller, error) {
	if c.GetBool("isNode") {
		return Caller{
			User:    fmt.Sprintf("node:%s", c.GetHeader("Node")),
			IsNode:  true,
			IsAdmin: true,
		}, nil
	}

	switch c.GetString("authType") {
	case auths.Openstack:
		return resolveOpenstackCaller(c)
	case auths.Oidc:
		return resolveKeycloakCaller(c, c.GetString("authUser"))
	case auths.Saml:
		return resolveKeycloakCaller(c, samlsp.AttributeFromContext(c.Request.Context(), "username"))
	default:
		return Caller{}, errors.New("unknown authentication type")
	}
}

// openstackToken is a variable so tests can stand in for Keystone.
var openstackToken = func(token string) tokens.GetResult {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(30))
	defer cancel()

	return tokens.Get(ctx, openstack.GetGlobalHelper().Identity, token)
}

func resolveOpenstackCaller(c *gin.Context) (Caller, error) {
	reqId := queries.GetReqId(c)
	token := openstackToken(c.GetHeader("X-Auth-Token"))
	if token.Err != nil && !gophercloud.ResponseCodeIs(token.Err, http.StatusNotFound) {
		log.Errorf("authz(%s): failed to look up openstack token(%v)", reqId, token.Err)
		return Caller{}, ErrUnavailable
	}

	user, err := token.ExtractUser()
	if err != nil {
		log.Errorf("authz(%s): failed to resolve openstack token(%v)", reqId, err)
		return Caller{}, errors.New("cannot resolve the openstack token")
	}

	roles, err := token.ExtractRoles()
	if err != nil {
		log.Errorf("authz(%s): failed to resolve openstack token roles(%v)", reqId, err)
		return Caller{}, errors.New("cannot resolve the openstack token")
	}

	project, err := token.ExtractProject()
	if err != nil {
		log.Errorf("authz(%s): failed to resolve openstack token project(%v)", reqId, err)
		return Caller{}, errors.New("cannot resolve the openstack token")
	}

	caller := Caller{User: user.Name}
	for _, role := range roles {
		caller.Roles = append(caller.Roles, role.Name)
	}

	if project != nil {
		caller.ProjectId = project.ID
		caller.ProjectName = project.Name
		caller.DomainId = project.Domain.ID
		caller.DomainName = project.Domain.Name
	}

	hasAdminRole := slices.ContainsFunc(caller.Roles, func(role string) bool {
		return slices.Contains(adminRoles, role)
	})
	if !hasAdminRole || caller.ProjectId == "" {
		return caller, nil
	}

	adminProject, err := adminProjectId()
	if err != nil {
		log.Errorf("authz(%s): failed to find the admin project(%v)", reqId, err)
		return Caller{}, ErrUnavailable
	}

	caller.IsAdmin = caller.ProjectId == adminProject
	return caller, nil
}

// adminProjectId returns the project the API itself signs in to Keystone
// with, which is the cluster's admin project. Roles held in any other
// project do not make the caller an admin. It is a variable so tests can
// stand in for Keystone.
var adminProjectId = func() (string, error) {
	result, ok := openstack.GetGlobalHelper().Provider.GetAuthResult().(interface {
		ExtractProject() (*tokens.Project, error)
	})
	if !ok {
		return "", errors.New("the api's openstack sign-in result is not available")
	}

	project, err := result.ExtractProject()
	if err != nil {
		return "", err
	}

	if project == nil {
		return "", errors.New("the api's openstack token is not scoped to a project")
	}

	return project.ID, nil
}

func resolveKeycloakCaller(c *gin.Context, username string) (Caller, error) {
	caller := Caller{User: username}
	if username == "" {
		return caller, nil
	}

	isAdmin, err := IsRealmAdmin(username)
	if err != nil {
		log.Errorf("authz(%s): failed to look up realm roles of user %s(%v)", queries.GetReqId(c), username, err)
		return caller, ErrUnavailable
	}

	caller.IsAdmin = isAdmin
	return caller, nil
}
