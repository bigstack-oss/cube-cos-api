package fixpacks

import (
	"context"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/http"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/fixpacks"
	"github.com/bigstack-oss/cube-cos-api/internal/service"
	"k8s.io/client-go/util/workqueue"
)

var (
	module   = "fixpacks"
	ReqQueue workqueue.TypedInterface[*fixpacks.ReqOpts]
)

func init() {
	ReqQueue = workqueue.NewTyped[*fixpacks.ReqOpts]()
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
	cubecos.RemoveHostPendingReq(fixpacks.Db, fixpacks.ReqCollection)
	cubecos.RemoveFixpackRebootingMarker()
	return nil
}

func (o *Operator) Run() {
	for {
		select {
		case <-o.ctx.Done():
			return
		default:
			req, shutdown := ReqQueue.Get()
			ReqQueue.Done(req)
			if shutdown {
				return
			}

			err := o.operate(req)
			o.handleExit(req, err)
		}
	}
}

func (o *Operator) Stop() {
	o.cancel()
}
