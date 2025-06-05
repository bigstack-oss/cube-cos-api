package tunings

import (
	"errors"
	"strings"

	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/queries"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/tunings"
	log "go-micro.dev/v5/logger"
)

func (h *helper) parseTuningUpdate() error {
	err := h.c.ShouldBindJSON(&h.update)
	if err != nil {

		return err
	}

	return h.convertReqToTuningSpec()
}

func (h *helper) convertReqToTuningSpec() error {
	var err error
	h.tuning.Name = h.c.Param("parameterName")
	h.tuning.Enabled, err = h.parseEnabled()
	if err != nil {
		return err
	}

	h.tuning.Value = h.update.Value
	h.tuning.SetHosts(h.update.Hosts)
	h.tuning.SetUpdating()
	return nil
}

func (h *helper) parseEnableValue() error {
	err := h.c.ShouldBindJSON(&h.toggle)
	if err != nil {
		return err
	}

	h.tuning.Name = h.c.Param("parameterName")
	if !h.isTuningModified(h.toggle.Hosts) {
		return errors.New("can't enable/disable unmodified tuning")
	}

	tuning, err := h.getTuningByNameAndHosts(h.tuning.Name, h.toggle.Hosts)
	if err != nil {
		log.Errorf("tunings(%s): failed to get %s: %v", h.reqId, h.tuning.Name, err)
		return err
	}

	h.tuning = *tuning
	h.tuning.Enabled = h.toggle.Enable
	h.tuning.SetHosts(h.toggle.Hosts)
	h.tuning.SetUpdating()
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
	h.tuning.Value = spec.Limitation.Default
	h.tuning.Enabled = true
	h.tuning.SetHosts(h.reset.Hosts)
	if !h.isTuningModified(h.reset.Hosts) {
		return errors.New("can't reset unmodified tuning")
	}

	h.tuning.IsModified = false
	h.tuning.SetResetting()
	return nil
}

func (h *helper) parseEnabled() (bool, error) {
	tuning, err := h.getTuningByNameAndHosts(h.tuning.Name, h.hosts)
	if err != nil {
		log.Errorf("tunings(%s): failed to get %s(%v)", h.reqId, h.tuning.Name, err)
		return false, err
	}

	if !tuning.IsModified {
		return true, nil
	}

	return tuning.Enabled, nil
}

func (h *helper) parsePage() error {
	var err error
	h.Page, err = queries.GetPage(h.c)
	return err
}

func (h *helper) parseScope() {
	h.allNodes = h.c.DefaultQuery("allNodes", "true") == "true"
}

func (h *helper) parseKeyword() {
	keyword := h.c.DefaultQuery("keyword", "")
	h.keyword = strings.ToLower(keyword)
}

func (h *helper) parseHosts() {
	h.hosts = h.c.QueryArray("host")
}

func (h *helper) parseModified() {
	modifies := h.c.QueryArray("modified")
	for _, m := range modifies {
		h.modified = append(
			h.modified,
			strings.ToLower(m) == "true",
		)
	}
}

func (h *helper) parseWatch() error {
	var err error
	h.watch, err = queries.GetWatch(h.c)
	return err
}

func (h *helper) isFilterRequired() bool {
	return h.isKeywordRequired() || h.isHostsRequired() || h.isModifiedRequired()
}

func (h *helper) isKeywordRequired() bool {
	return h.keyword != ""
}

func (h *helper) isHostsRequired() bool {
	return len(h.hosts) > 0
}

func (h *helper) isModifiedRequired() bool {
	_, required := h.c.GetQuery("modified")
	return required
}

func (h *helper) isTuningModified(hosts []string) bool {
	tuning, err := h.getTuningByNameAndHosts(h.tuning.Name, hosts)
	if err != nil {
		log.Errorf("tunings(%s): failed to get tuning: %v", h.reqId, err)
		return false
	}

	return tuning.IsModified
}
