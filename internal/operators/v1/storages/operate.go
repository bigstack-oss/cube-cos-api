package storages

import (
	"context"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/http"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/storages"
	"github.com/bigstack-oss/cube-cos-api/internal/service"
	"k8s.io/client-go/util/workqueue"
)

var (
	ReqQueue      workqueue.TypedInterface[*storages.ReqOpts]
	ModelReqQueue workqueue.TypedInterface[*storages.ModelReqOpts]
	module        = "storages"
)

func init() {
	ReqQueue = workqueue.NewTyped[*storages.ReqOpts]()
	ModelReqQueue = workqueue.NewTyped[*storages.ModelReqOpts]()
	service.RegisterOperator(module, NewOperator())
}

type Operator struct {
	http   *http.Helper
	ctx    context.Context
	cancel context.CancelFunc
}

func NewOperator() *Operator {
	return &Operator{}
}

func (o *Operator) Name() string {
	return module
}

func (o *Operator) Init() error {
	o.http = http.GetGlobalHelper()
	o.ctx, o.cancel = context.WithCancel(context.Background())
	cubecos.RemoveHostPendingReq(storages.Db, storages.ReqCollection)
	cubecos.RemoveHostPendingReq(storages.Db, storages.ModelReqCollection)
	return nil
}

func (o *Operator) Run() {
	go o.handleStorageReqs()
	go o.handleModelReqs()
}

func (o *Operator) handleStorageReqs() {
	for {
		select {
		case <-o.ctx.Done():
			return
		default:
			req, shutdown := ReqQueue.Get()
			if shutdown {
				return
			}

			ReqQueue.Done(req)
			if req == nil {
				continue
			}

			err := o.operateStorageReq(*req)
			o.handleStorageExit(*req, err)
		}
	}
}

func (o *Operator) handleModelReqs() {
	for {
		select {
		case <-o.ctx.Done():
			return
		default:
			req, shutdown := ModelReqQueue.Get()
			if shutdown {
				return
			}

			ModelReqQueue.Done(req)
			if req == nil {
				continue
			}

			err := o.operateModelReq(*req)
			o.handleModelExit(*req, err)
		}
	}
}

func (o *Operator) Stop() {
	o.cancel()
}
