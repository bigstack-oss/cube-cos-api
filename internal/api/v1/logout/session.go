package logout

import (
	"fmt"
	"strings"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/keycloak"
	conf "github.com/bigstack-oss/cube-cos-api/internal/config"
	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/bigstack-oss/cube-cos-api/internal/saml"
	"github.com/crewjam/saml/samlsp"
	"github.com/gin-gonic/gin"
	log "go-micro.dev/v5/logger"
)

func genRedirectUrl() string {
	return fmt.Sprintf(
		"https://%s:%d%s",
		definition.ControllerVip,
		conf.Opts.Spec.Saml.ServiceProvider.Host.Port,
		conf.Opts.Spec.Identity.LogoutRedirect,
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
	h := keycloak.GetGlobalHelper()
	err := h.LoginAdmin()
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
