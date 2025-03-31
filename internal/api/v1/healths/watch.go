package healths

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

// M1 TODO:
// have to discuss with the FE to about should we auto shift the period in every round of watch?
func streamHealth() {
	for {
		wait.Seconds(2)
		if len(stream.Watchers) == 0 {
			continue
		}

		stream.Lock()
		for _, w := range stream.Watchers {
			health, err := streamHealthByHandlerType(&w.helper)
			if err != nil {
				continue
			}

			select {
			case w.dataChan <- health:
			default:
			}
		}

		stream.Unlock()
	}
}

func streamHealthByHandlerType(h *helper) (any, error) {
	switch h.handler {
	case "getHealthSummary":
		// return h.genFakeHealthSummary(), nil
		return h.getHealthSummary(), nil
	case "getHealthHistoryOfService":
		return h.genFakeHealthHistoryOfService(), nil
	case "getHealthHistoryOfModule":
		return h.genFakeHealthHistoryOfModule(), nil
	}

	return nil, errors.New("no internal function supported")
}

func watchHealth(h *helper, health any) {
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
			api.SetStatusOk(c, "health summary watching is stopped successfully", nil)
			return
		}
	}
}

func streamingResp(health any) []byte {
	b, err := json.Marshal(gin.H{
		"code":   http.StatusOK,
		"status": "ok",
		"msg":    "fetch health summary successfully",
		"data":   health,
	})
	if err != nil {
		return []byte{}
	}

	return b
}
