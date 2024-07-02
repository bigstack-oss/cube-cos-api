package node

import (
	"context"
	"fmt"
	"time"

	"github.com/bigstack-oss/cube-cos-api/internal/status"
	"github.com/cnf/structhash"
	"go-micro.dev/v5/cache"
	log "go-micro.dev/v5/logger"
	"go-micro.dev/v5/registry"
)

var (
	logCache = cache.NewCache(cache.Expiration(time.Second * 10))
)

func convertAction(action string) string {
	switch action {
	case status.Create:
		return "joined"
	case status.Delete:
		return "left"
	}

	return action
}

func logWithThrottling(event *registry.Result) {
	key, err := structhash.Hash(event, 1)
	if err != nil {
		return
	}

	getCtx, getCancel := context.WithTimeout(context.Background(), time.Second*10)
	defer getCancel()
	_, _, err = logCache.Get(getCtx, key)
	if err == nil {
		return
	}
	if len(event.Service.Nodes) == 0 {
		return
	}

	putCtx, putCancel := context.WithTimeout(context.Background(), time.Second*10)
	defer putCancel()
	err = logCache.Put(putCtx, key, []byte{}, time.Second*10)
	if err != nil {
		return
	}

	log.Infof(
		"Node resynced: %s %s %s",
		event.Service.Name,
		fmt.Sprintf("%s(%s)", event.Service.Metadata["hostname"], event.Service.Nodes[0].Address),
		convertAction(event.Action),
	)
}
