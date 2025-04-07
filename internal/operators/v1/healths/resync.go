package healths

import (
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/influx"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
)

func (o *Operator) initHealthHistoryResync() {
	for {
		if influx.GetGlobalHelper() == nil {
			wait.Seconds(1)
			continue
		}

		select {
		case <-o.ctx.Done():
			return
		default:
			cubecos.SyncHealthHistory()
			wait.Seconds(60)
		}
	}
}
