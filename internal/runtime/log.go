package runtime

import (
	"time"

	"github.com/gin-gonic/gin"
	log "go-micro.dev/v5/logger"
)

var (
	count                   int
	thresholdToLogLiveCheck = 300
)

func logLiveCheck(c *gin.Context, reqId string, elapsed time.Duration) {
	count++
	if count < thresholdToLogLiveCheck {
		return
	}

	count = 0
	log.Infof(
		"req(%s): %s (%s)",
		reqId,
		genReqMsg(c),
		elapsed,
	)
}
