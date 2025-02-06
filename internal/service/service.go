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
	Operators map[string]Operator
)

type Operator interface {
	Name() string
	Init() error
	Sync()
	Stop()
}

func init() {
	Operators = make(map[string]Operator)
}

func RegisterOperator(name string, operator Operator) {
	Operators[name] = operator
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
		micro.AfterStart(runOperators),
		micro.AfterStop(stopOperators),
	)
}

func runOperators() error {
	for _, o := range Operators {
		err := o.Init()
		if err != nil {
			log.Errorf("operator: %s init failed: %v: ", o.Name(), err)
			return err
		}

		go runInBackground(o.Sync)
		log.Infof("operator: %s is running", o.Name())
	}

	return nil
}

func runInBackground(f func()) {
	for {
		f()
	}
}

func stopOperators() error {
	for _, o := range Operators {
		log.Infof("operator: %s is shutting down", o.Name())
		o.Stop()
	}

	return nil
}
