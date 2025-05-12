package cubecos

import (
	"context"
	ostime "time"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/influx"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/math"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/events"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/time"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/api/query"
	json "github.com/json-iterator/go"
	log "go-micro.dev/v5/logger"
)

func CountEvents(stmt string) (int64, error) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(60))
	defer cancel()

	h := influx.GetGlobalHelper()
	c, err := h.QueryApiClient.Query(ctx, stmt)
	if err != nil {
		log.Errorf("events: failed to get query cursor: %v", err)
		return 0, err
	}

	defer c.Close()
	return countEvents(c)
}

func countEvents(c *api.QueryTableResult) (int64, error) {
	count := int64(0)
	for c.Next() {
		count++
	}
	if c.Err() != nil {
		return 0, c.Err()
	}

	return count, nil
}

func ListEvents(stmt string) ([]events.Event, error) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(60))
	defer cancel()

	h := influx.GetGlobalHelper()
	c, err := h.QueryApiClient.Query(ctx, stmt)
	if err != nil {
		log.Errorf("events: failed to get query cursor: %v", err)
		return nil, err
	}

	defer c.Close()
	events := []events.Event{}
	err = parseEvents(c, &events)
	if err != nil {
		log.Errorf("events: failed to parse events from cursor: %v", err)
		return nil, err
	}

	return events, nil
}

func GetEventRank(stmt string) ([]events.Stat, error) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(60))
	defer cancel()

	h := influx.GetGlobalHelper()
	c, err := h.QueryApiClient.Query(ctx, stmt)
	if err != nil {
		log.Errorf("events: failed to get query cursor: %v", err)
		return nil, err
	}

	defer c.Close()
	events := []events.Stat{}
	err = parseEventStats(c, &events)
	if err != nil {
		log.Errorf("events: failed to parse events from cursor: %v", err)
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
		log.Errorf("events: failed to get cursor of event filter condition: %v", err)
		return nil, err
	}

	defer c.Close()
	values := []string{}
	err = parseEventValues(c, &values)
	if err != nil {
		log.Errorf("events: failed to parse filter condition from cursor: %v", err)
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

func setPercentageToEachEvent(events *[]events.Stat) {
	total := int64(0)
	for _, event := range *events {
		total = total + event.Number
	}

	for i := range *events {
		percent := float64((*events)[i].Number) / float64(total) * 100
		(*events)[i].Percent = math.RoundDown(percent, 4)
	}
}

func parseEventStats(c *api.QueryTableResult, events *[]events.Stat) error {
	for c.Next() {
		event := genEventStatsByRecord(c.Record())
		*events = append(*events, event)
	}
	if c.Err() != nil {
		return c.Err()
	}

	return nil
}

func genEventStatsByRecord(record *query.FluxRecord) events.Stat {
	return events.Stat{
		Id:           record.ValueByKey("key").(string),
		Number:       record.ValueByKey("number").(int64),
		Severity:     parseSeverity(record),
		Category:     parseCategory(record),
		Host:         parseHost(record),
		InstanceId:   parseInstanceId(record),
		InstanceName: parseInstanceName(record),
	}
}

func parseEvents(c *api.QueryTableResult, events *[]events.Event) error {
	for c.Next() {
		record := c.Record()
		event := parseEvent(record)
		setMetadata(&event, record)
		*events = append(*events, event)
	}
	if c.Err() != nil {
		return c.Err()
	}

	return nil
}

func parseEvent(record *query.FluxRecord) events.Event {
	date, err := ostime.Parse(events.TimeLayout, record.Time().Local().String())
	if err != nil {
		log.Debugf("events: failed to parse date from record: %v", record)
	}

	severity, ok := record.ValueByKey("severity").(string)
	if !ok {
		log.Debugf("events: failed to parse severity from record: %v", record)
	}

	eventId, ok := record.ValueByKey("key").(string)
	if !ok {
		log.Debugf("events: failed to parse key from record: %v", record)
	}

	msg, ok := record.ValueByKey("message").(string)
	if !ok {
		log.Debugf("events: failed to parse message from record: %v", record)
	}

	host, ok := record.ValueByKey("host").(string)
	if !ok {
		log.Debugf("events: failed to parse host from record: %v", record)
	}

	return events.Event{
		Type:        record.Measurement(),
		Severity:    events.GetSeverityFullName(severity),
		Id:          eventId,
		Description: msg,
		Host:        host,
		Time:        time.RFC3339Z(date),
	}
}

func setMetadata(event *events.Event, record *query.FluxRecord) {
	metadata, ok := record.ValueByKey("metadata").(string)
	if !ok {
		log.Debugf("events: failed to parse metadata from record: %v", record)
		return
	}

	metaObj := map[string]any{}
	err := json.Unmarshal([]byte(metadata), &metaObj)
	if err != nil {
		log.Debugf("events: failed to parse metadata from record: %v", record)
		return
	}

	event.Metadata = metaObj
	event.SetCategory(metaObj)
	event.SetService(metaObj)
	event.SetHostname(metaObj)
}

func parseSeverity(record *query.FluxRecord) string {
	severity, ok := record.ValueByKey("severity").(string)
	if !ok {
		log.Debugf("events: failed to parse severity from record: %v", record)
		return ""
	}

	return events.GetSeverityFullName(severity)
}

func parseCategory(record *query.FluxRecord) string {
	category, ok := record.ValueByKey("category").(string)
	if !ok {
		log.Debugf("events: failed to parse category from record: %v", record)
		return ""
	}

	return category
}

func parseInstanceId(record *query.FluxRecord) string {
	id, ok := record.ValueByKey("instance").(string)
	if !ok {
		log.Debugf("events: failed to parse instance from record: %v", record)
		return ""
	}

	return id
}

func parseInstanceName(record *query.FluxRecord) string {
	name, ok := record.ValueByKey("vm_name").(string)
	if !ok {
		log.Debugf("events: failed to parse instance name from record: %v", record)
		return ""
	}

	return name
}
