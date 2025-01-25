package tunings

import (
	"time"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/mongo"
	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/bigstack-oss/cube-cos-api/internal/service"
	"k8s.io/client-go/util/workqueue"
)

var (
	ReqQueue workqueue.Interface
	module   = "tunings"
)

func init() {
	ReqQueue = workqueue.New()
	service.RegisterOperator(module, NewOperator())
}

type Operator struct {
	mongo *mongo.Helper
}

func NewOperator() *Operator {
	return &Operator{mongo: mongo.GetGlobalHelper()}
}

func (o *Operator) Name() string {
	return module
}

func (o *Operator) Sync() {
	req, shutdown := ReqQueue.Get()
	if shutdown {
		return
	}

	tuning := req.(definition.Tuning)
	err := o.operateReq(tuning)

	o.handleExit(tuning, err)
	ReqQueue.Done(req)
}

func (o *Operator) Stop() {
	ReqQueue.ShutDown()
	o.waitForLastTask()
	o.mongo.Close()
}

func (o *Operator) waitForLastTask() {
	for ReqQueue.Len() >= 1 {
		time.Sleep(time.Second * 1)
	}
}
