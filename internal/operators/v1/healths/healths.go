package healths

import (
	"context"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/bigstack-oss/cube-cos-api/internal/service"
	log "go-micro.dev/v5/logger"
	"k8s.io/client-go/util/workqueue"
)

var (
	ReqQueue workqueue.Interface
	module   = "health"
)

func init() {
	ReqQueue = workqueue.New()
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
	cubecos.RemovePendingReq(v1.Healths, v1.HealthRepairingCollection)
	go o.initHealthHistoryResync()
	return nil
}

func (o *Operator) Run() {
	defer v1.CapturePanic()

	for {
		req, shutdown := ReqQueue.Get()
		if shutdown {
			return
		}

		health := req.(*cubecos.Health)
		err := o.operateReq(*health)
		if err != nil {
			log.Errorf("health: failed to operate request: %s", err.Error())
			health.Overall.Status.SetCurrentToError(err)
		}

		ReqQueue.Done(req)
	}
}

func (o *Operator) Stop() {
	ReqQueue.ShutDown()
	o.waitForLastTask()
}

func (o *Operator) waitForLastTask() {
	for ReqQueue.Len() >= 1 {
		wait.Seconds(1)
	}

	if o.cancel != nil {
		o.cancel()
	}
}
