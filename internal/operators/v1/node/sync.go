package node

import (
	"time"

	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
)

func (o *Operator) syncNodeDetails() {
	for {
		select {
		case <-o.ctx.Done():
			return
		default:
			nodes := definition.ListNodes()
			o.addLicenseInfoToNodes(&nodes)
			o.addDetailsToNodes(&nodes)
			definition.SetNodeDetails(nodes)
			time.Sleep(time.Second * 30)
		}
	}
}
