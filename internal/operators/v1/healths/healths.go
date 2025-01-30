package healths

import (
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/bigstack-oss/cube-cos-api/internal/service"
	"github.com/bigstack-oss/cube-cos-api/internal/wait"
	log "go-micro.dev/v5/logger"
	"k8s.io/client-go/util/workqueue"
)

var (
	ReqQueue workqueue.Interface
	module   = "health"
)

func init() {
	ReqQueue = workqueue.New()
	service.RegisterOperator(module, NewOperator())
}

func NewOperator() *Operator {
	return &Operator{}
}

type Operator struct{}

func (o *Operator) Name() string {
	return module
}

func (o *Operator) Sync() {
	defer definition.CapturePanic()

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

	o.reportToController(health)
	ReqQueue.Done(req)
}

func (o *Operator) Stop() {
	ReqQueue.ShutDown()
	o.waitForLastTask()
}

func (o *Operator) waitForLastTask() {
	for ReqQueue.Len() >= 1 {
		wait.Seconds(1)
	}
}
