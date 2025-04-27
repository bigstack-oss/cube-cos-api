package node

import (
	"context"
	"sync"

	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	cubelog "github.com/bigstack-oss/cube-cos-api/internal/log"
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
	ctx             context.Context
	cancel          context.CancelFunc
	isFirstTimeSync bool
	sync            sync.Mutex
}

func (o *Operator) Name() string {
	return module
}

func (o *Operator) Init() error {
	ctx, cancel := context.WithCancel(context.Background())
	o.ctx = ctx
	o.cancel = cancel
	o.isFirstTimeSync = true
	o.sync = sync.Mutex{}
	go o.traceNodeDetails()
	return nil
}

func (o *Operator) Run() {
	for {
		watcher, err := registry.Watch(
			registry.WatchService(v1.ServiceDiscoveryIdentity),
		)
		if err != nil {
			log.Errorf("nodes: failed to create watcher (%s)", err.Error())
			return
		}

		select {
		case <-o.ctx.Done():
			return
		default:
			o.watchAndSyncNodes(&watcher)
		}
	}
}

func (o *Operator) Stop() {
	o.cancel()
}

func (o *Operator) watchAndSyncNodes(watcher *registry.Watcher) {
	defer (*watcher).Stop()
	event, err := (*watcher).Next()
	if err != nil {
		log.Errorf("nodes: failed to get service discovery event", err.Error())
		return
	}

	o.syncNodeDetails()
	cubelog.Throttle("node", genDiscoveryMsg(event))
}
