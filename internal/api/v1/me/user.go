package me

import (
	"errors"

	"github.com/crewjam/saml/samlsp"
	"github.com/gin-gonic/gin"
)

func getUsername(c *gin.Context) (string, error) {
	authType, found := c.Get("authType")
	if !found {
		return "", errors.New("the authed method is not support to fetch personal info")
	}

	username := ""
	switch authType.(string) {
	case "saml":
		username = samlsp.AttributeFromContext(c.Request.Context(), "username")
	case "oidc":
		authUser, found := c.Get("authUser")
		if found {
			username = authUser.(string)
		}
	}

	return username, nil
}
