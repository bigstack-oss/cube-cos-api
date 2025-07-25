package events

import (
	"fmt"

	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/queries"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/events"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/pages"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/time"
	"github.com/gin-gonic/gin"
	log "go-micro.dev/v5/logger"
)

type helper struct {
	c       *gin.Context
	reqId   string
	handler string

	eventType  string
	eventTypes []string
	eventId    string
	eventIds   []string
	category   string
	categories []string
	severity   string
	severities []string
	host       string
	hosts      []string
	instance   string
	instances  []string
	keyword    string

	period *time.Period
	past   string

	page  *pages.Page
	limit int

	watch bool
}

func initHelper(c *gin.Context, handler string) (*helper, error) {
	h := &helper{c: c, reqId: queries.GetReqId(c), handler: handler}
	return h, h.parseParamsByHandler()
}

func (h *helper) listEvents() (*data, error) {
	stmt := h.genListingStmt()
	events, err := cubecos.ListEvents(stmt)
	if err != nil {
		log.Errorf("events(%s): failed to get events: %v", h.reqId, err)
		return nil, err
	}

	filteredEvents := h.filteredByKeyword(events)
	pagedEvents, err := h.paginateEvents(filteredEvents)
	if err != nil {
		log.Errorf("tunings(%s): failed to paginate tunings: %v", h.reqId, err)
		return nil, err
	}

	page, err := h.genPageInfo(filteredEvents)
	if err != nil {
		log.Errorf("events(%s): failed to gen page info: %v", h.reqId, err)
		return nil, err
	}

	return &data{
		Events: pagedEvents,
		Page:   &page,
	}, nil
}

func (h *helper) listPredefinedEvents() ([]predefinedEvent, error) {
	events, err := cubecos.GetPredefinedEvents()
	if err != nil {
		return nil, err
	}

	predefinedEvents := []predefinedEvent{}
	for _, event := range events {
		predefinedEvents = append(predefinedEvents, predefinedEvent{
			Type:        event.Type,
			Id:          event.Id,
			Severity:    event.Severity,
			Category:    event.Category,
			Description: event.Message,
		})
	}

	return h.filteredPredefinedEvents(predefinedEvents), nil
}

func (h *helper) listEventAbstract() (*data, error) {
	stmt := h.genAbstractStmt()
	events, err := cubecos.ListEvents(stmt)
	if err != nil {
		log.Errorf("events(%s): failed to get events: %v", h.reqId, err)
		return nil, err
	}

	return &data{
		Events: events,
		Limit: &pages.Limit{
			Number:      h.limit,
			Description: fmt.Sprintf("the top %d recent events", h.limit),
		},
	}, nil
}

func (h *helper) getEventRank() (*data, error) {
	stmt, err := h.genRankStmt()
	if err != nil {
		log.Errorf("events(%s): %v", h.reqId, err)
		return nil, err
	}

	rank, err := cubecos.GetEventRank(stmt)
	if err != nil {
		log.Errorf("events(%s): failed to get events: %v", h.reqId, err)
		return nil, err
	}

	return &data{
		Events: rank,
		Limit: &pages.Limit{
			Number:      h.limit,
			Description: fmt.Sprintf("The top %d event IDs with the highest proportion", len(rank)),
		},
	}, nil
}

func (h *helper) getEventFilterConditions() *events.Filter {
	systemCategories, err := cubecos.GetEventFilterConditions(h.genFilterConditionStmt("system", "category"))
	if err != nil {
		log.Errorf("events(%s): failed to get system categories: %v", h.reqId, err)
	}

	systemSeverities, err := cubecos.GetEventFilterConditions(h.genFilterConditionStmt("system", "severity"))
	if err != nil {
		log.Errorf("events(%s): failed to get system severities: %v", h.reqId, err)
	}

	hostCategories, err := cubecos.GetEventFilterConditions(h.genFilterConditionStmt("host", "category"))
	if err != nil {
		log.Errorf("events(%s): failed to get host categories: %v", h.reqId, err)
	}

	hostnames, err := cubecos.GetEventFilterConditions(h.genFilterConditionStmt("host", "host"))
	if err != nil {
		log.Errorf("events(%s): failed to get host names: %v", h.reqId, err)
	}

	instanceCategories, err := cubecos.GetEventFilterConditions(h.genFilterConditionStmt("instance", "category"))
	if err != nil {
		log.Errorf("events(%s): failed to get instance categories: %v", h.reqId, err)
	}

	instanceIds, err := cubecos.GetEventFilterConditions(h.genFilterConditionStmt("instance", "instance"))
	if err != nil {
		log.Errorf("events(%s): failed to get instances: %v", h.reqId, err)
	}

	return &events.Filter{
		System: events.SystemFilter{
			Categories: systemCategories,
			Severities: convertSystemSeverities(systemSeverities),
		},
		Host: events.HostFilter{
			Categories: hostCategories,
			Names:      hostnames,
		},
		Instance: events.InstanceFilter{
			Categories: instanceCategories,
			Ids:        instanceIds,
		},
	}
}
