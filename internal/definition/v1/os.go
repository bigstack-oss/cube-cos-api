package v1

import (
	"os"
	"runtime"
	"strings"

	log "go-micro.dev/v5/logger"
)

var (
	SerialNumber = ""
)

func CapturePanic() {
	recovery := recover()
	if recovery != nil {
		buf := make([]byte, 1<<16)
		stackSize := runtime.Stack(buf, true)
		log.Errorf("panic: captured %v\n stack trace:\n%s", recovery, buf[:stackSize])
	}
}

func GetSystemSerial(file string) (string, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return "no serial number found", err
	}

	serial := strings.TrimSpace(string(data))
	return serial, nil
}
