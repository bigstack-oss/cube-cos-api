package metrics

import (
	"errors"
	"net/http"
	"strconv"
	"sync"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	"github.com/bigstack-oss/cube-cos-api/internal/api"
	"github.com/gin-gonic/gin"
	json "github.com/json-iterator/go"
)

type dataChan chan interface{}

type watcher struct {
	helper
	dataChan
}

var (
	stream = struct {
		sync.Mutex
		Watchers []watcher
	}{}
)

func streamMetrics() {
	for {
		wait.Seconds(2)
		if len(stream.Watchers) == 0 {
			continue
		}

		stream.Lock()
		for _, w := range stream.Watchers {
			resp, err := streamMetricsByHandlerType(&w.helper)
			if err != nil {
				continue
			}

			select {
			case w.dataChan <- resp:
			default:
			}
		}

		stream.Unlock()
	}
}

func streamMetricsByHandlerType(h *helper) (interface{}, error) {
	switch h.handler {
	case "getDataCenterSummary":
		return h.getDataCenterSummary()
	case "getMetrics":
		return h.getMetrics()
	}

	return nil, errors.New("no internal function supported")
}

func parseWatch(c *gin.Context) (bool, error) {
	rawParam := c.DefaultQuery("watch", "false")
	watch, err := strconv.ParseBool(rawParam)
	if err != nil {
		return false, errors.New("watch parameter is invalid, it should be true or false if provided")
	}

	return watch, nil
}

func watchHealth(h *helper, health interface{}) {
	setChunkedTransfer(h.c)
	flusher, ok := h.c.Writer.(http.Flusher)
	if !ok {
		api.SetBadRequest(h.c, errors.New("http chunked transfer is not supported"))
		return
	}

	watcher := watcher{helper: *h, dataChan: make(dataChan)}
	addWatcher(watcher)
	defer removeWatcher(watcher)

	sendFirstHealth(h.c, flusher, health)
	streamingHealth(h.c, flusher, watcher)
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

func sendFirstHealth(c *gin.Context, flusher http.Flusher, health interface{}) {
	c.Writer.Write(streamingResp(health))
	c.Writer.Write([]byte("\n"))
	flusher.Flush()
}

func streamingHealth(c *gin.Context, flusher http.Flusher, watcher watcher) {
	ctx := c.Request.Context()
	for {
		select {
		case health := <-watcher.dataChan:
			c.Writer.Write(streamingResp(&health))
			c.Writer.Write([]byte("\n"))
			flusher.Flush()
		case <-ctx.Done():
			api.SetStatusOk(c, "health watching is stopped successfully", nil)
			return
		}
	}
}

func streamingResp(health interface{}) []byte {
	b, err := json.Marshal(gin.H{
		"code":   http.StatusOK,
		"status": "ok",
		"msg":    "fetch data center health successfully",
		"data":   health,
	})
	if err != nil {
		return []byte{}
	}

	return b
}
