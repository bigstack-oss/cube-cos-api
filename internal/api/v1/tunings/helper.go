package tunings

import (
	"sort"

	"github.com/bigstack-oss/cube-cos-api/internal/api"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mohae/deepcopy"
	log "go-micro.dev/v5/logger"
)

type helper struct {
	c       *gin.Context
	handler string
	tuning  definition.Tuning

	allNodes bool
	hosts    []string
	keyword  string

	definition.Page
	watch bool
}

func initReqHelper(c *gin.Context, handler string) (*helper, error) {
	h := &helper{c: c, handler: handler}
	err := h.parsePage()
	if err != nil {
		return nil, err
	}

	h.parseScope()
	h.parseKeyword()
	h.parseWatch()
	h.parseHosts()

	return h, nil
}

func (h *helper) parseTuningUpdate() error {
	tuning, err := h.decodeTuningReq(h.c.Request.Body)
	if err != nil {
		log.Errorf("request(%s): failed to decode tuning request: %s", api.GetReqId(h.c), err.Error())
		return err
	}

	h.tuning = *tuning
	h.tuning.Id = uuid.New().String()
	h.tuning.Name = h.c.Param("parameterName")
	h.tuning.InitStatus("updating", "update")
	return nil
}

func (h *helper) parseTuningReset() error {
	tuning, err := h.decodeTuningReq(h.c.Request.Body)
	if err != nil {
		log.Errorf("request(%s): failed to decode tuning request: %s", api.GetReqId(h.c), err.Error())
		return err
	}

	name := h.c.Param("parameterName")
	spec, err := definition.GetTuningSpec(name)
	if err != nil {
		return err
	}

	h.tuning = *tuning
	h.tuning.Id = uuid.New().String()
	h.tuning.Name = name
	h.tuning.Value = spec.Limitation.Default
	h.tuning.Enabled = true
	h.tuning.IsModified = false
	h.tuning.InitStatus("updating", "reset")
	return nil
}

func (h *helper) ListTunings() (*data, error) {
	tunings, err := cubecos.ListTunings(definition.ListTuningOptions{AllNodes: h.allNodes})
	if err != nil {
		log.Errorf("request(%s): failed to get tunings: %s", api.GetReqId(h.c), err.Error())
		return nil, err
	}

	tunings = h.filterTunings(tunings)
	h.sortTunings(&tunings)

	pagedTunings, err := h.paginateTunings(tunings)
	if err != nil {
		log.Errorf("request(%s): failed to paginate tunings: %s", api.GetReqId(h.c), err.Error())
		return nil, err
	}

	page, err := h.genPageInfo(tunings)
	if err != nil {
		log.Errorf("request(%s): failed to gen page info: %s", api.GetReqId(h.c), err.Error())
		return nil, err
	}

	return &data{
		Tunings: pagedTunings,
		Page:    page,
	}, nil
}

func (h *helper) ListTuningSpecs() ([]definition.TuningSpec, error) {
	specs := []definition.TuningSpec{}
	definition.GetTuningSpecs().Range(func(key, value any) bool {
		spec := deepcopy.Copy(value).(*definition.TuningSpec)
		spec.Roles = selectRolesUsingActivityAndLabels(spec)
		specs = append(specs, *spec)
		return true
	})

	h.sortTuningSpecs(&specs)
	return specs, nil
}

func (h *helper) sortTuningSpecs(specs *[]definition.TuningSpec) {
	sort.Slice(*specs, func(i, j int) bool {
		return (*specs)[i].Name < (*specs)[j].Name
	})
}
