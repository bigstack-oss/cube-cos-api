package supportfiles

import (
	"fmt"
	"path/filepath"

	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/support"
	bslog "github.com/bigstack-oss/cube-cos-api/internal/log"
	"github.com/fsnotify/fsnotify"
	log "go-micro.dev/v5/logger"
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
				syncSourceSupportFiles(event)
			}
		case err, ok := <-o.watcher.Errors:
			if !ok {
				continue
			}
			if err != nil {
				log.Errorf("supportfiles: failed to fetch support file change event(%v)", err)
				continue
			}
		}
	}
}

func syncSourceSupportFiles(event fsnotify.Event) {
	filename := filepath.Base(event.Name)
	if !cubecos.IsSupportFile(filename) {
		return
	}

	if shouldSync(event) {
		bslog.Throttle("supportFiles", fmt.Sprintf("support file %s created", event.Name))
		cubecos.SyncSupportFiles()
	}
}

func shouldSync(event fsnotify.Event) bool {
	return event.Has(fsnotify.Create) ||
		event.Has(fsnotify.Write) ||
		event.Has(fsnotify.Rename) ||
		event.Has(fsnotify.Remove)
}
