package trigger

import "sync"

var (
	updateList = sync.Mutex{}
)

func SyncList(triggers []ApiOptions) {
	updateList.Lock()
	defer updateList.Unlock()
	List = triggers
}

func GetList() []ApiOptions {
	return List
}
