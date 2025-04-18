package cubecos

import (
	"context"
	"time"

	"github.com/Nerzal/gocloak/v13"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/keycloak"
	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
)

func CreateToken(user *v1.User) (*gocloak.JWT, error) {
	h := keycloak.GetGlobalHelper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return h.Login(
		ctx,
		v1.DefaultOidcClientId,
		v1.DefaultOidcClientSecret,
		v1.DefaultKeycloakRealm,
		user.Name,
		user.Password,
	)
}
