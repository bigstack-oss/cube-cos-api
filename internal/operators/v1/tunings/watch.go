package tunings

import (
	"fmt"

	conf "github.com/bigstack-oss/cube-cos-api/internal/config"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	cubelog "github.com/bigstack-oss/cube-cos-api/internal/log"
	"github.com/fsnotify/fsnotify"
	log "go-micro.dev/v5/logger"
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
				log.Errorf("tunings: failed to fetch policy change event: %v", err)
				continue
			}
		}
	}
}

func checkAndSyncTunings(event fsnotify.Event) {
	if event.Name != conf.Opts.Spec.Identity.Policy {
		return
	}

	if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) {
		cubelog.Throttle("tunings", fmt.Sprintf("%s changed, syncing tunings", event.Name))
		cubecos.SyncTunings()
	}
}
