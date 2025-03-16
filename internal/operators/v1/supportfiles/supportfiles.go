package supportfiles

import (
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/bigstack-oss/cube-cos-api/internal/service"
	"github.com/fsnotify/fsnotify"
	"k8s.io/client-go/util/workqueue"
)

var (
	ReqQueue workqueue.Interface
	module   = "supportfiles"
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
	return o.initWatcher()
}

func (o *Operator) Sync() {
	req, shutdown := ReqQueue.Get()
	if shutdown {
		return
	}

	supportFile := req.(*v1.SupportFile)
	err := o.operateReq(*supportFile)
	o.handleExit(*supportFile, err)

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
