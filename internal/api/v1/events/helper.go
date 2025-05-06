package events

import (
	"fmt"

	"github.com/bigstack-oss/cube-cos-api/internal/api"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/event"
	"github.com/gin-gonic/gin"
	log "go-micro.dev/v5/logger"
)

type helper struct {
	c       *gin.Context
	handler string

	eventType  string
	eventId    string
	category   string
	categories []string
	severity   string
	severities []string
	host       string
	hosts      []string
	instance   string
	instances  []string
	keyword    string

	period *v1.Period
	past   string

	page  *v1.Page
	limit int

	watch bool
}

func initHelper(c *gin.Context, handler string) (*helper, error) {
	h := &helper{c: c, handler: handler}
	return h, h.parseParamsByHandler()
}

func (h *helper) listEvents() (*data, error) {
	stmt := h.genListingStmt()
	events, err := cubecos.ListEvents(stmt)
	if err != nil {
		log.Errorf("events(%s): failed to get events: %v", api.GetReqId(h.c), err)
		return nil, err
	}

	filteredEvents := h.filteredByKeyword(events)
	pagedEvents, err := h.paginateEvents(filteredEvents)
	if err != nil {
		log.Errorf("tunings(%s): failed to paginate tunings: %s", api.GetReqId(h.c), err.Error())
		return nil, err
	}

	page, err := h.genPageInfo(filteredEvents)
	if err != nil {
		log.Errorf("events(%s): failed to gen page info: %v", api.GetReqId(h.c), err)
		return nil, err
	}

	return &data{
		Events: pagedEvents,
		Page:   &page,
	}, nil
}

func (h *helper) listEventAbstract() (*data, error) {
	stmt := h.genAbstractStmt()
	events, err := cubecos.ListEvents(stmt)
	if err != nil {
		log.Errorf("events(%s): failed to get events: %v", api.GetReqId(h.c), err)
		return nil, err
	}

	return &data{
		Events: events,
		Limit: &v1.Limit{
			Number:      h.limit,
			Description: fmt.Sprintf("the top %d recent events", h.limit),
		},
	}, nil
}

func (h *helper) getEventRank() (*data, error) {
	stmt, err := h.genRankStmt()
	if err != nil {
		log.Errorf("events(%s): %v", api.GetReqId(h.c), err)
		return nil, err
	}

	rank, err := cubecos.GetEventRank(stmt)
	if err != nil {
		log.Errorf("events(%s): failed to get events: %v", api.GetReqId(h.c), err)
		return nil, err
	}

	return &data{
		Events: rank,
		Limit: &v1.Limit{
			Number:      h.limit,
			Description: fmt.Sprintf("The top %d event IDs with the highest proportion", len(rank)),
		},
	}, nil
}

func (h *helper) getEventFilterConditions() *event.Filter {
	systemCategories, err := cubecos.GetEventFilterConditions(h.genFilterConditionStmt("system", "category"))
	if err != nil {
		log.Errorf("events(%s): failed to get system categories: %v", api.GetReqId(h.c), err)
	}

	systemSeverities, err := cubecos.GetEventFilterConditions(h.genFilterConditionStmt("system", "severity"))
	if err != nil {
		log.Errorf("events(%s): failed to get system severities: %v", api.GetReqId(h.c), err)
	}

	hostCategories, err := cubecos.GetEventFilterConditions(h.genFilterConditionStmt("host", "category"))
	if err != nil {
		log.Errorf("events(%s): failed to get host categories: %v", api.GetReqId(h.c), err)
	}

	hostnames, err := cubecos.GetEventFilterConditions(h.genFilterConditionStmt("host", "host"))
	if err != nil {
		log.Errorf("events(%s): failed to get host names: %v", api.GetReqId(h.c), err)
	}

	instanceCategories, err := cubecos.GetEventFilterConditions(h.genFilterConditionStmt("instance", "category"))
	if err != nil {
		log.Errorf("events(%s): failed to get instance categories: %v", api.GetReqId(h.c), err)
	}

	instanceIds, err := cubecos.GetEventFilterConditions(h.genFilterConditionStmt("instance", "instance"))
	if err != nil {
		log.Errorf("events(%s): failed to get instances: %v", api.GetReqId(h.c), err)
	}

	return &event.Filter{
		System: event.SystemFilter{
			Categories: systemCategories,
			Severities: convertSystemSeverities(systemSeverities),
		},
		Host: event.HostFilter{
			Categories: hostCategories,
			Names:      hostnames,
		},
		Instance: event.InstanceFilter{
			Categories: instanceCategories,
			Ids:        instanceIds,
		},
	}
}
