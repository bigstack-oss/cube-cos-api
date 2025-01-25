package v1

import (
	"runtime"

	log "go-micro.dev/v5/logger"
)

func CapturePanic() {
	if r := recover(); r != nil {
		buf := make([]byte, 1<<16)
		stackSize := runtime.Stack(buf, true)
		log.Errorf("panic captured: %v\n stack trace:\n%s", r, buf[:stackSize])
	}
}
