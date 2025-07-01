package licenses

import (
	"encoding/json"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/mongo"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	log "go-micro.dev/v5/logger"
	"go.mongodb.org/mongo-driver/bson"
)

func (h *helper) getTemprorayNodeDetails(hostname string) *nodes.Node {
	mongo := mongo.GetGlobalHelper()
	doc, err := mongo.Get(
		nodes.Db,
		nodes.CollectionTemporaryNodeDetails,
		bson.M{"hostname": hostname},
	)
	if err != nil {
		log.Errorf("licenses(%s): failed to get temporary node details for %s(%v)", h.reqId, hostname, err)
		return nil
	}

	node := &nodes.Node{}
	err = doc.Decode(node)
	if err != nil {
		log.Errorf("licenses(%s): failed to decode temporary node details for %s(%v)", h.reqId, hostname, err)
		return nil
	}

	log.Infof("----------------------------------")
	b, _ := json.MarshalIndent(node, "", "  ")
	log.Infof(string(b))

	return node
}
