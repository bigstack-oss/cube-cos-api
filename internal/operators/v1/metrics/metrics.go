package supportfiles

import (
	"context"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/service"
)

var (
	module = "metrics"
)

func init() {
	service.RegisterOperator(module, NewOperator())
}

type Operator struct {
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
	o.ctx, o.cancel = context.WithCancel(context.Background())
	return nil
}

func (o *Operator) Sync() {
	for {
		select {
		case <-o.ctx.Done():
			return
		default:
			cubecos.SyncDataCenterMetricsSummary()
			wait.Seconds(60)
		}
	}
}

func (o *Operator) Stop() {
	if o.cancel != nil {
		o.cancel()
	}
}
