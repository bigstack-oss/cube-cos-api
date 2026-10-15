package storages

import (
	"context"
	"fmt"
	"sync"
	ostime "time"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/mongo"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/storages"
	log "go-micro.dev/v5/logger"
	"go.mongodb.org/mongo-driver/bson"
)

// The receiving control restarts cinder-volume once every control has applied.

const (
	peerWaitBound = 10 * ostime.Minute
	peerPollEvery = 5 * ostime.Second
)

// restartOps wraps mongo, hex_sdk and the clock; replaced in tests
var restartOps = struct {
	pending func(name string, hosts []string) ([]string, error)
	restart func() error
	now     func() ostime.Time
	sleep   func(ostime.Duration)
}{
	pending: pendingPeers,
	restart: cubecos.RestartCinderVolume,
	now:     ostime.Now,
	sleep:   ostime.Sleep,
}

// one cinder-volume restart at a time on this node
var restartMu sync.Mutex

// processStorageReq applies the request; the receiving control reports only
// after the peers apply and cinder-volume restarts
func (o *Operator) processStorageReq(req *storages.ReqOpts) {
	err := o.operateStorageReq(req)
	if req.IsOrigin && req.ChangesBackends() {
		go o.restartAfterPeers(*req, err)
		return
	}

	o.handleStorageExit(*req, err)
}

func (o *Operator) restartAfterPeers(req storages.ReqOpts, applyErr error) {
	name := req.CinderDetails.Name
	left := waitForPeers(name, req.Peers)
	if len(left) > 0 {
		log.Warnf("storages: %s: no result from %v after %s, restarting cinder-volume anyway", name, left, peerWaitBound)
	}

	restartMu.Lock()
	err := restartOps.restart()
	restartMu.Unlock()
	if err != nil {
		log.Errorf("storages: %s: failed to restart cinder-volume(%v)", name, err)
		if applyErr == nil {
			applyErr = fmt.Errorf("restart cinder-volume: %w", err)
		}
	}

	o.handleStorageExit(req, applyErr)
}

// waitForPeers returns the peers still applying when the bound ran out
func waitForPeers(name string, peers []string) []string {
	if len(peers) == 0 {
		return nil
	}

	deadline := restartOps.now().Add(peerWaitBound)
	for {
		left, err := restartOps.pending(name, peers)
		if err != nil {
			log.Errorf("storages: %s: failed to check peer results(%v)", name, err)
			left = peers
		}
		if len(left) == 0 {
			return nil
		}
		if !restartOps.now().Before(deadline) {
			return left
		}

		restartOps.sleep(peerPollEvery)
	}
}

// pendingPeers lists the hosts whose request record still exists
func pendingPeers(name string, hosts []string) ([]string, error) {
	cur, err := mongo.GetGlobalHelper().GetQueryCursor(
		storages.Db,
		storages.ReqCollection,
		bson.M{
			"cinderDetails.name": name,
			"hostname":           bson.M{"$in": hosts},
		},
	)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(wait.CtxSeconds(5))
	defer cancel()
	defer cur.Close(ctx)

	reqs := []storages.ReqOpts{}
	if err := cur.All(ctx, &reqs); err != nil {
		return nil, err
	}

	left := []string{}
	for _, req := range reqs {
		left = append(left, req.Hostname)
	}

	return left, nil
}
