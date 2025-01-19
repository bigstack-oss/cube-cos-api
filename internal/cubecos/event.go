package cubecos

import (
	"context"
	"encoding/json"
	"time"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/influx"
	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/api/query"
	log "go-micro.dev/v5/logger"
)

const (
	eventTimeLayout = "2006-01-02 15:04:05.999999999 -0700 MST"
)

func ListEvents(stmt string) ([]definition.Event, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	h := influx.GetGlobalHelper()
	c, err := h.QueryApiClient.Query(ctx, stmt)
	if err != nil {
		log.Errorf("failed to get query cursor: %v", err)
		return nil, err
	}

	defer c.Close()
	events := []definition.Event{}
	err = parseEventsFromCursor(c, &events)
	if err != nil {
		log.Errorf("failed to parse events from cursor: %v", err)
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
	date, err := time.Parse(eventTimeLayout, record.Time().String())
	if err != nil {
		log.Warnf("failed to parse date from record: %v", record)
	}

	severity, ok := record.ValueByKey("severity").(string)
	if !ok {
		log.Warnf("failed to parse severity from record: %v", record)
	}

	eventId, ok := record.ValueByKey("key").(string)
	if !ok {
		log.Warnf("failed to parse key from record: %v", record)
	}

	msg, ok := record.ValueByKey("message").(string)
	if !ok {
		log.Warnf("failed to parse message from record: %v", record)
	}

	return definition.Event{
		Type:        definition.SeverityFullName(severity),
		ID:          eventId,
		Description: msg,
		Host:        "",
		Time:        date.Format(time.RFC3339),
	}
}

func setMetadataToEvent(event *definition.Event, record *query.FluxRecord) {
	metadata, ok := record.ValueByKey("metadata").(string)
	if !ok {
		log.Warnf("failed to parse metadata from record: %v", record)
		return
	}

	metaObj := map[string]interface{}{}
	err := json.Unmarshal([]byte(metadata), &metaObj)
	if err != nil {
		log.Warnf("failed to parse metadata from record: %v", record)
		return
	}

	event.Metadata = metaObj
	setCategoryToEvent(event, metaObj)
	setServiceToEvent(event, metaObj)
	setHostnameToEvent(event, metaObj)
}

func setCategoryToEvent(event *definition.Event, metaObj map[string]interface{}) {
	metaCategory, found := metaObj["category"]
	if !found {
		log.Warnf("category not found in the metadata")
		return
	}

	ok := false
	event.Category, ok = metaCategory.(string)
	if !ok {
		log.Warnf("failed to parse category")
	}
}

func setServiceToEvent(event *definition.Event, metaObj map[string]interface{}) {
	metaService, found := metaObj["service"]
	if !found {
		log.Warnf("service not found in the metadata")
		return
	}

	ok := false
	event.Service, ok = metaService.(string)
	if !ok {
		log.Warnf("failed to parse service")
	}
}

func setHostnameToEvent(event *definition.Event, metaObj map[string]interface{}) {
	metaHost, found := metaObj["host"]
	if !found {
		log.Warnf("hostname not found in the metadata")
		return
	}

	ok := false
	event.Host, ok = metaHost.(string)
	if !ok {
		log.Warnf("failed to parse host")
	}
}
