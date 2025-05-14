package supportfiles

import (
	"fmt"
	"net/url"

	conf "github.com/bigstack-oss/cube-cos-api/internal/config"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/base"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/support"
)

func (h *helper) syncHostPortInUrl(files *[]support.File) {
	for i, file := range *files {
		url, err := url.Parse(file.Url)
		if err != nil {
			continue
		}

		url.Host = fmt.Sprintf(
			"%s:%d", base.DataCenterVip,
			conf.Opts.Saml.ServiceProvider.Host.Port,
		)

		(*files)[i].Url = url.String()
	}
}
