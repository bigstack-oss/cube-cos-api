package service

import (
	"time"

	hystrix "github.com/micro/plugins/v5/wrapper/breaker/hystrix"
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
	Run()
	Stop()
}

func init() {
	Operators = make(map[string]Operator)
}

func RegisterOperator(name string, operator Operator) {
	Operators[name] = operator
}

func Micro(server *server.Server) micro.Service {
	return micro.NewService(
		micro.Server(*server),
		micro.WrapClient(hystrix.NewClientWrapper()),
		micro.WrapHandler(ratelimit.NewHandlerWrapper(10)),
		micro.Registry(registry.NewRegistry()),
		micro.RegisterTTL(time.Second*60),
		micro.RegisterInterval(time.Second*20),
		micro.AfterStart(runOperators),
		micro.AfterStop(stopOperators),
	)
}

func runOperators() error {
	for _, operator := range Operators {
		err := operator.Init()
		if err != nil {
			log.Errorf("operator: %s init failed: %v: ", operator.Name(), err)
			return err
		}

		go operator.Run()
		log.Infof("operator: %s is running", operator.Name())
	}

	return nil
}

func stopOperators() error {
	for _, o := range Operators {
		log.Infof("operator: %s is shutting down", o.Name())
		o.Stop()
	}

	return nil
}
