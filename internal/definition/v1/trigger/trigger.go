package trigger

import "sync"

var (
	updateList = sync.Mutex{}
)

func SyncList(triggers []ApiOptions) {
	updateList.Lock()
	defer updateList.Unlock()
	list = triggers
}

func List() []ApiOptions {
	return list
}
