package log

import (
	"context"
	"time"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	"go-micro.dev/v5/cache"
	log "go-micro.dev/v5/logger"
)

var (
	throttleCache = cache.NewCache(cache.Expiration(time.Second * 3))
)

func Throttle(module, msg string) {
	key := key(module, msg)
	if isLogThrottled(key) {
		return
	}

	log.Infof("%s: %s", module, msg)
	setThrottling(key)
}

func isLogThrottled(key string) bool {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(10))
	defer cancel()
	_, _, err := throttleCache.Get(ctx, key)
	return err == nil
}

func setThrottling(key string) error {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(10))
	defer cancel()
	return throttleCache.Put(
		ctx,
		key,
		[]byte{},
		time.Second*3,
	)
}

func key(strs ...string) string {
	if len(strs) == 0 {
		return ""
	}

	key := strs[0]
	for i := 1; i < len(strs); i++ {
		key += ":" + strs[i]
	}

	return key
}
