package logout

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/keycloak"
	"github.com/bigstack-oss/cube-cos-api/internal/api"
	conf "github.com/bigstack-oss/cube-cos-api/internal/config"
	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/bigstack-oss/cube-cos-api/internal/saml"
	"github.com/crewjam/saml/samlsp"
	"github.com/gin-gonic/gin"
	log "go-micro.dev/v5/logger"
)

var (
	Handlers = []api.Handler{
		{
			Version:              api.V1,
			Method:               http.MethodGet,
			Path:                 "/logout",
			Func:                 logout,
			IsNotUnderDataCenter: true,
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
	session, err := saml.SpAuth.Session.GetSession(c.Request)
	if err != nil {
		log.Errorf("failed to get session for logout: %s", err.Error())
		c.JSON(
			http.StatusInternalServerError,
			gin.H{
				"code":   http.StatusInternalServerError,
				"status": "internal server error",
				"msg":    err.Error(),
			},
		)
		return
	}

	err = cleanSession(c, session)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			gin.H{
				"code":   http.StatusInternalServerError,
				"status": "internal server error",
				"msg":    err.Error(),
			},
		)
		return
	}

	c.Redirect(
		http.StatusFound,
		fmt.Sprintf("https://%s:4443%s", definition.ControllerVip, conf.Opts.Spec.Identity.LogoutRedirect),
	)
}

func cleanSession(c *gin.Context, session samlsp.Session) error {
	jwtSession := session.(samlsp.JWTSessionClaims)
	err := deleteSessionInSamlAuth(c, jwtSession)
	if err != nil {
		log.Errorf("failed to delete session in saml auth: %s", err.Error())
		return err
	}

	err = deleteSessionInKeycloak(jwtSession)
	if err != nil {
		log.Errorf("failed to delete session in keycloak: %s", err.Error())
		return err
	}

	return nil
}

func deleteSessionInSamlAuth(c *gin.Context, jwtSession samlsp.JWTSessionClaims) error {
	_, err := saml.SpAuth.ServiceProvider.MakeRedirectLogoutRequest(jwtSession.Subject, "")
	if err != nil {
		log.Errorf("failed to get signout url: %s", err.Error())
		return err
	}

	err = saml.SpAuth.Session.DeleteSession(c.Writer, c.Request)
	if err != nil {
		log.Errorf("failed to delete session for logout: %s", err.Error())
		return err
	}

	return nil
}

func deleteSessionInKeycloak(jwtSession samlsp.JWTSessionClaims) error {
	h, err := keycloak.NewHelper(
		keycloak.Host(conf.Opts.Spec.Identity.Keycloak.Host),
		keycloak.Realm(conf.Opts.Spec.Identity.Keycloak.Realm),
		keycloak.Username(conf.Opts.Spec.Identity.Keycloak.Username),
		keycloak.Password(conf.Opts.Spec.Identity.Keycloak.Password),
		keycloak.Insecure(conf.Opts.Spec.Identity.Keycloak.TlsInsecureSkipVerify),
	)
	if err != nil {
		log.Errorf("failed to create keycloak helper: %s", err.Error())
		return err
	}

	err = h.LoginAdmin()
	if err != nil {
		log.Errorf("failed to login admin: %s", err.Error())
		return err
	}

	activeSpSessions, found := jwtSession.Attributes["SessionIndex"]
	if !found {
		err := fmt.Errorf("session index not found in jwt session")
		log.Errorf(err.Error())
		return err
	}

	for _, activeSpSession := range activeSpSessions {
		spSessionInfo := strings.Split(activeSpSession, "::")
		if len(spSessionInfo) < 2 {
			log.Warnf("invalid alive session index to logout: %s", spSessionInfo)
			continue
		}

		sessionId := spSessionInfo[0]
		err := h.LogoutUserSession(conf.Opts.Spec.Identity.Keycloak.Realm, sessionId)
		if err != nil {
			log.Errorf("failed to logout user session(%s): %s", sessionId, err.Error())
		}
	}

	return nil
}
