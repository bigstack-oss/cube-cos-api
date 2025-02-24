package tunings

import (
	"encoding/json"
	"errors"
	"net/http"
	"reflect"
	"sync"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	"github.com/bigstack-oss/cube-cos-api/internal/api"
	"github.com/gin-gonic/gin"
)

type dataChan chan data

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

func streamTunings() {
	for {
		wait.Seconds(2)
		if len(stream.Watchers) == 0 {
			continue
		}

		stream.Lock()
		for _, w := range stream.Watchers {
			resp, err := streamTuningsByHandlerType(&w.helper)
			if err != nil {
				continue
			}

			select {
			case w.dataChan <- *resp:
			default:
			}
		}

		stream.Unlock()
	}
}

func streamTuningsByHandlerType(h *helper) (*data, error) {
	switch h.handler {
	case "getTunings":
		return h.ListTunings()
	}

	return nil, errors.New("no internal function supported")
}

func watchTunings(h *helper, data *data) {
	setChunkedTransfer(h.c)
	flusher, ok := h.c.Writer.(http.Flusher)
	if !ok {
		api.SetBadRequest(h.c, errors.New("http chunked transfer is not supported"))
		return
	}

	watcher := watcher{helper: *h, dataChan: make(dataChan)}
	addWatcher(watcher)
	defer removeWatcher(watcher)

	sendFirstTuning(h.c, flusher, data)
	streamingTuning(h.c, flusher, watcher)
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

		stream.Watchers = append(
			stream.Watchers[:i],
			stream.Watchers[i+1:]...,
		)
		return
	}
}

func sendFirstTuning(c *gin.Context, flusher http.Flusher, data *data) {
	c.Writer.Write(streamingResp(data))
	c.Writer.Write([]byte("\n"))
	flusher.Flush()
}

func streamingTuning(c *gin.Context, flusher http.Flusher, watcher watcher) {
	ctx := c.Request.Context()
	for {
		select {
		case data := <-watcher.dataChan:
			c.Writer.Write(streamingResp(&data))
			c.Writer.Write([]byte("\n"))
			flusher.Flush()
		case <-ctx.Done():
			api.SetStatusOk(c, "tuning watching is stopped successfully", nil)
			return
		}
	}
}

func streamingResp(data *data) []byte {
	b, err := json.Marshal(gin.H{
		"code":   http.StatusOK,
		"status": "ok",
		"msg":    "fetch tuning successfully",
		"data":   *data,
	})
	if err != nil {
		return []byte{}
	}

	return b
}
