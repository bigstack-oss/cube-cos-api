package tokens

import (
	"errors"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/auth"
	"github.com/gin-gonic/gin"
	json "github.com/json-iterator/go"
)

func parseUserBody(c *gin.Context) (*auth.User, error) {
	user := &auth.User{}
	err := json.NewDecoder(c.Request.Body).Decode(&user)
	if err != nil {
		return nil, err
	}

	if user.IsNameEmpty() {
		return nil, errors.New("user name is empty")
	}

	if user.IsPasswordEmpty() {
		return nil, errors.New("user password is empty")
	}

	return user, nil
}
