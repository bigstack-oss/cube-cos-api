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
	if r := recover(); r != nil {
		buf := make([]byte, 1<<16)
		stackSize := runtime.Stack(buf, true)
		log.Errorf("panic captured: %v\n stack trace:\n%s", r, buf[:stackSize])
	}
}

func GetSystemSerial() (string, error) {
	data, err := os.ReadFile("/sys/class/dmi/id/product_serial")
	if err != nil {
		return "", err
	}

	serial := strings.TrimSpace(string(data))
	return serial, nil
}
