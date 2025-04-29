package api

import (
	"bytes"
	"io"

	"github.com/gin-gonic/gin"
	log "go-micro.dev/v5/logger"
)

func ParseBody(c *gin.Context) []byte {
	body, err := c.GetRawData()
	if err != nil {
		log.Errorf("api: failed to parse request body: %v", err)
		return nil
	}

	// in the gin context, the request body will be closed after reading
	// so, we need to set it back to the request body if we want to read the Request.Body again by the gin context
	c.Request.Body = io.NopCloser(bytes.NewReader(body))
	return body
}
