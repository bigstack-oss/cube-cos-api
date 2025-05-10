package cubecos

import (
	"context"

	"github.com/Nerzal/gocloak/v13"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/keycloak"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/auth"
)

func CreateToken(user *auth.User) (*gocloak.JWT, error) {
	h := keycloak.GetGlobalHelper()
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(5))
	defer cancel()
	return h.Login(
		ctx,
		auth.DefaultOidcClientId,
		auth.DefaultOidcClientSecret,
		auth.DefaultKeycloakRealm,
		user.Name,
		user.Password,
	)
}
