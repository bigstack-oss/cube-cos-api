package nodes

import (
	"fmt"
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
			updatePendingStatus(hostname, operation)
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
		nodes.ReqCollection,
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
		nodes.ReqCollection,
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
		return status.Up
	case "poweroff":
		return status.Down
	default:
		return status.Syncing
	}
}

func (h *helper) genIpmiUpsertPayload() bson.M {
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

func (h *helper) syncTemporaryNodeDetails() error {
	if h.operation != "poweroff" {
		return nil
	}

	err := h.saveTemporaryNodeDetails()
	if err != nil {
		log.Errorf("nodes(%s): failed to save temporary node details(%v)", h.reqId, err)
		return fmt.Errorf(
			"failed to save node info due to the db issue",
		)
	}

	return nil
}

func (h *helper) saveTemporaryNodeDetails() error {
	node, err := nodes.Get(h.node)
	if err != nil {
		log.Errorf("nodes(%s): failed to get node details(%v)", h.reqId, err)
		return err
	}

	return h.mongo.UpdateOne(
		nodes.Db,
		nodes.CollectionTemporaryNodeDetails,
		bson.M{"hostname": h.node},
		bson.M{"$set": node},
		options.Update().SetUpsert(true),
	)
}

func (h *helper) upsertDeviceReqRecord() {
	err := h.mongo.UpdateOne(
		nodes.Db,
		nodes.ReqDeviceCollection,
		bson.M{"hostname": h.node},
		bson.M{"$set": h.deviceReqOpts},
		options.Update().SetUpsert(true),
	)
	if err != nil {
		log.Errorf(
			"nodes(%s): failed to add device request record for %s(%v)",
			h.reqId,
			h.node,
			err,
		)
	}
}

func (h *helper) syncUpdatingBlockDevices(blockDevs *[]nodes.BlockDevice) {
	for i, dev := range *blockDevs {
		if h.hasUpdatingReq(dev) {
			(*blockDevs)[i].Status.IsProcessing = true
			continue
		}
	}
}

func (h *helper) syncCachedBlockDevices(blockDevs []nodes.BlockDevice) {
	lastDeviceList.Store(h.node, blockDevs)
}

func (h *helper) hasUpdatingReq(dev nodes.BlockDevice) bool {
	count, err := h.mongo.GetCount(
		nodes.Db,
		nodes.ReqDeviceCollection,
		bson.M{"hostname": h.node, "device": fmt.Sprintf("/dev/%s", dev.Name)},
	)
	if err != nil {
		log.Errorf("nodes(%s): failed to get updating record for device %s(%v)", h.reqId, dev.Name, err)
		return false
	}

	return count > 0
}
