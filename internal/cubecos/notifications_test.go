package cubecos

import (
	"testing"
	"time"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/notifications"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// The same instant, 2026-10-09T09:04:31Z, as producers on differently zoned
// nodes store it.
var sameInstantTimes = []string{
	"2026-10-09T09:04:31Z",
	"2026-10-09T09:04:31+00:00",
	"2026-10-09T17:04:31+08:00",
	"2026-10-09T18:04:31+09:00",
}

func stageValue(t *testing.T, stage bson.D, op string) any {
	t.Helper()

	require.Len(t, stage, 1)
	require.Equal(t, op, stage[0].Key)
	return stage[0].Value
}

func requireTimeAtStage(t *testing.T, stage bson.D) {
	t.Helper()

	fields := stageValue(t, stage, "$addFields").(bson.D)
	require.Equal(t, timeAtField, fields[0].Key)

	parse := fields[0].Value.(bson.D)
	require.Equal(t, "$dateFromString", parse[0].Key)
	require.Equal(t, bson.D{
		{Key: "dateString", Value: "$time"},
		{Key: "onError", Value: nil},
		{Key: "onNull", Value: nil},
	}, parse[0].Value)
}

func requireDropsTimeAt(t *testing.T, stage bson.D) {
	t.Helper()

	require.Equal(t, bson.D{{Key: timeAtField, Value: 0}}, stageValue(t, stage, "$project"))
}

func TestGenListPipelineFiltersAndSortsByInstantNotByString(t *testing.T) {
	for _, offsetSeconds := range []int{0, 8 * 60 * 60, 9 * 60 * 60, -5 * 60 * 60} {
		zone := time.FixedZone("", offsetSeconds)
		start := time.Date(2026, 10, 9, 8, 0, 0, 0, time.UTC).In(zone)
		stop := time.Date(2026, 10, 9, 10, 0, 0, 0, time.UTC).In(zone)

		pipeline := genListPipeline(notifications.ListOpts{
			Limit:     4,
			Desending: true,
			Start:     start,
			Stop:      stop,
		})

		require.Len(t, pipeline, 5)
		requireTimeAtStage(t, pipeline[0])

		// The bounds stay time.Time, so they are sent as BSON dates and
		// compared with the parsed instant, whatever offset they carry.
		match := stageValue(t, pipeline[1], "$match").(bson.D)
		require.Equal(t, timeAtField, match[0].Key)
		bounds := match[0].Value.(bson.D)
		require.Equal(t, "$gte", bounds[0].Key)
		require.IsType(t, time.Time{}, bounds[0].Value)
		require.True(t, bounds[0].Value.(time.Time).Equal(start))
		require.Equal(t, "$lte", bounds[1].Key)
		require.IsType(t, time.Time{}, bounds[1].Value)
		require.True(t, bounds[1].Value.(time.Time).Equal(stop))

		require.Equal(t, bson.D{{Key: timeAtField, Value: -1}}, stageValue(t, pipeline[2], "$sort"))
		require.Equal(t, int64(4), stageValue(t, pipeline[3], "$limit"))
		requireDropsTimeAt(t, pipeline[4])
	}
}

func TestGenListPipelineKeepsStoredOrderWhenNotDescending(t *testing.T) {
	pipeline := genListPipeline(notifications.ListOpts{Limit: 4, Start: time.Now(), Stop: time.Now()})

	require.Len(t, pipeline, 4)
	stageValue(t, pipeline[1], "$match")
	stageValue(t, pipeline[2], "$limit")
}

func TestGenLastPipelineSortsByInstant(t *testing.T) {
	pipeline := genLastPipeline()

	require.Equal(t, mongo.Pipeline{
		pipeline[0],
		{{Key: "$sort", Value: bson.D{{Key: timeAtField, Value: -1}}}},
		{{Key: "$limit", Value: 1}},
		pipeline[3],
	}, pipeline)
	requireTimeAtStage(t, pipeline[0])
	requireDropsTimeAt(t, pipeline[3])
}
