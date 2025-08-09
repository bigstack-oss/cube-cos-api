package triggers

import (
	"fmt"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	conf "github.com/bigstack-oss/cube-cos-api/internal/config"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/triggers"
	bslog "github.com/bigstack-oss/cube-cos-api/internal/log"
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

	o.syncTriggers()
	go o.watchChanges()
	return nil
}

func (o *Operator) watchChanges() {
	for {
		select {
		case event, ok := <-o.policy.Events:
			if ok {
				o.checkTriggers(event)
			}
		case err, ok := <-o.policy.Errors:
			if !ok {
				continue
			}
			if err != nil {
				log.Errorf("triggers: failed to fetch policy change event(%v)", err)
				continue
			}
		}
	}
}

func (o *Operator) checkTriggers(event fsnotify.Event) {
	if event.Name != conf.Opts.Spec.Identity.Policy {
		return
	}

	if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) {
		bslog.Throttle("triggers", fmt.Sprintf("%s changed, syncing triggers", event.Name))
		o.syncTriggers()
	}
}

func (o *Operator) syncTriggers() {
	list, err := o.tryGetTriggers()
	if err != nil {
		log.Errorf("triggers: failed to sync triggers(%v)", err)
		return
	}

	triggers.SyncList(list)
}

func (o *Operator) tryGetTriggers() ([]triggers.Trigger, error) {
	for range 60 {
		list, err := cubecos.GetTriggers()
		if err == nil {
			return list, nil
		}

		log.Errorf("triggers: failed to get triggers, retrying(%v)", err)
		wait.Seconds(3)
	}

	return nil, fmt.Errorf(
		"triggers: failed to get triggers after retries",
	)
}
