package login

import (
	"net/http"

	"github.com/bigstack-oss/cube-cos-api/internal/api"
	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/bigstack-oss/cube-cos-api/internal/keycloak"
	"github.com/crewjam/saml/samlsp"
	"github.com/gin-gonic/gin"
)

var (
	Handlers = []api.Handler{
		{
			Version: api.V1,
			Method:  http.MethodPost,
			Path:    "/logout",
			Func:    logout,
		},
	}
)

// @BasePath /api/v1
// @Summary	Logout from the system and redirect to login page
// @Schemes
// @Description
// @Tags		logout      specifications
// @Success	302	{array}     string	""
// @Failure	500	{string}	string	""
// @Router		/logout     [post]
func logout(c *gin.Context) {
	nameID := samlsp.AttributeFromContext(c, definition.DefaultIdentifierFormat)
	_, err := keycloak.SamlAuth.ServiceProvider.MakeRedirectLogoutRequest(nameID, "")
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			gin.H{"error": err.Error()},
		)
		return
	}

	err = keycloak.SamlAuth.Session.DeleteSession(c.Writer, c.Request)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			gin.H{"error": err.Error()},
		)
		return
	}

	c.Redirect(302, "/login")
}
