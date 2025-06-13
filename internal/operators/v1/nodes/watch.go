package node

import (
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/pacemaker"
	"github.com/fsnotify/fsnotify"
	log "go-micro.dev/v5/logger"
)

func (o *Operator) initPacemakerWatcher() error {
	var err error
	o.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	err = o.watcher.Add(pacemaker.AlertLogDir)
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
				syncPacemakerStatus(event)
			}
		case err, ok := <-o.watcher.Errors:
			if !ok {
				continue
			}
			if err != nil {
				log.Errorf("nodes: failed to fetch pacemaker event(%v)", err)
				continue
			}
		}
	}
}

func syncPacemakerStatus(event fsnotify.Event) {
	if event.Name != pacemaker.AlertRecord {
		return
	}

	if event.Has(fsnotify.Create) || event.Has(fsnotify.Write) {
		log.Info("node: sync pacemaker update")
		nodes.Sync()
	}
}
