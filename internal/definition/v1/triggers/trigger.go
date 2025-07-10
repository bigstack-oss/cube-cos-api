package triggers

import (
	"sync/atomic"
)

const (
	Module = "triggers"

	DryRunNamespace = "triggers-scripts-dry-run"
)

var (
	list = atomic.Pointer[[]ApiSchema]{}
)

func SyncList(triggers []ApiSchema) {
	list.Swap(&triggers)
}

func List() []ApiSchema {
	schema := list.Load()
	if schema == nil {
		return []ApiSchema{}
	}

	return *schema
}
