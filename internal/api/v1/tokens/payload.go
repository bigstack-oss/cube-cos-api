package tokens

import (
	"encoding/json"
	"errors"

	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/gin-gonic/gin"
)

func parseUserBody(c *gin.Context) (*v1.User, error) {
	u := &v1.User{}
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
