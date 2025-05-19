package licenses

import (
	"fmt"
	"os"

	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/licenses"
	bslog "github.com/bigstack-oss/cube-cos-api/internal/log"
	"github.com/fsnotify/fsnotify"
)

func (o *Operator) initWatcher() error {
	var err error
	o.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	os.MkdirAll(licenses.Dir, os.ModeDir)
	return o.watcher.Add(licenses.Dir)
}

func syncLicense(event fsnotify.Event) {
	if !cubecos.IsLicenseFile(event.Name) {
		return
	}

	if shouldSync(event) {
		bslog.Throttle("licenses", fmt.Sprintf("%s changed, syncing license", event.Name))
		cubecos.SyncSourceLicense()
	}
}

func shouldSync(event fsnotify.Event) bool {
	return event.Has(fsnotify.Create) ||
		event.Has(fsnotify.Write) ||
		event.Has(fsnotify.Remove)
}
