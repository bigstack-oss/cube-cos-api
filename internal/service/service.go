package service

import (
	"time"

	hystrix "github.com/micro/plugins/v5/wrapper/breaker/hystrix"
	"github.com/micro/plugins/v5/wrapper/monitoring/prometheus"
	ratelimit "github.com/micro/plugins/v5/wrapper/ratelimiter/uber"
	"go-micro.dev/v5"
	log "go-micro.dev/v5/logger"
	"go-micro.dev/v5/registry"
	"go-micro.dev/v5/server"
)

var (
	Controllers map[string]Controller
)

type Controller interface {
	Name() string
	Sync()
	Stop()
}

func init() {
	Controllers = make(map[string]Controller)
}

func RegisterController(name string, controller Controller) {
	Controllers[name] = controller
}

func WrapGoMicro(server *server.Server) micro.Service {
	return micro.NewService(
		micro.Server(*server),
		micro.WrapClient(hystrix.NewClientWrapper()),
		micro.WrapHandler(ratelimit.NewHandlerWrapper(10)),
		micro.WrapHandler(prometheus.NewHandlerWrapper()),
		micro.Registry(registry.NewRegistry()),
		micro.RegisterTTL(time.Second*60),
		micro.RegisterInterval(time.Second*20),
		micro.AfterStart(runControllers),
		micro.AfterStop(stopControllers),
	)
}

func runControllers() error {
	for _, c := range Controllers {
		go run(c.Sync)
		log.Infof("Controller: %s is running", c.Name())
	}

	return nil
}

func run(f func()) {
	for {
		f()
	}
}

func stopControllers() error {
	for _, c := range Controllers {
		log.Infof("Controller: %s is shutting down", c.Name())
		c.Stop()
	}

	return nil
}
