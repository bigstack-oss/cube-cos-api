package images

import (
	"context"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/images"
	"github.com/bigstack-oss/cube-cos-api/internal/service"
	"k8s.io/client-go/util/workqueue"
)

var (
	module   = "image"
	ReqQueue workqueue.TypedInterface[*images.ReqOpts]
)

func init() {
	ReqQueue = workqueue.NewTyped[*images.ReqOpts]()
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
	go o.removePendingReqs()
	return nil
}

func (o *Operator) Run() {
	for {
		select {
		case <-o.ctx.Done():
			return
		default:
			req, shutdown := ReqQueue.Get()
			if shutdown {
				return
			}

			err := o.operate(*req)
			o.handleExit(*req, err)
			ReqQueue.Done(req)
		}
	}
}

func (o *Operator) Stop() {
	o.cancel()
}
