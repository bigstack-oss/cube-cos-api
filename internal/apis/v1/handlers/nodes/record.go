package nodes

import (
	"time"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/mongo"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"
	ping "github.com/prometheus-community/pro-bing"
	log "go-micro.dev/v5/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func traceNodeStatus(hostname, operation string) {
	err := updatePendingStatus(hostname, operation)
	if err != nil {
		log.Errorf("nodes: failed to update pending status for node %s(%v)", hostname, err)
		return
	}

	checkNodeProgress(hostname, operation)
}

func checkNodeProgress(hostname, operation string) {
	p, err := ping.NewPinger(hostname)
	if err != nil {
		log.Errorf("nodes: failed to create pinger for %s(%v)", hostname, err)
		return
	}

	p.SetPrivileged(true)
	p.Count = 600
	p.Interval = time.Second * 1
	setTrackerCallback(p, hostname, operation)
	p.Run()
}

func setTrackerCallback(p *ping.Pinger, hostname, operation string) {
	if operation == "poweron" {
		p.OnRecv = func(pkt *ping.Packet) {
			log.Infof("nodes: node %s is starting", hostname)
			updateFinalStatus(hostname, operation)
			p.Stop()
		}
	}

	if operation == "poweroff" {
		p.OnRecvError = func(err error) {
			log.Errorf("nodes: node %s is not reachable, marking as down", hostname)
			updateFinalStatus(hostname, operation)
			p.Stop()
		}
	}
}

func updatePendingStatus(hostname, operation string) error {
	status := getPendingStatus(operation)
	m := mongo.GetGlobalHelper()
	return m.UpdateOne(
		nodes.Db,
		nodes.RequestsCollection,
		bson.M{"hostname": hostname},
		bson.M{"$set": bson.M{"hostname": hostname, "status": status}},
		options.Update().SetUpsert(true),
	)
}

func updateFinalStatus(hostname, operation string) {
	status := getFinalStatus(operation)
	m := mongo.GetGlobalHelper()
	err := m.UpdateOne(
		nodes.Db,
		nodes.RequestsCollection,
		bson.M{"hostname": hostname},
		bson.M{"$set": bson.M{"hostname": hostname, "status": status}},
		options.Update().SetUpsert(true),
	)
	if err != nil {
		log.Errorf("nodes: failed to update final status for %s(%v)", hostname, err)
	}
}

func getPendingStatus(operation string) string {
	switch operation {
	case "poweron":
		return status.PoweringOn
	case "poweroff":
		return status.PoweringOff
	case "powercycle":
		return status.PoweringCycle
	default:
		return "unknown inprogress status"
	}
}

func getFinalStatus(operation string) string {
	switch operation {
	case "poweron":
		return "up"
	case "poweroff":
		return "down"
	default:
		return "unknown final status"
	}
}

func (h *helper) genUpsertPayload() bson.M {
	return bson.M{
		"$set": bson.M{
			"host":     h.node,
			"ip":       h.ipmi.Ip,
			"port":     h.ipmi.Port,
			"username": h.ipmi.Username,
			"password": h.ipmi.Password,
		},
	}
}
