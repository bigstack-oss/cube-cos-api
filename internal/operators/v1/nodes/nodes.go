package node

import (
	"context"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/base"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	bslog "github.com/bigstack-oss/cube-cos-api/internal/log"
	"github.com/bigstack-oss/cube-cos-api/internal/service"
	log "go-micro.dev/v5/logger"
	"go-micro.dev/v5/registry"
	"k8s.io/client-go/util/workqueue"
)

var (
	module         = "node"
	OsdReqQueue    workqueue.TypedInterface[nodes.OsdReqOpts]
	DeviceReqQueue workqueue.TypedInterface[nodes.DeviceReqOpts]
)

func init() {
	OsdReqQueue = workqueue.NewTyped[nodes.OsdReqOpts]()
	DeviceReqQueue = workqueue.NewTyped[nodes.DeviceReqOpts]()
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
	go o.removeHostPendingReq()
	return nil
}

func (o *Operator) Run() {
	go o.continuesSyncNodes()
	go o.handleOsdReqs()
	go o.handleDeviceReqs()
}

func (o *Operator) continuesSyncNodes() {
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

func (o *Operator) handleOsdReqs() {
	for {
		select {
		case <-o.ctx.Done():
			return
		default:
			req, shutdown := OsdReqQueue.Get()
			if shutdown {
				return
			}

			err := o.operateOsd(req)
			o.handleOsdExit(req, err)
			OsdReqQueue.Done(req)
		}
	}
}

func (o *Operator) handleDeviceReqs() {
	for {
		select {
		case <-o.ctx.Done():
			return
		default:
			req, shutdown := DeviceReqQueue.Get()
			if shutdown {
				return
			}

			err := o.operateDevice(req)
			o.handleDeviceExit(req, err)
			DeviceReqQueue.Done(req)
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
