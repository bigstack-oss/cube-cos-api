package cubecos

import (
	"context"

	"github.com/Nerzal/gocloak/v13"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/keycloak"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/auths"
)

func CreateToken(user *auths.User) (*gocloak.JWT, error) {
	h := keycloak.GetGlobalHelper()
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(5))
	defer cancel()
	return h.Login(
		ctx,
		auths.DefaultOidcClientId,
		auths.DefaultOidcClientSecret,
		auths.DefaultKeycloakRealm,
		user.Name,
		user.Password,
	)
}
