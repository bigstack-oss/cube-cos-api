package cubecos

import (
	"encoding/json"
	"time"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/influx"
	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/api/query"
	log "go-micro.dev/v5/logger"
)

func ListEvents(stmt string) ([]definition.Event, error) {
	h := influx.GetGlobalHelper()
	c, err := h.GetQueryCursor(stmt)
	if err != nil {
		return nil, err
	}

	defer c.Close()
	events := []definition.Event{}
	err = parseEventsFromCursor(c, &events)
	if err != nil {
		return nil, err
	}

	return events, nil
}

func parseEventsFromCursor(c *api.QueryTableResult, events *[]definition.Event) error {
	for c.Next() {
		record := c.Record()
		event := genEventByRecord(record)
		setMetadataToEvent(&event, record)
		*events = append(*events, event)
	}
	if c.Err() != nil {
		return c.Err()
	}

	return nil
}

func genEventByRecord(record *query.FluxRecord) definition.Event {
	date, err := time.Parse(definition.Iso8601, record.Time().String())
	if err != nil {
		log.Errorf("failed to parse date from record: %v", record)
	}

	severity, ok := record.ValueByKey("severity").(string)
	if !ok {
		log.Errorf("failed to parse severity from record: %v", record)
	}

	eventId, ok := record.ValueByKey("key").(string)
	if !ok {
		log.Errorf("failed to parse key from record: %v", record)
	}

	msg, ok := record.ValueByKey("message").(string)
	if !ok {
		log.Errorf("failed to parse message from record: %v", record)
	}

	return definition.Event{
		Type:        definition.SeverityFullName(severity),
		ID:          eventId,
		Description: msg,
		Host:        "",
		Time:        date.String(),
	}
}

func setMetadataToEvent(event *definition.Event, record *query.FluxRecord) {
	metadata, ok := record.ValueByKey("metadata").(string)
	if !ok {
		log.Errorf("failed to parse metadata from record: %v", record)
		return
	}

	metaObj := map[string]interface{}{}
	err := json.Unmarshal([]byte(metadata), &metaObj)
	if err != nil {
		log.Errorf("failed to parse metadata from record: %v", record)
		return
	}

	setCategoryToEvent(event, metaObj)
	setServiceToEvent(event, metaObj)
}

func setCategoryToEvent(event *definition.Event, metaObj map[string]interface{}) {
	metaCategory, found := metaObj["category"]
	if !found {
		log.Errorf("failed to get category from metadata")
		return
	}

	ok := false
	event.Category, ok = metaCategory.(string)
	if !ok {
		log.Errorf("failed to parse category")
	}
}

func setServiceToEvent(event *definition.Event, metaObj map[string]interface{}) {
	metaService, found := metaObj["service"]
	if !found {
		log.Errorf("failed to get service from metadata")
		return
	}

	ok := false
	event.Service, ok = metaService.(string)
	if !ok {
		log.Errorf("failed to parse service")
	}
}
