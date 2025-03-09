package triggers

import (
	"context"
	"time"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	conf "github.com/bigstack-oss/cube-cos-api/internal/config"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/cnf/structhash"
	"github.com/fsnotify/fsnotify"
	"go-micro.dev/v5/cache"
	log "go-micro.dev/v5/logger"
)

var (
	logCache = cache.NewCache(cache.Expiration(time.Second * 3))
)

func (o *Operator) initPolicyWatcher() error {
	var err error
	o.policy, err = fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	err = o.policy.Add("/etc")
	if err != nil {
		return err
	}

	go o.watchChanges()
	return nil
}

func (o *Operator) watchChanges() {
	for {
		select {
		case event, ok := <-o.policy.Events:
			if ok {
				checkAndSyncTunings(event)
			}
		case err, ok := <-o.policy.Errors:
			if !ok {
				continue
			}
			if err != nil {
				log.Errorf("tunings: failed to fetch policy change event: %s", err.Error())
				continue
			}
		}
	}
}

func checkAndSyncTunings(event fsnotify.Event) {
	if event.Name != conf.Opts.Spec.Identity.Policy {
		return
	}

	if event.Has(fsnotify.Write) {
		logThrottling(event)
		cubecos.SyncTunings()
	}
}

func logThrottling(event fsnotify.Event) {
	key, err := structhash.Hash(event, 1)
	if err != nil {
		return
	}

	getCtx, getCancel := context.WithTimeout(wait.CtxSeconds(10))
	defer getCancel()
	_, _, err = logCache.Get(getCtx, key)
	if err == nil {
		return
	}

	putCtx, putCancel := context.WithTimeout(wait.CtxSeconds(10))
	defer putCancel()
	err = logCache.Put(putCtx, key, []byte{}, time.Second*3)
	if err != nil {
		return
	}

	log.Infof("tunings: %s changed, syncing tunings", event.Name)
}
