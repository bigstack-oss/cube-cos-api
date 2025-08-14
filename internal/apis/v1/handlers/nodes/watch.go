package nodes

import (
	"errors"
	"net/http"
	"slices"
	"sync"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/bodies"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/notifications"
	"github.com/gin-gonic/gin"
	json "github.com/json-iterator/go"
)

type dataChan chan any

type watcher struct {
	helper
	dataChan
	isWatchStopped bool
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
		bodies.SetBadRequest(h.c, errors.New("http chunked transfer is not supported"), nil)
		return
	}

	watcher := &watcher{helper: *h, dataChan: make(dataChan)}
	setWatcher(watcher)
	defer removeWatcher(watcher)

	sendFirstData(h.c, flusher, data)
	go runConnectionKeeper(h.c, flusher)
	streamingData(h.c, flusher, watcher)
}

func streamWatchers() {
	for {
		change, shutdown := changes.Get()
		changes.Done(change)
		if shutdown {
			return
		}

		sendNotification(change)
		waitWatchers()

		for _, w := range stream.Watchers {
			data, err := streamDataByHandler(&w.helper, change)
			if err != nil {
				continue
			}
			if w.isWatchStopped {
				continue
			}

			w.dataChan <- data
		}
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

func sendNotification(change nodes.Change) {
	if !change.NeedsNotification {
		return
	}

	payload, found := notifications.GetCacheById(change.Id)
	if !found {
		return
	}

	cubecos.InsertNotification(payload)
	notifications.DeleteCacheById(change.Id)
}

func streamDataByHandler(h *helper, change nodes.Change) (any, error) {
	switch h.handler {
	case "listNodes":
		return h.listNodes()
	case "getNode":
		return h.getNode()
	case "listNodeDevices":
		opts := genDeviceListOpts(change)
		return h.listNodeDevices(opts)
	}

	return nil, errors.New("no internal function supported")
}

func genDeviceListOpts(changes nodes.Change) nodes.DeviceListOpts {
	payload, found := notifications.GetCacheById(changes.Id)
	if !found {
		payload = notifications.Notification{}
	}

	return nodes.DeviceListOpts{
		UseCache: changes.IsTaskInprogress,
		Notify: nodes.Notify{
			Changes: changes.NeedsNotification,
			Payload: payload,
		},
	}
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

func sendFirstData(c *gin.Context, flusher http.Flusher, nodes any) {
	c.Writer.Write(streamingResp(nodes))
	c.Writer.Write([]byte("\n"))
	flusher.Flush()
}

func runConnectionKeeper(c *gin.Context, flusher http.Flusher) {
	for {
		wait.Seconds(10)
		select {
		case <-c.Request.Context().Done():
			return
		default:
			c.Writer.Write([]byte("\n"))
			flusher.Flush()
		}
	}
}

func streamingData(c *gin.Context, flusher http.Flusher, watcher *watcher) {
	ctx := c.Request.Context()
	for {
		select {
		case nodes := <-watcher.dataChan:
			c.Writer.Write(streamingResp(nodes))
			c.Writer.Write([]byte("\n"))
			flusher.Flush()
		case <-ctx.Done():
			bodies.SetOk(c, "nodes watching is stopped successfully", nil)
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
