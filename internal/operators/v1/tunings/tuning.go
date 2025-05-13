package tunings

import (
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/tunings"
	"github.com/bigstack-oss/cube-cos-api/internal/service"
	"github.com/fsnotify/fsnotify"
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
	policy *fsnotify.Watcher
}

func NewOperator() *Operator {
	return &Operator{}
}

func (o *Operator) Name() string {
	return module
}

func (o *Operator) Init() error {
	err := o.initPolicyWatcher()
	if err != nil {
		return err
	}

	cubecos.SyncTunings()
	cubecos.RemovePendingReq(tunings.DB(), tunings.ReqCollection())
	return nil
}

func (o *Operator) Run() {
	for {
		req, shutdown := ReqQueue.Get()
		if shutdown {
			return
		}

		tuning := req.(*tunings.Tuning)
		err := o.operateReq(*tuning)
		o.handleExit(*tuning, err)

		ReqQueue.Done(req)
	}
}

func (o *Operator) Stop() {
	ReqQueue.ShutDown()
	o.waitForLastTask()
	o.policy.Close()
}

func (o *Operator) waitForLastTask() {
	for ReqQueue.Len() >= 1 {
		wait.Seconds(1)
	}
}
