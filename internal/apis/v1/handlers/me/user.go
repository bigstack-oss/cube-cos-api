package me

import (
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/auths"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/errors"
	"github.com/crewjam/saml/samlsp"
	"github.com/gin-gonic/gin"
)

func getUsername(c *gin.Context) (string, error) {
	authType, found := c.Get("authType")
	if !found {
		return "", errors.ErrAuthMethodCannotGetUserInfo
	}

	username := ""
	switch authType.(string) {
	case auths.Saml:
		username = samlsp.AttributeFromContext(c.Request.Context(), "username")
	case auths.Oidc:
		authUser, found := c.Get("authUser")
		if found {
			username = authUser.(string)
		}
	}

	return username, nil
}
