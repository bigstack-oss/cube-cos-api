package firmwares

import (
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/mongo"
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/queries"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/pages"
	"github.com/gin-gonic/gin"
)

type helper struct {
	c       *gin.Context
	reqId   string
	handler string
	mongo   *mongo.Helper

	file string
	page *pages.Page
}

func initHelper(c *gin.Context, handler string) (*helper, error) {
	h := &helper{
		c:       c,
		mongo:   mongo.GetGlobalHelper(),
		reqId:   queries.GetReqId(c),
		handler: handler,
	}

	return h, h.parseParamsByHandler()
}
