package nodes

import (
	"errors"
	"net/http"
	"sync"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	"github.com/bigstack-oss/cube-cos-api/internal/api"
	"github.com/gin-gonic/gin"
	json "github.com/json-iterator/go"
)

type respChan chan data

type watcher struct {
	helper
	respChan
}

var (
	stream = struct {
		sync.Mutex
		Watchers []watcher
	}{}
)

func streamNodes() {
	for {
		wait.Seconds(2)
		if len(stream.Watchers) == 0 {
			continue
		}

		stream.Lock()
		for _, w := range stream.Watchers {
			nodes, err := w.genNodeResp()
			if err != nil {
				continue
			}

			select {
			case w.respChan <- *nodes:
			default:
			}
		}

		stream.Unlock()
	}
}

func watchNodes(h *helper, nodes data) {
	setChunkedTransfer(h.c)
	flusher, ok := h.c.Writer.(http.Flusher)
	if !ok {
		api.SetBadRequest(h.c, errors.New("http chunked transfer is not supported"))
		return
	}

	watcher := watcher{helper: *h, respChan: make(respChan)}
	addWatcher(watcher)
	defer removeWatcher(watcher)

	sendFirstSummary(h.c, flusher, nodes)
	streamingSummary(h.c, flusher, watcher)
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

func sendFirstSummary(c *gin.Context, flusher http.Flusher, nodes data) {
	c.Writer.Write(streamingResp(nodes))
	c.Writer.Write([]byte("\n"))
	flusher.Flush()
}

func streamingSummary(c *gin.Context, flusher http.Flusher, watcher watcher) {
	ctx := c.Request.Context()
	for {
		select {
		case nodes := <-watcher.respChan:
			c.Writer.Write(streamingResp(nodes))
			c.Writer.Write([]byte("\n"))
			flusher.Flush()
		case <-ctx.Done():
			api.SetStatusOk(c, "nodes watching is stopped successfully", nil)
			return
		}
	}
}

func streamingResp(nodes data) []byte {
	b, err := json.Marshal(gin.H{
		"code":   http.StatusOK,
		"status": "ok",
		"msg":    "fetch nodes successfully",
		"data":   nodes,
	})
	if err != nil {
		return []byte{}
	}

	return b
}
