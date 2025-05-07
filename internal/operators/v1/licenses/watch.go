package licenses

import (
	"fmt"
	"os"

	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/license"
	cubelog "github.com/bigstack-oss/cube-cos-api/internal/log"
	"github.com/fsnotify/fsnotify"
)

func (o *Operator) initWatcher() error {
	var err error
	o.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	os.MkdirAll(license.Dir, os.ModeDir)
	return o.watcher.Add(license.Dir)
}

func syncLicense(event fsnotify.Event) {
	if !cubecos.IsLicenseFile(event.Name) {
		return
	}

	if event.Has(fsnotify.Create) || event.Has(fsnotify.Write) {
		cubelog.Throttle("licenses", fmt.Sprintf("%s changed, syncing license", event.Name))
		cubecos.SyncSourceLicense()
	}
}
