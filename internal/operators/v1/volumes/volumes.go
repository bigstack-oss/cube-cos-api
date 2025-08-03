package volumes

import (
	"context"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/http"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/images"
	"github.com/bigstack-oss/cube-cos-api/internal/service"
	"k8s.io/client-go/util/workqueue"
)

var (
	module                  = "volume"
	ImageConvertionReqQueue workqueue.TypedInterface[*images.ReqOpts]
)

func init() {
	ImageConvertionReqQueue = workqueue.NewTyped[*images.ReqOpts]()
	service.RegisterOperator(module, &Operator{})
}

type Operator struct {
	ctx    context.Context
	cancel context.CancelFunc
	http   *http.Helper
}

func (o *Operator) Name() string {
	return module
}

func (o *Operator) Init() error {
	o.ctx, o.cancel = context.WithCancel(context.Background())
	o.http = http.GetGlobalHelper()
	go o.removePendingReqs()
	return nil
}

func (o *Operator) Run() {
	for {
		select {
		case <-o.ctx.Done():
			return
		default:
			req, shutdown := ImageConvertionReqQueue.Get()
			ImageConvertionReqQueue.Done(req)
			if shutdown {
				return
			}

			err := o.operateImageConvertion(req)
			o.handleExit(req, err)
		}
	}
}

func (o *Operator) Stop() {
	o.cancel()
}
