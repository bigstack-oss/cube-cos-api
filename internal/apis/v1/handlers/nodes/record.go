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
	changes.Add(nodes.Change{IsTaskInprogress: true})
	err := h.mongo.UpdateOne(
		nodes.Db,
		nodes.ReqDeviceCollection,
		bson.M{"hostname": h.node, "reqId": h.deviceReqOpts.ReqId},
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

func (h *helper) upsertOsdReqRecord() {
	changes.Add(nodes.Change{IsTaskInprogress: true})
	err := h.mongo.UpdateOne(
		nodes.Db,
		nodes.ReqOsdCollection,
		bson.M{"hostname": h.node, "reqId": h.osdReqOpts.ReqId},
		bson.M{"$set": h.osdReqOpts},
		options.Update().SetUpsert(true),
	)
	if err != nil {
		log.Errorf(
			"nodes(%s): failed to add osd request record for %s(%v)",
			h.reqId,
			h.node,
			err,
		)
	}
}

func (h *helper) syncUpdatingBlockDevices(blockDevs *[]nodes.BlockDevice) {
	for i, dev := range *blockDevs {
		if h.hasUpdatingDeviceReq(dev) {
			h.syncUpdatingDevice(&(*blockDevs)[i])
		}

		if h.hasUpdatingOsdReq(dev) {
			h.syncUpdatingOsd(&(*blockDevs)[i])
		}
	}
}

func (h *helper) syncCachedBlockDevices(blockDevs []nodes.BlockDevice) {
	lastDeviceList.Store(h.node, blockDevs)
}

func (h *helper) hasUpdatingDeviceReq(dev nodes.BlockDevice) bool {
	count, err := h.mongo.GetCount(
		nodes.Db,
		nodes.ReqDeviceCollection,
		bson.M{
			"hostname": h.node,
			"device":   fmt.Sprintf("/dev/%s", dev.Name),
		},
	)
	if err != nil {
		log.Errorf("nodes(%s): failed to get updating device req for %s(%v)", h.reqId, dev.Name, err)
		return false
	}

	return count > 0
}

func (h *helper) hasUpdatingOsdReq(dev nodes.BlockDevice) bool {
	count, err := h.mongo.GetCount(
		nodes.Db,
		nodes.ReqOsdCollection,
		bson.M{
			"hostname": h.node,
			"device":   fmt.Sprintf("/dev/%s", dev.Name),
		},
	)
	if err != nil {
		log.Errorf("nodes(%s): failed to get updating osd req for %s(%v)", h.reqId, dev.Name, err)
		return false
	}

	return count > 0
}

func (h *helper) syncUpdatingDevice(dev *nodes.BlockDevice) {
	dev.Status.Current = status.Processing
	dev.Status.IsProcessing = true

	update, err := h.getUpdatingDevice(dev)
	if err != nil {
		log.Errorf("nodes(%s): failed to get updating device req for %s(%v)", h.reqId, dev.Name, err)
		return
	}
	if update.Class != "" {
		dev.Class = update.Class
	}
}

func (h *helper) syncUpdatingOsd(dev *nodes.BlockDevice) {
	dev.Status.Current = status.Processing
	dev.Status.IsProcessing = true

	updates, err := h.getUpdatingOsds(dev)
	if err != nil {
		return
	}

	updatingMaxReweight := 0.0
	for _, update := range updates {
		if update.Reweight > updatingMaxReweight {
			updatingMaxReweight = update.Reweight
		}
	}

	dev.Osd.Reweigth = updatingMaxReweight
}

func (h *helper) getUpdatingDevice(dev *nodes.BlockDevice) (*nodes.BlockDevice, error) {
	doc, err := h.mongo.Get(
		nodes.Db,
		nodes.ReqDeviceCollection,
		bson.M{
			"hostname": h.node,
			"device":   fmt.Sprintf("/dev/%s", dev.Name),
		},
	)
	if err != nil {
		log.Errorf("nodes(%s): failed to get updating device req for %s(%v)", h.reqId, dev.Name, err)
		return nil, err
	}
	if doc == nil {
		log.Warnf("nodes(%s): no updating device req found for %s", h.reqId, dev.Name)
		return nil, err
	}

	device := &nodes.BlockDevice{}
	err = doc.Decode(device)
	if err != nil {
		log.Errorf("nodes(%s): failed to decode updating device req for %s(%v)", h.reqId, dev.Name, err)
		return nil, err
	}

	return device, nil
}

func (h *helper) getUpdatingOsds(dev *nodes.BlockDevice) ([]nodes.OsdReqOpts, error) {
	updates := []nodes.OsdReqOpts{}
	for _, osd := range dev.Osd.Daemons {
		update, err := h.getUpdatingOsd(dev.Name, osd.Id)
		if err == nil {
			updates = append(updates, *update)
		}
	}

	if len(updates) == 0 {
		err := fmt.Errorf("no updating osd req found for %s", dev.Name)
		log.Warnf("nodes(%s): %v", h.reqId, err)
		return updates, err
	}

	return updates, nil
}

func (h *helper) getUpdatingOsd(device, id string) (*nodes.OsdReqOpts, error) {
	doc, err := h.mongo.Get(
		nodes.Db,
		nodes.ReqOsdCollection,
		bson.M{
			"hostname": h.node,
			"device":   fmt.Sprintf("/dev/%s", device),
			"osdId":    id,
		},
	)
	if err != nil {
		log.Errorf("nodes(%s): failed to get updating osd req for %s(%v)", h.reqId, device, err)
		return nil, err
	}
	if doc == nil {
		log.Warnf("nodes(%s): no updating osd req found for %s", h.reqId, device)
		return nil, err
	}

	osd := &nodes.OsdReqOpts{}
	err = doc.Decode(osd)
	if err != nil {
		log.Errorf("nodes(%s): failed to decode updating osd req for %s(%v)", h.reqId, device, err)
		return nil, err
	}

	return osd, nil
}
