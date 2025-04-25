package opensearch

import (
	"fmt"

	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/gin-gonic/gin"
)

func genInstanceLink(c *gin.Context) string {
	return fmt.Sprintf(
		"https://%s/opensearch/fake/instance/url/%s",
		v1.DataCenterVip,
		c.Param("instanceId"),
	)
}
