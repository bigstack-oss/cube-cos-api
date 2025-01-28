package tokens

import (
	"encoding/json"
	"errors"

	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/gin-gonic/gin"
)

func parseUserBody(c *gin.Context) (*definition.User, error) {
	u := &definition.User{}
	err := json.NewDecoder(c.Request.Body).Decode(&u)
	if err != nil {
		return nil, err
	}

	if u.IsNameEmpty() {
		return nil, errors.New("user name is empty")
	}

	if u.IsPasswordEmpty() {
		return nil, errors.New("user password is empty")
	}

	return u, nil
}
