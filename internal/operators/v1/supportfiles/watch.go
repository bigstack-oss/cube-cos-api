package supportfiles

import (
	"context"
	"path/filepath"
	"time"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/support"
	"github.com/cnf/structhash"
	"github.com/fsnotify/fsnotify"
	"go-micro.dev/v5/cache"
	log "go-micro.dev/v5/logger"
)

var (
	logCache = cache.NewCache(cache.Expiration(time.Second * 3))
)

func (o *Operator) initWatcher() error {
	var err error
	o.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	err = o.watcher.Add(support.DefaultFileDir)
	if err != nil {
		return err
	}

	go o.watchChanges()
	return nil
}

func (o *Operator) watchChanges() {
	for {
		select {
		case event, ok := <-o.watcher.Events:
			if ok {
				checkAndSyncSupportFiles(event)
			}
		case err, ok := <-o.watcher.Errors:
			if !ok {
				continue
			}
			if err != nil {
				log.Errorf("supportFiles: failed to fetch support file change event: %s", err.Error())
				continue
			}
		}
	}
}

func checkAndSyncSupportFiles(event fsnotify.Event) {
	filename := filepath.Base(event.Name)
	if !cubecos.IsSupportFile(filename) {
		return
	}

	if event.Has(fsnotify.Create) {
		printOrThrottleLog(event)
		cubecos.SyncSupportFiles()
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

	log.Infof("supportFiles: %s changed, syncing supportFiles", event.Name)
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
