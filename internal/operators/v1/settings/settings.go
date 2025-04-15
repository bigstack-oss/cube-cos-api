package settings

import (
	"errors"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/mongo"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/setting"
	"github.com/bigstack-oss/cube-cos-api/internal/service"
	"github.com/fsnotify/fsnotify"
	"k8s.io/client-go/util/workqueue"
)

var (
	ReqQueue workqueue.Interface
	module   = "settings"
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

	return nil
}

func (o *Operator) Sync() {
	req, shutdown := ReqQueue.Get()
	if shutdown {
		return
	}

	setting := req.(*setting.Options)
	err := o.operateReq(*setting)
	o.handleExit(*setting, err)

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
