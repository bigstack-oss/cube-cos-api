package cubecos

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/notifications"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Runs the notification pipelines on a real MongoDB, because $dateFromString
// is evaluated by the server. Skipped unless a URI is given, for example:
//
//	docker run -d --rm --name mongo-test -p 27017:27017 mongo:8.0.29
//	CUBE_COS_API_TEST_MONGO_URI=mongodb://localhost:27017 GOWORK=off \
//	    go test ./internal/cubecos/ -run OnMongo
const testMongoURIEnv = "CUBE_COS_API_TEST_MONGO_URI"

func newTestToasts(t *testing.T, times map[string]string) *mongo.Collection {
	t.Helper()

	uri := os.Getenv(testMongoURIEnv)
	if uri == "" {
		t.Skipf("%s is not set", testMongoURIEnv)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	require.NoError(t, err)
	t.Cleanup(func() { _ = client.Disconnect(context.Background()) })

	coll := client.Database("cube_cos_api_test").Collection(fmt.Sprintf("toasts_%d", time.Now().UnixNano()))
	t.Cleanup(func() { _ = coll.Drop(context.Background()) })

	docs := []any{}
	for id, at := range times {
		docs = append(docs, notifications.Notification{Id: id, NodeName: "node", Time: at})
	}
	_, err = coll.InsertMany(ctx, docs)
	require.NoError(t, err)

	return coll
}

func aggregateIds(t *testing.T, coll *mongo.Collection, pipeline mongo.Pipeline) []string {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	c, err := coll.Aggregate(ctx, pipeline)
	require.NoError(t, err)

	list, err := parseNotifications(c)
	require.NoError(t, err)

	ids := []string{}
	for _, n := range list {
		ids = append(ids, n.Id)
	}
	return ids
}

func utcAt(value string) time.Time {
	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		panic(err)
	}
	return t
}

// Stored offsets are mixed, as on a node whose zone changed: string order is
// not instant order ("...T18:04:31+09:00" is 09:04:31Z, and
// "2026-10-10T00:03:00+08:00" is 16:03:00Z on the 9th).
var mixedOffsetTimes = map[string]string{
	"osd3":     "2026-10-09T16:04:07Z",
	"osd2":     "2026-10-09T16:03:41+00:00",
	"taipei":   "2026-10-10T00:03:00+08:00",
	"tokyo":    "2026-10-09T18:04:31+09:00",
	"early":    "2026-10-09T08:00:00.500Z",
	"newest":   "2026-10-09T20:00:00-05:00",
	"garbage":  "not a time",
	"no-time":  "",
	"too-late": "2026-10-09T17:00:01Z",
}

func TestListPipelineOnMongoFiltersAndSortsByInstant(t *testing.T) {
	coll := newTestToasts(t, mixedOffsetTimes)

	opts := notifications.ListOpts{
		Limit:     100,
		Desending: true,
		Start:     utcAt("2026-10-09T09:00:00Z"),
		Stop:      utcAt("2026-10-09T17:00:00+00:00"),
	}
	require.Equal(t, []string{"osd3", "osd2", "taipei", "tokyo"}, aggregateIds(t, coll, genListPipeline(opts)))

	opts.Limit = 2
	require.Equal(t, []string{"osd3", "osd2"}, aggregateIds(t, coll, genListPipeline(opts)))

	// The toast poll's next start, 1 s after osd.2, sent with a browser's
	// +08:00 offset: osd.3 must still be returned.
	opts.Limit = 4
	opts.Start = utcAt("2026-10-10T00:03:42+08:00")
	require.Equal(t, []string{"osd3"}, aggregateIds(t, coll, genListPipeline(opts)))
}

func TestListPipelineOnMongoMatchesOneInstantInEveryOffset(t *testing.T) {
	times := map[string]string{}
	for i, at := range sameInstantTimes {
		times[fmt.Sprintf("n%d", i)] = at
	}
	coll := newTestToasts(t, times)

	at := utcAt("2026-10-09T09:04:31Z")
	ids := aggregateIds(t, coll, genListPipeline(notifications.ListOpts{Limit: 10, Desending: true, Start: at, Stop: at}))
	require.ElementsMatch(t, []string{"n0", "n1", "n2", "n3"}, ids)
}

func TestLastPipelineOnMongoReturnsTheLatestInstant(t *testing.T) {
	coll := newTestToasts(t, mixedOffsetTimes)

	// By string, "2026-10-10T00:03:00+08:00" would win.
	require.Equal(t, []string{"newest"}, aggregateIds(t, coll, genLastPipeline()))
}

func TestPipelinesOnMongoDropTheParsedTime(t *testing.T) {
	coll := newTestToasts(t, map[string]string{"one": "2026-10-09T16:04:07Z"})

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	c, err := coll.Aggregate(ctx, genLastPipeline())
	require.NoError(t, err)

	docs := []bson.M{}
	require.NoError(t, c.All(ctx, &docs))
	require.Len(t, docs, 1)
	require.NotContains(t, docs[0], timeAtField)
	require.Equal(t, "2026-10-09T16:04:07Z", docs[0]["time"])
}
