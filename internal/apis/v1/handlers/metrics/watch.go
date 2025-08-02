package metrics

import (
	"errors"
	"net/http"
	"reflect"
	"sync"

	"slices"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/bodies"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/gin-gonic/gin"
	json "github.com/json-iterator/go"
)

type dataChan chan any

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

func streamWatchers() {
	for {
		wait.Seconds(2)
		if len(stream.Watchers) == 0 {
			continue
		}

		stream.Lock()
		for _, w := range stream.Watchers {
			resp, err := streamMetricsByHandler(&w.helper)
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

func streamMetricsByHandler(h *helper) (any, error) {
	switch h.handler {
	case "getDataCenterSummary":
		return cubecos.GetMetricsSummary(), nil
	case "getMetrics":
		return h.getMetrics()
	default:
		return nil, errors.New("no internal function supported")
	}
}

func watchHealth(h *helper, health any) {
	setChunkedTransfer(h.c)
	flusher, ok := h.c.Writer.(http.Flusher)
	if !ok {
		bodies.SetBadRequest(h.c, errors.New("http chunked transfer is not supported"))
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
		if !reflect.DeepEqual(watcher, watcherToRemove) {
			continue
		}

		stream.Watchers = slices.Delete(stream.Watchers, i, i+1)
		return
	}
}

func sendFirstHealth(c *gin.Context, flusher http.Flusher, health any) {
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
			bodies.SetOk(c, "health watching is stopped successfully", nil)
			return
		}
	}
}

func streamingResp(health any) []byte {
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
