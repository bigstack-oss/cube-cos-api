package triggers

import (
	"sync/atomic"
)

const (
	Module = "triggers"

	MaxDryRunJobs   = 15
	DryRunNamespace = "triggers-scripts-dry-run"
	DryRunOciImage  = "localhost:5080/bigstack/shell:latest"
)

var (
	list = atomic.Pointer[[]Trigger]{}
)

func SyncList(triggers []Trigger) {
	list.Swap(&triggers)
}

func List() []Trigger {
	trigger := list.Load()
	if trigger == nil {
		return []Trigger{}
	}

	return *trigger
}
