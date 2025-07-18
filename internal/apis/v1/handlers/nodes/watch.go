package nodes

import (
	"errors"
	"net/http"
	"slices"
	"sync"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/bodies"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
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

func streamData(h *helper, data any) {
	setChunkedTransfer(h.c)
	flusher, ok := h.c.Writer.(http.Flusher)
	if !ok {
		bodies.SetBadRequest(h.c, errors.New("http chunked transfer is not supported"))
		return
	}

	watcher := watcher{helper: *h, dataChan: make(dataChan)}
	setWatcher(watcher)
	defer removeWatcher(watcher)

	sendFirstData(h.c, flusher, data)
	streamingData(h.c, flusher, watcher)
}

func streamingWatcher() {
	for {
		if len(stream.Watchers) == 0 {
			wait.Seconds(2)
			continue
		}

		change, shutdown := changes.Get()
		if shutdown {
			return
		}

		stream.Lock()
		for _, w := range stream.Watchers {
			data, err := streamDataByHandler(&w.helper, change)
			changes.Done(change)
			if err != nil {
				continue
			}

			select {
			case w.dataChan <- data:
			default:
			}
		}

		stream.Unlock()
	}
}

func streamDataByHandler(h *helper, change nodes.Change) (any, error) {
	switch h.handler {
	case "listNodes":
		return h.listNodes()
	case "getNode":
		return h.getNode()
	case "listNodeDevices":
		return h.listNodeDevices(change.UseCacheInStream)
	}

	return nil, errors.New("no internal function supported")
}

func setChunkedTransfer(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")
}

func setWatcher(w watcher) {
	stream.Lock()
	stream.Watchers = append(stream.Watchers, w)
	stream.Unlock()
}

func removeWatcher(watcherToRemove watcher) {
	stream.Lock()
	defer stream.Unlock()

	for i, watcher := range stream.Watchers {
		if watcher.reqId != watcherToRemove.reqId {
			continue
		}

		stream.Watchers = slices.Delete(stream.Watchers, i, i+1)
		return
	}
}

func sendFirstData(c *gin.Context, flusher http.Flusher, nodes any) {
	c.Writer.Write(streamingResp(nodes))
	c.Writer.Write([]byte("\n"))
	flusher.Flush()
}

func streamingData(c *gin.Context, flusher http.Flusher, watcher watcher) {
	ctx := c.Request.Context()
	for {
		select {
		case nodes := <-watcher.dataChan:
			c.Writer.Write(streamingResp(nodes))
			c.Writer.Write([]byte("\n"))
			flusher.Flush()
		case <-ctx.Done():
			bodies.SetOk(c, "nodes watching is stopped successfully", nil)
			return
		}
	}
}

func streamingResp(data any) []byte {
	b, err := json.Marshal(gin.H{
		"code":   http.StatusOK,
		"status": "ok",
		"msg":    "fetch streaming data successfully",
		"data":   data,
	})
	if err != nil {
		return []byte{}
	}

	return b
}
