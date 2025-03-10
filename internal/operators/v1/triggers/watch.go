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
				checkAndSyncTriggers(event)
			}
		case err, ok := <-o.policy.Errors:
			if !ok {
				continue
			}
			if err != nil {
				log.Errorf("triggers: failed to fetch policy change event: %s", err.Error())
				continue
			}
		}
	}
}

func checkAndSyncTriggers(event fsnotify.Event) {
	if event.Name != conf.Opts.Spec.Identity.Policy {
		return
	}

	if event.Has(fsnotify.Write) {
		printOrThrottleLog(event)
		cubecos.SyncTriggers()
	}
}

func printOrThrottleLog(event fsnotify.Event) {
	key, err := structhash.Hash(event, 1)
	if err != nil {
		return
	}

	if isLogThrottled(key) {
		return
	}

	log.Infof("triggers: %s changed, syncing triggers", event.Name)
	throttleLog(key)
}

func isLogThrottled(key string) bool {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(10))
	defer cancel()
	_, _, err := logCache.Get(ctx, key)
	return err == nil
}

func throttleLog(key string) error {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(10))
	defer cancel()
	return logCache.Put(
		ctx,
		key,
		[]byte{},
		time.Second*3,
	)
}
