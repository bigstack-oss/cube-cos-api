package node

import (
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/mongo"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/base"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	log "go-micro.dev/v5/logger"
	"go.mongodb.org/mongo-driver/bson"
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
