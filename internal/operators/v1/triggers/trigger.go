package triggers

import (
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/http"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/triggers"
	"github.com/bigstack-oss/cube-cos-api/internal/service"
	"github.com/fsnotify/fsnotify"
	"k8s.io/client-go/util/workqueue"
)

var (
	ReqQueue workqueue.TypedInterface[*triggers.ReqOpts]
	module   = "triggers"
)

func init() {
	ReqQueue = workqueue.NewTyped[*triggers.ReqOpts]()
	service.RegisterOperator(module, NewOperator())
}

type Operator struct {
	http   *http.Helper
	policy *fsnotify.Watcher
}

func NewOperator() *Operator {
	return &Operator{}
}

func (o *Operator) Name() string {
	return module
}

func (o *Operator) Init() error {
	o.http = http.GetGlobalHelper()
	err := o.initPolicyWatcher()
	if err != nil {
		return err
	}

	go o.syncTriggers()
	cubecos.ForceRemovePendingReq(triggers.DB, triggers.ReqCollection)
	return nil
}

func (o *Operator) Run() {
	for {
		req, shutdown := ReqQueue.Get()
		if shutdown {
			return
		}

		ReqQueue.Done(req)
		if req == nil {
			continue
		}

		err := o.operateReq(*req)
		o.handleExit(*req, err)
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
