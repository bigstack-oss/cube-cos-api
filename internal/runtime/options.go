package runtime

import (
	"fmt"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/influx"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/mongo"
	conf "github.com/bigstack-oss/cube-cos-api/internal/config"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/base"
)

func parseMongoOpts() mongo.Options {
	mongo := conf.Opts.Spec.Store.MongoDB
	if mongo.Host == "" {
		mongo.Host = base.DataCenterVip
	}

	mongo.Uri = fmt.Sprintf("mongodb://%s:%d", mongo.Host, mongo.Port)
	return mongo
}

func parseInfluxOpts() influx.Options {
	influx := conf.Opts.Spec.Store.InfluxDB
	if influx.Host == "" {
		influx.Host = base.DataCenterVip
	}

	influx.Url = fmt.Sprintf("%s://%s:%d", influx.Protocol, influx.Host, influx.Port)
	return influx
}
