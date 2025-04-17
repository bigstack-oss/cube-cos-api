package node

import (
	"fmt"
	"time"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/http"
	"github.com/bigstack-oss/cube-cos-api/internal/api"
	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	log "go-micro.dev/v5/logger"
)

func (o *Operator) traceNodeDetails() {
	for {
		select {
		case <-o.ctx.Done():
			return
		default:
			o.syncNodeDetails()
			time.Sleep(time.Second * 30)
		}
	}
}

func (o *Operator) syncNodeDetails() {
	o.sync.Lock()
	defer o.sync.Unlock()

	definition.SyncRoleNodes()
	nodes := definition.GetNodesFromRoles()
	o.setNodeDetails(&nodes)
	definition.SetNodeDetails(nodes)
}

func (o *Operator) setNodeDetails(nodes *[]definition.Node) {
	if len(*nodes) == 0 {
		return
	}

	for i, node := range *nodes {
		if node.IsLocal() {
			o.setLicenseToNode(&(*nodes)[i])
			o.setInfraSpecToNode(&(*nodes)[i])
			continue
		}

		n, err := o.askPeerNode(node)
		if err == nil {
			(*nodes)[i].ManagementIP = n.ManagementIP
			(*nodes)[i].StorageIP = n.StorageIP
			(*nodes)[i].Vcpu = n.Vcpu
			(*nodes)[i].Memory = n.Memory
			(*nodes)[i].Storage = n.Storage
			(*nodes)[i].CpuSpec = n.CpuSpec
			(*nodes)[i].NetworkInterfaces = n.NetworkInterfaces
			(*nodes)[i].BlockDevices = n.BlockDevices
			(*nodes)[i].License = n.License
			(*nodes)[i].Status = n.Status
			(*nodes)[i].UptimeSeconds = n.UptimeSeconds
		}
	}
}

func (o *Operator) askPeerNode(node definition.Node) (*definition.Node, error) {
	h := http.GetGlobalHelper()
	resp, err := h.R().
		SetResult(&api.NodeData{}).
		SetHeader(node.GenAuthHeader()).
		Get(node.GetNodeDetailsUrl())
	if err != nil {
		log.Errorf("nodes: failed to get node details %s: %s", node.Hostname, err.Error())
		return nil, err
	}
	if resp.IsError() {
		err := fmt.Errorf("get error for node details %s: %d(%s)", node.Hostname, resp.StatusCode(), string(resp.Body()))
		log.Errorf("nodes: %v", err)
		return nil, err
	}

	return &resp.Result().(*api.NodeData).Data, nil
}
