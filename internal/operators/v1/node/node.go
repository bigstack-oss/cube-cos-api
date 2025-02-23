package node

import (
	"context"

	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	log "go-micro.dev/v5/logger"
	"go-micro.dev/v5/registry"
)

var (
	module = "node"
)

func Name() string {
	return module
}

type Operator struct {
	ctx             context.Context
	cancel          context.CancelFunc
	isFirstTimeSync bool
}

func (o *Operator) Name() string {
	return module
}

func (o *Operator) Init() error {
	ctx, cancel := context.WithCancel(context.Background())
	o.ctx = ctx
	o.cancel = cancel
	o.isFirstTimeSync = true
	return nil
}

func (o *Operator) Sync() {
	watcher, err := registry.Watch(
		registry.WatchService(definition.DataCenterName),
	)
	if err != nil {
		log.Errorf("failed to create watcher (%s)", err.Error())
		return
	}

	defer watcher.Stop()
	select {
	case <-o.ctx.Done():
		return
	default:
		o.watchAndSyncNodeRoles(&watcher)
	}
}

func (o *Operator) Stop() {
	o.cancel()
}

func (o *Operator) watchAndSyncNodeRoles(watcher *registry.Watcher) {
	event, err := (*watcher).Next()
	if err == nil {
		definition.SyncNodesOfRole()
		logThrottling(event)
		return
	}

	log.Errorf(
		"failed to get service discovery event",
		err.Error(),
	)
}
