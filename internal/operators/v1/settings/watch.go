package settings

import (
	"fmt"

	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/settings"
	cubelog "github.com/bigstack-oss/cube-cos-api/internal/log"
	"github.com/fsnotify/fsnotify"
	log "go-micro.dev/v5/logger"
)

func (o *Operator) initWatcher() error {
	var err error
	o.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	err = o.watcher.Add(settings.PolicyDir)
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
				syncCosAlertSetting(event)
			}
		case err, ok := <-o.watcher.Errors:
			if !ok {
				continue
			}
			if err != nil {
				log.Errorf("settings: failed to fetch setting change event: %v", err)
				continue
			}
		}
	}
}

func syncCosAlertSetting(event fsnotify.Event) {
	if !cubecos.IsAlertSetting(event.Name) {
		return
	}

	if event.Has(fsnotify.Create) || event.Has(fsnotify.Write) {
		cubelog.Throttle("settings", fmt.Sprintf("alert setting %s created or updated", event.Name))
		cubecos.SyncAlertSettings()
	}
}
