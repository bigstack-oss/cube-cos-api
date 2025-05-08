package nodes

import (
	"errors"
	"net/http"
	"reflect"
	"sync"

	"slices"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	"github.com/bigstack-oss/cube-cos-api/internal/api"
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

func streamingWatcher() {
	for {
		wait.Seconds(2)
		if len(stream.Watchers) == 0 {
			continue
		}

		stream.Lock()
		for _, w := range stream.Watchers {
			resp, err := streamNodeByHandler(&w.helper)
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

func streamNodeByHandler(h *helper) (any, error) {
	switch h.handler {
	case "listNodes":
		return h.listNodes()
	case "getNode":
		return h.getNode()
	}

	return nil, errors.New("no internal function supported")
}

func watchNode(h *helper, data any) {
	setChunkedTransfer(h.c)
	flusher, ok := h.c.Writer.(http.Flusher)
	if !ok {
		api.SetBadRequest(h.c, errors.New("http chunked transfer is not supported"))
		return
	}

	watcher := watcher{helper: *h, dataChan: make(dataChan)}
	addWatcher(watcher)
	defer removeWatcher(watcher)

	sendFirstData(h.c, flusher, data)
	streamingData(h.c, flusher, watcher)
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
			api.SetStatusOk(c, "nodes watching is stopped successfully", nil)
			return
		}
	}
}

func streamingResp(data any) []byte {
	b, err := json.Marshal(gin.H{
		"code":   http.StatusOK,
		"status": "ok",
		"msg":    "fetch node successfully",
		"data":   data,
	})
	if err != nil {
		return []byte{}
	}

	return b
}
