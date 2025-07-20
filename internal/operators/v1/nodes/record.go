package node

import (
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/mongo"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/base"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	log "go-micro.dev/v5/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (o *Operator) removeHostPendingReqs() {
	h := mongo.GetGlobalHelper()
	h.DeleteAll(nodes.Db, nodes.ReqDeviceCollection, bson.M{"hostname": base.Hostname})
	h.DeleteAll(nodes.Db, nodes.ReqOsdCollection, bson.M{"hostname": base.Hostname})

	wait.Seconds(90)
	h.DeleteAll(nodes.Db, nodes.ReqCollection, bson.M{"hostname": base.Hostname})
}

func (o *Operator) setIpmiEnablement(node *nodes.Node) {
	h := mongo.GetGlobalHelper()
	err := h.UpdateOne(
		nodes.Db,
		nodes.CollectionIpmiSupport,
		bson.M{"host": base.Hostname},
		bson.M{
			"$set": bson.M{
				"host":      base.Hostname,
				"supported": node.IpmiEnablement.IsSupported,
			},
		},
		options.Update().SetUpsert(true),
	)
	if err != nil {
		log.Errorf("nodes: failed to set IPMI enablement for host %s: %v", base.Hostname, err)
	}
}
