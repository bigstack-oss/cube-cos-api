package firmwares

import (
	"errors"
	"sync"
	"time"

	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/firmwares"
	log "go-micro.dev/v5/logger"
)

// rollCacheTTL bounds how long a cached roll projection is served.
const rollCacheTTL = 2 * time.Minute

// lastRoll is the last roll projection served; a var-held clock for tests.
var (
	lastRoll struct {
		sync.Mutex
		upgrade *firmwares.Upgrade
		at      time.Time
	}
	rollCacheNow = time.Now
)

// fallBackToCachedRoll caches each roll projection and serves the last one
// while the roll job is unreadable; with none fresh, the error stands.
func (h *helper) fallBackToCachedRoll(upgrade *firmwares.Upgrade, err error) (*firmwares.Upgrade, error) {
	lastRoll.Lock()
	defer lastRoll.Unlock()

	if err == nil {
		lastRoll.upgrade, lastRoll.at = nil, time.Time{}
		if upgrade.Kind != "" {
			lastRoll.upgrade, lastRoll.at = copyRollUpgrade(upgrade), rollCacheNow()
		}

		return upgrade, nil
	}

	if !errors.Is(err, cubecos.ErrRollStateUnavailable) || lastRoll.upgrade == nil {
		return nil, err
	}

	if rollCacheNow().Sub(lastRoll.at) > rollCacheTTL {
		return nil, err
	}

	log.Warnf("firmwares(%s): serving the last roll projection(%v)", h.reqId, err)
	return copyRollUpgrade(lastRoll.upgrade), nil
}

func copyRollUpgrade(upgrade *firmwares.Upgrade) *firmwares.Upgrade {
	c := *upgrade
	c.Progresses = append([]firmwares.Progress(nil), upgrade.Progresses...)
	return &c
}
