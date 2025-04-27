package licenses

import (
	"fmt"

	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	cubelog "github.com/bigstack-oss/cube-cos-api/internal/log"
	"github.com/fsnotify/fsnotify"
)

func (o *Operator) initWatcher() error {
	var err error
	o.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	return o.watcher.Add(v1.LicenseDir)
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
