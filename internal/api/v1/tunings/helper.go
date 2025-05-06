package tunings

import (
	"errors"
	"sort"

	"github.com/bigstack-oss/cube-cos-api/internal/api"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/gin-gonic/gin"
	"github.com/mohae/deepcopy"
	log "go-micro.dev/v5/logger"
)

type helper struct {
	c       *gin.Context
	handler string
	tuning  v1.Tuning
	toggle  v1.TuningToggle
	update  v1.TuningUpdate
	reset   v1.TuningReset

	allNodes bool
	hosts    []string
	keyword  string
	modified []bool

	v1.Page
	watch bool
}

func initHelper(c *gin.Context, handler string) (*helper, error) {
	h := &helper{c: c, handler: handler}
	err := h.parsePage()
	if err != nil {
		return nil, err
	}

	h.parseScope()
	h.parseKeyword()
	h.parseWatch()
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
		log.Errorf("tunings(%s): failed to parse reset tuning: %s", api.GetReqId(h.c), err.Error())
		return err
	}

	name := h.c.Param("parameterName")
	spec, err := v1.GetTuningSpec(name)
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
		log.Errorf("tunings(%s): failed to get tuning: %s", api.GetReqId(h.c), err.Error())
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
		log.Errorf("tunings(%s): failed to get tuning: %s", api.GetReqId(h.c), err.Error())
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
	tunings, err := cubecos.ListTunings(v1.ListTuningOptions{AllNodes: h.allNodes})
	if err != nil {
		log.Errorf("tunings(%s): failed to get tunings: %s", api.GetReqId(h.c), err.Error())
		return nil, err
	}

	tunings = h.filterTunings(tunings)
	h.enrichTuningPayload(&tunings)

	pagedTunings, err := h.paginateTunings(tunings)
	if err != nil {
		log.Errorf("tunings(%s): failed to paginate tunings: %s", api.GetReqId(h.c), err.Error())
		return nil, err
	}

	page, err := h.genPageInfo(tunings)
	if err != nil {
		log.Errorf("tunings(%s): failed to gen page info: %s", api.GetReqId(h.c), err.Error())
		return nil, err
	}

	return &data{
		Tunings: pagedTunings,
		Page:    page,
	}, nil
}

func (h *helper) ListTuningSpecs() ([]v1.TuningSpec, error) {
	specs := []v1.TuningSpec{}
	v1.GetTuningSpecs().Range(func(key, value any) bool {
		spec := deepcopy.Copy(value).(*v1.TuningSpec)
		spec.Roles = selectRolesUsingActivityAndLabels(spec)
		specs = append(specs, *spec)
		return true
	})

	h.sortTuningSpecs(&specs)
	return specs, nil
}

func (h *helper) sortTuningSpecs(specs *[]v1.TuningSpec) {
	sort.Slice(*specs, func(i, j int) bool {
		return (*specs)[i].Name < (*specs)[j].Name
	})
}
