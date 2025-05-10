package logout

import (
	"fmt"
	"strings"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/keycloak"
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/queries"
	"github.com/bigstack-oss/cube-cos-api/internal/auths/saml"
	conf "github.com/bigstack-oss/cube-cos-api/internal/config"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/base"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/errors"
	"github.com/crewjam/saml/samlsp"
	log "go-micro.dev/v5/logger"
)

func genRedirectUrl() string {
	return fmt.Sprintf(
		"https://%s:%d%s",
		base.DataCenterVip,
		conf.Opts.Spec.Saml.ServiceProvider.Host.Port,
		conf.Opts.Spec.Identity.Redirect,
	)
}

func (h *helper) cleanSession(session *samlsp.Session) error {
	claims := (*session).(samlsp.JWTSessionClaims)
	err := h.deleteSamlSession(claims)
	if err != nil {
		log.Errorf("logout(%s): failed to delete saml session: %s", queries.GetReqId(h.c), err.Error())
		return err
	}

	err = h.deleteKeycloakSession(claims)
	if err != nil {
		log.Errorf("logout(%s): failed to delete keycloak session: %s", queries.GetReqId(h.c), err.Error())
		return err
	}

	return nil
}

func (h *helper) deleteSamlSession(jwtSession samlsp.JWTSessionClaims) error {
	_, err := saml.SpAuth.ServiceProvider.MakeRedirectLogoutRequest(jwtSession.Subject, "")
	if err != nil {
		log.Errorf("logout(%s): failed to get signout url: %s", queries.GetReqId(h.c), err.Error())
		return err
	}

	err = saml.SpAuth.Session.DeleteSession(h.c.Writer, h.c.Request)
	if err != nil {
		log.Errorf("logout(%s): failed to delete saml session: %s", queries.GetReqId(h.c), err.Error())
		return err
	}

	return nil
}

func (h *helper) deleteKeycloakSession(claims samlsp.JWTSessionClaims) error {
	keycloak := keycloak.GetGlobalHelper()
	err := keycloak.LoginAdmin()
	if err != nil {
		log.Errorf("logout(%s): failed to login admin: %s", err.Error())
		return err
	}

	sessions, found := claims.Attributes["SessionIndex"]
	if !found {
		log.Errorf("logout(%s): session index not found in jwt session", queries.GetReqId(h.c))
		return errors.ErrSessionIndexNotFound
	}

	for _, session := range sessions {
		fragments := strings.Split(session, "::")
		if len(fragments) < 2 {
			log.Warnf("logout(%s): invalid alive session index to logout: %s", queries.GetReqId(h.c), fragments)
			continue
		}

		sessionId := fragments[0]
		err := keycloak.LogoutUserSession(conf.Opts.Spec.Identity.Keycloak.Realm, sessionId)
		if err != nil {
			log.Errorf("logout(%s): failed to logout user(%s): %s", queries.GetReqId(h.c), sessionId, err.Error())
		}
	}

	return nil
}
