package tunings

import (
	conf "github.com/bigstack-oss/cube-cos-api/internal/config"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/fsnotify/fsnotify"
	log "go-micro.dev/v5/logger"
)

func (o *Operator) initPolicyWatcher() error {
	var err error
	o.policy, err = fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	err = o.policy.Add(conf.Opts.Spec.Identity.Policy)
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
				log.Errorf("failed to fetch policy change event: %s", err.Error())
				continue
			}
		}
	}
}

func checkAndSyncTunings(event fsnotify.Event) {
	if event.Has(fsnotify.Write) {
		log.Infof("tunings: policy file changed, syncing tunings")
		cubecos.SyncTunings()
	}
}
