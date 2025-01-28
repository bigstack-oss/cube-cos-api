package cubecos

import (
	"context"
	"time"

	"github.com/Nerzal/gocloak/v13"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/keycloak"
	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
)

func CreateToken(user *definition.User) (*gocloak.JWT, error) {
	h := keycloak.GetGlobalHelper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return h.Login(
		ctx,
		definition.DefaultOidcClientId,
		definition.DefaultOidcClientSecret,
		definition.DefaultKeycloakRealm,
		user.Name,
		user.Password,
	)
}
