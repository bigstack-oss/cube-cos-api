package settings

import (
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/settings"
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
	watcher *fsnotify.Watcher
}

func NewOperator() *Operator {
	return &Operator{}
}

func (o *Operator) Name() string {
	return module
}

func (o *Operator) Init() error {
	cubecos.SyncAlertSettings()
	cubecos.RemovePendingReq(settings.DB, settings.ReqCollection)
	return o.initWatcher()
}

func (o *Operator) Run() {
	for {
		req, shutdown := ReqQueue.Get()
		if shutdown {
			return
		}

		setting := req.(*settings.Setting)
		err := o.operateReq(*setting)
		o.handleExit(*setting, err)

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
}
