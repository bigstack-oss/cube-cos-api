package tunings

import (
	"sort"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/http"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/mongo"
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/queries"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/pages"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/tunings"
	"github.com/gin-gonic/gin"
	"github.com/mohae/deepcopy"
	log "go-micro.dev/v5/logger"
)

type helper struct {
	c       *gin.Context
	http    *http.Helper
	mongo   *mongo.Helper
	reqId   string
	handler string

	tuning tunings.Tuning
	toggle tunings.Toggle
	update tunings.Update
	reset  tunings.Reset

	allNodes bool
	hosts    []string
	keyword  string
	modified []bool

	*pages.Page
	watch bool
}

func initHelper(c *gin.Context, handler string) (*helper, error) {
	h := &helper{
		c:       c,
		http:    http.GetGlobalHelper(),
		mongo:   mongo.GetGlobalHelper(),
		reqId:   queries.GetReqId(c),
		handler: handler,
	}

	err := h.parsePage()
	if err != nil {
		return nil, err
	}

	err = h.parseWatch()
	if err != nil {
		return nil, err
	}

	h.parseScope()
	h.parseKeyword()
	h.parseHosts()
	h.parseModified()

	return h, nil
}

func (h *helper) listTunings() (*tuningPage, error) {
	tunings, err := h.listAggregatedTunings()
	if err != nil {
		log.Errorf("tunings(%s): failed to get parameters: %v", h.reqId, err)
		return nil, err
	}

	tunings = h.filterTunings(tunings)
	pagedTunings, err := h.paginateTunings(tunings)
	if err != nil {
		log.Errorf("tunings(%s): failed to paginate parameters: %v", h.reqId, err)
		return nil, err
	}

	page, err := h.genPageInfo(tunings)
	if err != nil {
		log.Errorf("tunings(%s): failed to gen page info: %v", h.reqId, err)
		return nil, err
	}

	return &tuningPage{
		Tunings: pagedTunings,
		Page:    page,
	}, nil
}

func (h *helper) listTuningSpecs() ([]tunings.Spec, error) {
	specs := []tunings.Spec{}
	tunings.GetSpecs().Range(func(key, value any) bool {
		spec := deepcopy.Copy(value).(*tunings.Spec)
		spec.Roles = selectRolesUsingActivityAndLabels(spec)
		specs = append(specs, *spec)
		return true
	})

	h.sortTuningSpecs(&specs)
	return specs, nil
}

func (h *helper) sortTuningSpecs(specs *[]tunings.Spec) {
	sort.Slice(*specs, func(i, j int) bool {
		return (*specs)[i].Name < (*specs)[j].Name
	})
}
