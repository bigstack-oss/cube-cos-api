package cubecos

import (
	"context"
	"encoding/json"
	"time"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/influx"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/math"
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

func GetEventRank(stmt string) ([]definition.EventStat, error) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(60))
	defer cancel()

	h := influx.GetGlobalHelper()
	c, err := h.QueryApiClient.Query(ctx, stmt)
	if err != nil {
		log.Errorf("failed to get query cursor: %v", err)
		return nil, err
	}

	defer c.Close()
	events := []definition.EventStat{}
	err = parseEventStats(c, &events)
	if err != nil {
		log.Errorf("failed to parse events from cursor: %v", err)
		return nil, err
	}

	setPercentageToEachEvent(&events)
	return events, nil
}

func GetEventFilterConditions(stmt string) ([]string, error) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(60))
	defer cancel()

	h := influx.GetGlobalHelper()
	c, err := h.QueryApiClient.Query(ctx, stmt)
	if err != nil {
		log.Errorf("failed to get cursor of event filter condition: %v", err)
		return nil, err
	}

	defer c.Close()
	values := []string{}
	err = parseEventValues(c, &values)
	if err != nil {
		log.Errorf("failed to parse filter condition from cursor: %v", err)
		return nil, err
	}

	return values, nil
}

func parseEventValues(c *api.QueryTableResult, values *[]string) error {
	for c.Next() {
		*values = append(
			*values,
			c.Record().Value().(string),
		)
	}
	if c.Err() != nil {
		return c.Err()
	}

	return nil
}

func setPercentageToEachEvent(events *[]definition.EventStat) {
	total := int64(0)
	for _, event := range *events {
		total = total + event.Number
	}

	for i := range *events {
		percent := float64((*events)[i].Number) / float64(total) * 100
		(*events)[i].Percent = math.RoundDown(percent, 4)
	}
}

func parseEventStats(c *api.QueryTableResult, events *[]definition.EventStat) error {
	for c.Next() {
		event := genEventStatsByRecord(c.Record())
		*events = append(*events, event)
	}
	if c.Err() != nil {
		return c.Err()
	}

	return nil
}

func genEventStatsByRecord(record *query.FluxRecord) definition.EventStat {
	return definition.EventStat{
		Id:     record.ValueByKey("key").(string),
		Number: record.ValueByKey("number").(int64),
	}
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
		log.Debugf("failed to parse date from record: %v", record)
	}

	severity, ok := record.ValueByKey("severity").(string)
	if !ok {
		log.Debugf("failed to parse severity from record: %v", record)
	}

	eventId, ok := record.ValueByKey("key").(string)
	if !ok {
		log.Debugf("failed to parse key from record: %v", record)
	}

	msg, ok := record.ValueByKey("message").(string)
	if !ok {
		log.Debugf("failed to parse message from record: %v", record)
	}

	host, ok := record.ValueByKey("host").(string)
	if !ok {
		log.Debugf("failed to parse host from record: %v", record)
	}

	return definition.Event{
		Type:        record.Measurement(),
		Severity:    definition.SeverityFullName(severity),
		Id:          eventId,
		Description: msg,
		Host:        host,
		Time:        definition.TimeISO8601Z(date),
	}
}

func setMetadataToEvent(event *definition.Event, record *query.FluxRecord) {
	metadata, ok := record.ValueByKey("metadata").(string)
	if !ok {
		log.Debugf("failed to parse metadata from record: %v", record)
		return
	}

	metaObj := map[string]interface{}{}
	err := json.Unmarshal([]byte(metadata), &metaObj)
	if err != nil {
		log.Debugf("failed to parse metadata from record: %v", record)
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
