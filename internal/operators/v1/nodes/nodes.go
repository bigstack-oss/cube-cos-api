package node

import (
	"context"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/base"
	bslog "github.com/bigstack-oss/cube-cos-api/internal/log"
	"github.com/bigstack-oss/cube-cos-api/internal/service"
	log "go-micro.dev/v5/logger"
	"go-micro.dev/v5/registry"
)

var (
	module = "node"
)

func init() {
	service.RegisterOperator(module, &Operator{})
}

type Operator struct {
	ctx    context.Context
	cancel context.CancelFunc
}

func (o *Operator) Name() string {
	return module
}

func (o *Operator) Init() error {
	o.ctx, o.cancel = context.WithCancel(context.Background())
	go o.syncOrderSensitiveServices()
	go o.periodicSyncNodes()
	o.removeHostPendingReq()
	return nil
}

func (o *Operator) Run() {
	for {
		watcher, err := registry.Watch(
			registry.WatchService(base.ServiceDiscoveryIdentity),
		)
		if err != nil {
			log.Errorf("nodes: failed to create watcher(%v)", err)
			return
		}

		select {
		case <-o.ctx.Done():
			return
		default:
			o.checkAndSyncNodes(&watcher)
		}
	}
}

func (o *Operator) Stop() {
	o.cancel()
}

func (o *Operator) checkAndSyncNodes(watcher *registry.Watcher) {
	defer (*watcher).Stop()
	event, err := (*watcher).Next()
	if err != nil {
		log.Errorf("nodes: failed to get service discovery event(%v)", err)
		return
	}

	o.syncNodes()
	bslog.Throttle("node", genDiscoveryMsg(event))
}
