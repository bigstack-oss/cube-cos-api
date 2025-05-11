package triggers

import "sync"

const (
	Module = "triggers"
)

var (
	updateList = sync.Mutex{}
)

func SyncList(triggers []ApiSchema) {
	updateList.Lock()
	defer updateList.Unlock()
	list = triggers
}

func List() []ApiSchema {
	return list
}
