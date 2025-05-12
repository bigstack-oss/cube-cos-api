package tunings

import (
	"errors"
	"sort"

	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/queries"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/pages"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/tunings"
	"github.com/gin-gonic/gin"
	"github.com/mohae/deepcopy"
	log "go-micro.dev/v5/logger"
)

type helper struct {
	c       *gin.Context
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

func (h *helper) parseTuningUpdate() error {
	err := h.c.ShouldBindJSON(&h.update)
	if err != nil {

		return err
	}

	h.convertUpdateToTuning()
	return nil
}

func (h *helper) parseTuningReset() error {
	err := h.c.ShouldBindJSON(&h.reset)
	if err != nil {
		log.Errorf("tunings(%s): failed to parse reset tuning: %v", h.reqId, err)
		return err
	}

	name := h.c.Param("parameterName")
	spec, err := tunings.GetSpec(name)
	if err != nil {
		return err
	}

	h.tuning.Name = name
	if !h.isTuningModified() {
		return errors.New("can't reset unmodified tuning")
	}

	h.tuning.Value = spec.Limitation.Default
	h.tuning.Enabled = true
	h.tuning.IsModified = false
	h.tuning.Id = h.tuning.GenerateId()
	h.tuning.InitResetStatus()
	h.tuning.InitHosts(h.reset.Hosts)
	return nil
}

func (h *helper) parseTuningEnablement() error {
	err := h.c.ShouldBindJSON(&h.toggle)
	if err != nil {
		return err
	}

	h.tuning.Name = h.c.Param("parameterName")
	if !h.isTuningModified() {
		return errors.New("can't enable/disable unmodified tuning")
	}

	tuning, err := h.getTuningByNameAndHosts(h.tuning.Name, h.toggle.Hosts)
	if err != nil {
		log.Errorf("tunings(%s): failed to get tuning: %v", h.reqId, err)
		return err
	}

	h.tuning = *tuning
	h.tuning.Enabled = h.toggle.Enable
	h.tuning.InitUpdateStatus()
	h.tuning.InitHosts(h.toggle.Hosts)
	return nil
}

func (h *helper) isTuningModified() bool {
	tuning, err := h.getTuningByNameAndHosts(h.tuning.Name, h.toggle.Hosts)
	if err != nil {
		log.Errorf("tunings(%s): failed to get tuning: %v", h.reqId, err)
		return false
	}

	return tuning.IsModified
}

func (h *helper) convertUpdateToTuning() {
	h.tuning.Name = h.c.Param("parameterName")
	h.tuning.Enabled = h.update.Enabled
	h.tuning.Value = h.update.Value
	h.tuning.InitUpdateStatus()
	h.tuning.InitHosts(h.update.Hosts)
}

func (h *helper) ListTunings() (*data, error) {
	tunings, err := cubecos.ListTunings(tunings.ListOptions{AllNodes: h.allNodes})
	if err != nil {
		log.Errorf("tunings(%s): failed to get tunings: %v", h.reqId, err)
		return nil, err
	}

	h.enrichTunings(&tunings)
	tunings = h.filterTunings(tunings)
	pagedTunings, err := h.paginateTunings(tunings)
	if err != nil {
		log.Errorf("tunings(%s): failed to paginate tunings: %v", h.reqId, err)
		return nil, err
	}

	page, err := h.genPageInfo(tunings)
	if err != nil {
		log.Errorf("tunings(%s): failed to gen page info: %v", h.reqId, err)
		return nil, err
	}

	return &data{
		Tunings: pagedTunings,
		Page:    page,
	}, nil
}

func (h *helper) ListTuningSpecs() ([]tunings.Spec, error) {
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
