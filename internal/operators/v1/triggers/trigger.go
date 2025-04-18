package triggers

import (
	"errors"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/mongo"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/trigger"
	"github.com/bigstack-oss/cube-cos-api/internal/service"
	"github.com/fsnotify/fsnotify"
	"k8s.io/client-go/util/workqueue"
)

var (
	ReqQueue workqueue.Interface
	module   = "triggers"
)

func init() {
	ReqQueue = workqueue.New()
	service.RegisterOperator(module, NewOperator())
}

type Operator struct {
	mongo  *mongo.Helper
	policy *fsnotify.Watcher
}

func NewOperator() *Operator {
	return &Operator{}
}

func (o *Operator) Name() string {
	return module
}

func (o *Operator) Init() error {
	o.mongo = mongo.GetGlobalHelper()
	if o.mongo == nil {
		return errors.New("mongo helper is not initialized")
	}

	err := o.initPolicyWatcher()
	if err != nil {
		return err
	}

	return nil
}

func (o *Operator) Run() {
	req, shutdown := ReqQueue.Get()
	if shutdown {
		return
	}

	trigger := req.(*trigger.Options)
	err := o.operateReq(*trigger)
	o.handleExit(*trigger, err)

	ReqQueue.Done(req)
}

func (o *Operator) Stop() {
	ReqQueue.ShutDown()
	o.waitForLastTask()
	o.mongo.Close()
	o.policy.Close()
}

func (o *Operator) waitForLastTask() {
	for ReqQueue.Len() >= 1 {
		wait.Seconds(1)
	}
}
