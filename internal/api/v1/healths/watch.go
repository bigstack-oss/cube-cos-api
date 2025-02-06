package healths

import (
	"errors"
	"net/http"
	"sync"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	"github.com/bigstack-oss/cube-cos-api/internal/api"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/gin-gonic/gin"
	json "github.com/json-iterator/go"
)

type watcher chan cubecos.Health

var (
	stream = struct {
		sync.Mutex
		Watchers []watcher
	}{}
)

func streamHealthSummary() {
	for {
		wait.Seconds(2)
		if len(stream.Watchers) == 0 {
			continue
		}

		summary := genFakeHealthSummary()
		stream.Lock()
		for _, w := range stream.Watchers {
			select {
			case w <- summary:
			default:
			}
		}

		stream.Unlock()
	}
}

func watchHealthSummary(c *gin.Context, summary *cubecos.Health) {
	setChunkedTransfer(c)
	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		api.SetBadRequest(c, errors.New("http chunked transfer is not supported"))
		return
	}

	watcher := make(watcher)
	addWatcher(watcher)
	defer removeWatcher(watcher)

	sendFirstSummary(c, flusher, summary)
	streamingSummary(c, flusher, watcher)
}

func setChunkedTransfer(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")
}

func addWatcher(w watcher) {
	stream.Lock()
	stream.Watchers = append(stream.Watchers, w)
	stream.Unlock()
}

func removeWatcher(watcherToRemove watcher) {
	stream.Lock()
	defer stream.Unlock()

	for i, watcher := range stream.Watchers {
		if watcher != watcherToRemove {
			continue
		}

		stream.Watchers = append(
			stream.Watchers[:i],
			stream.Watchers[i+1:]...,
		)
		return
	}
}

func sendFirstSummary(c *gin.Context, flusher http.Flusher, healthSummary *cubecos.Health) {
	c.Writer.Write(streamingResp(healthSummary))
	c.Writer.Write([]byte("\n"))
	flusher.Flush()
}

func streamingSummary(c *gin.Context, flusher http.Flusher, watcher watcher) {
	ctx := c.Request.Context()
	for {
		select {
		case healthSummary := <-watcher:
			c.Writer.Write(streamingResp(&healthSummary))
			c.Writer.Write([]byte("\n"))
			flusher.Flush()
		case <-ctx.Done():
			api.SetStatusOk(c, "health summary watching is stopped successfully", nil)
			return
		}
	}
}

func streamingResp(healthSummary *cubecos.Health) []byte {
	b, err := json.Marshal(gin.H{
		"code":   http.StatusOK,
		"status": "ok",
		"msg":    "fetch health summary successfully",
		"data":   *healthSummary,
	})
	if err != nil {
		return []byte{}
	}

	return b
}
