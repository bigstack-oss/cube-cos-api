package runtime

import (
	"fmt"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/influx"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/mongo"
	conf "github.com/bigstack-oss/cube-cos-api/internal/config"
	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
)

func parseMongoOpts() mongo.Options {
	if conf.Opts.Spec.Store.MongoDB.Host == "" {
		conf.Opts.Spec.Store.MongoDB.Host = v1.DataCenterVip
	}

	conf.Opts.Spec.Store.MongoDB.Uri = fmt.Sprintf(
		"mongodb://%s:%d",
		conf.Opts.Spec.Store.MongoDB.Host,
		conf.Opts.Spec.Store.MongoDB.Port,
	)

	return conf.Opts.Spec.Store.MongoDB
}

func parseInfluxOpts() influx.Options {
	if conf.Opts.Spec.Store.InfluxDB.Host == "" {
		conf.Opts.Spec.Store.InfluxDB.Host = v1.DataCenterVip
	}

	conf.Opts.Spec.Store.InfluxDB.Url = fmt.Sprintf(
		"%s://%s:%d",
		conf.Opts.Spec.Store.InfluxDB.Protocol,
		conf.Opts.Spec.Store.InfluxDB.Host,
		conf.Opts.Spec.Store.InfluxDB.Port,
	)

	return conf.Opts.Spec.Store.InfluxDB
}
