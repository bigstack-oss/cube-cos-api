package cubecos

import (
	"context"
	"encoding/json"
	"time"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/influx"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/api/query"
	log "go-micro.dev/v5/logger"
)

const (
	eventTimeLayout = "2006-01-02 15:04:05.999999999 -0700 MST"
)

var (
	isValidEventMeasurement = map[string]bool{
		"system":   true,
		"host":     true,
		"instance": true,
	}
)

func IsEventTypeValid(t string) bool {
	return isValidEventMeasurement[t]
}

func CountEvents(stmt string) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	h := influx.GetGlobalHelper()
	c, err := h.QueryApiClient.Query(ctx, stmt)
	if err != nil {
		log.Errorf("failed to get query cursor: %v", err)
		return 0, err
	}

	defer c.Close()
	return countEvents(c)
}

func countEvents(c *api.QueryTableResult) (int64, error) {
	count := int64(0)

	for c.Next() {
		record := c.Record()
		rawCount := record.Value().(int64)
		count = count + rawCount
	}
	if c.Err() != nil {
		return 0, c.Err()
	}

	return count, nil
}

func GetEvents(stmt string) ([]definition.Event, error) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(60))
	defer cancel()

	h := influx.GetGlobalHelper()
	c, err := h.QueryApiClient.Query(ctx, stmt)
	if err != nil {
		log.Errorf("failed to get query cursor: %v", err)
		return nil, err
	}

	defer c.Close()
	events := []definition.Event{}
	err = parseEvents(c, &events)
	if err != nil {
		log.Errorf("failed to parse events from cursor: %v", err)
		return nil, err
	}

	return events, nil
}

func parseEvents(c *api.QueryTableResult, events *[]definition.Event) error {
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
	date, err := time.Parse(eventTimeLayout, record.Time().Local().String())
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
		Type:        record.Measurement(),
		Severity:    definition.SeverityFullName(severity),
		Id:          eventId,
		Description: msg,
		Host:        "",
		Time:        definition.TimeLocalISO8601(date),
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
		return
	}

	event.Category, _ = metaCategory.(string)
}

func setServiceToEvent(event *definition.Event, metaObj map[string]interface{}) {
	metaService, found := metaObj["service"]
	if !found {
		return
	}

	event.Service, _ = metaService.(string)
}

func setHostnameToEvent(event *definition.Event, metaObj map[string]interface{}) {
	metaHost, found := metaObj["host"]
	if !found {
		return
	}

	event.Host, _ = metaHost.(string)
}
