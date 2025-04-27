package licenses

import (
	"context"

	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/service"
	"github.com/fsnotify/fsnotify"
	log "go-micro.dev/v5/logger"
)

var (
	module = "licenses"
)

func init() {
	service.RegisterOperator(module, NewOperator())
}

type Operator struct {
	watcher *fsnotify.Watcher
	ctx     context.Context
	cancel  context.CancelFunc
}

func NewOperator() *Operator {
	return &Operator{}
}

func (o *Operator) Name() string {
	return module
}

func (o *Operator) Init() error {
	cubecos.SyncSourceLicense()
	o.ctx, o.cancel = context.WithCancel(context.Background())
	return o.initWatcher()
}

func (o *Operator) Run() {
	for {
		select {
		case <-o.ctx.Done():
			return
		case event, ok := <-o.watcher.Events:
			if ok {
				syncLicense(event)
			}
		case err, ok := <-o.watcher.Errors:
			if !ok {
				continue
			}
			if err != nil {
				log.Errorf("licenses: failed to fetch license change event: %s", err.Error())
				continue
			}
		}
	}
}

func (o *Operator) Stop() {
	o.cancel()
	if o.watcher != nil {
		_ = o.watcher.Close()
	}
}
