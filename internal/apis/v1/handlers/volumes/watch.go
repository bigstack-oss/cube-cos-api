package volumes

import (
	"errors"
	"net/http"
	"slices"
	"sync"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/bodies"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/volumes"
	"github.com/gin-gonic/gin"
	json "github.com/json-iterator/go"
	"k8s.io/client-go/util/workqueue"
)

var (
	changes = workqueue.NewTyped[volumes.Change]()
)

type dataChan chan any

type watcher struct {
	helper
	isWatchStopped bool
	dataChan
}

var (
	stream = struct {
		sync.Mutex
		Watchers []*watcher
	}{}
)

func streamData(h *helper, data any) {
	setChunkedTransfer(h.c)
	flusher, ok := h.c.Writer.(http.Flusher)
	if !ok {
		bodies.SetBadRequest(h.c, errors.New("http chunked transfer is not supported"))
		return
	}

	watcher := &watcher{helper: *h, dataChan: make(dataChan)}
	setWatcher(watcher)
	defer removeWatcher(watcher)

	sendFirstData(h.c, flusher, data)
	go periodicUpdateData(h.c)
	streamingData(h.c, flusher, watcher)
}

func streamWatchers() {
	for {
		change, shutdown := changes.Get()
		changes.Done(change)
		if shutdown {
			return
		}

		waitWatchers()
		syncWatchers()
	}
}

func syncWatchers() {
	stream.Lock()
	defer stream.Unlock()

	for _, w := range stream.Watchers {
		data, err := streamDataByHandler(&w.helper)
		if err != nil {
			continue
		}
		if w.isWatchStopped {
			continue
		}

		w.dataChan <- data
	}
}

func waitWatchers() {
	for range 60 {
		if len(stream.Watchers) != 0 {
			return
		}

		wait.Seconds(1)
	}
}

func streamDataByHandler(h *helper) (any, error) {
	switch h.handler {
	case "listVolumes":
		return h.listVolumes()
	}

	return nil, errors.New("no internal function supported")
}

func setChunkedTransfer(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")
}

func setWatcher(w *watcher) {
	stream.Lock()
	defer stream.Unlock()
	stream.Watchers = append(stream.Watchers, w)
}

func removeWatcher(watcherToRemove *watcher) {
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

func sendFirstData(c *gin.Context, flusher http.Flusher, volumes any) {
	c.Writer.Write(streamingResp(volumes))
	c.Writer.Write([]byte("\n"))
	flusher.Flush()
}

func periodicUpdateData(c *gin.Context) {
	for {
		wait.Seconds(5)
		select {
		case <-c.Request.Context().Done():
			return
		default:
			changes.Add(volumes.Change{})
		}
	}
}

func streamingData(c *gin.Context, flusher http.Flusher, watcher *watcher) {
	ctx := c.Request.Context()
	for {
		select {
		case volumes := <-watcher.dataChan:
			c.Writer.Write(streamingResp(volumes))
			c.Writer.Write([]byte("\n"))
			flusher.Flush()
		case <-ctx.Done():
			bodies.SetOk(c, "volumes watching is stopped successfully", nil)
			watcher.isWatchStopped = true
			close(watcher.dataChan)
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
