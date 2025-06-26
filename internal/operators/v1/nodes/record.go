package node

import (
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/mongo"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/base"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	log "go-micro.dev/v5/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (o *Operator) removeHostPendingReq() {
	h := mongo.GetGlobalHelper()
	err := h.DeleteAll(
		nodes.Db,
		nodes.RequestsCollection,
		bson.M{"hostname": base.Hostname},
	)
	if err != nil {
		log.Errorf("nodes: failed to reset pending requests for host %s: %v", base.Hostname, err)
	}
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
