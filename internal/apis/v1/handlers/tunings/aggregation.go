package tunings

import (
	"slices"

	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	bstuning "github.com/bigstack-oss/cube-cos-api/internal/definition/v1/tunings"
)

func (h *helper) listAggregatedTunings() ([]bstuning.Tuning, error) {
	all := bstuning.ListOptions{AllNodes: h.allNodes}
	hostTunings, err := cubecos.ListTunings(all)
	if err != nil {
		return nil, err
	}

	for host, tunings := range hostTunings {
		if len(tunings) == 0 {
			continue
		}

		for i, tuning := range tunings {
			tunings[i] = h.getUpdatingTuning(tuning, host)
		}

		hostTunings[host] = tunings
	}

	tunings := h.convergeTunings(hostTunings)
	h.sortTunings(&tunings)
	return tunings, nil
}

func (h *helper) convergeTunings(nodeTunings map[string][]bstuning.Tuning) []bstuning.Tuning {
	mergedMap := make(map[string]bstuning.Tuning)
	for _, tunings := range nodeTunings {
		h.setTunings(mergedMap, tunings)
	}

	tunings := []bstuning.Tuning{}
	for _, item := range mergedMap {
		tunings = append(tunings, item)
	}

	return tunings
}

func (h *helper) setTunings(mergedMap map[string]bstuning.Tuning, tunings []bstuning.Tuning) {
	for _, tuning := range tunings {
		key := tuning.IndexKey()
		existing, found := mergedMap[key]
		if !found {
			mergedMap[key] = tuning
			continue
		}

		if tuning.Status.IsUpdating {
			existing.Status = tuning.Status
		}

		existing.Hosts = slices.Concat(existing.Hosts, tuning.Hosts)
		mergedMap[key] = existing
	}
}
