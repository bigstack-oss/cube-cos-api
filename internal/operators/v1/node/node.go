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

func NewOperator() *Operator {
	ctx, cancel := context.WithCancel(context.Background())
	return &Operator{
		ctx:             ctx,
		cancel:          cancel,
		isFirstTimeSync: true,
	}
}

func (o *Operator) Name() string {
	return module
}

func (o *Operator) Sync() {
	watcher, err := registry.Watch()
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
		logWithThrottling(event)
		return
	}

	log.Errorf(
		"failed to get service discovery event",
		err.Error(),
	)
}
