package integrations

import (
	"testing"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"
	defstorages "github.com/bigstack-oss/cube-cos-api/internal/definition/v1/storages"
	"github.com/stretchr/testify/require"
	"k8s.io/client-go/util/workqueue"
)

func modelReq(driver, modelType string) defstorages.ModelReqOpts {
	return defstorages.ModelReqOpts{
		Model: defstorages.Model{Driver: driver, Type: modelType},
	}
}

// Replacing the model list updates the drivers that exist and creates the new
// ones. It must never mark an uploaded model for deletion.
func TestMarkBatchStorageModelReqOptsCreatesNewAndUpdatesExistingDrivers(t *testing.T) {
	currents := []defstorages.Model{
		{Driver: "cinder.volume.drivers.rt694_1659.ModelA", Type: "model-a-v2"},
		{Driver: "cinder.volume.drivers.rt694_1659.ModelC", Type: "model-c-v1"},
		{Driver: "cinder.volume.drivers.nfs.NfsDriver", Type: "nfs"},
	}

	tests := []struct {
		name     string
		uploaded []defstorages.ModelReqOpts
		want     map[string]string
	}{
		{
			name: "existing driver is updated",
			uploaded: []defstorages.ModelReqOpts{
				modelReq("cinder.volume.drivers.rt694_1659.ModelA", "model-a-v3"),
			},
			want: map[string]string{
				"cinder.volume.drivers.rt694_1659.ModelA": status.Updated,
			},
		},
		{
			name: "new driver is created",
			uploaded: []defstorages.ModelReqOpts{
				modelReq("cinder.volume.drivers.rt694_1659.ModelB", "model-b-v1"),
			},
			want: map[string]string{
				"cinder.volume.drivers.rt694_1659.ModelB": status.Created,
			},
		},
		{
			name: "mixed list updates the existing and creates the new",
			uploaded: []defstorages.ModelReqOpts{
				modelReq("cinder.volume.drivers.rt694_1659.ModelA", "model-a-v3"),
				modelReq("cinder.volume.drivers.rt694_1659.ModelB", "model-b-v1"),
			},
			want: map[string]string{
				"cinder.volume.drivers.rt694_1659.ModelA": status.Updated,
				"cinder.volume.drivers.rt694_1659.ModelB": status.Created,
			},
		},
		{
			name:     "empty upload issues nothing",
			uploaded: []defstorages.ModelReqOpts{},
			want:     map[string]string{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := markBatchStorageModelReqOpts(test.uploaded, currents, "req-1", "node1")

			require.Len(t, got, len(test.want))
			for _, req := range got {
				require.Equal(t, test.want[req.Driver], req.Status.Desired, req.Driver)
				require.NotEqual(t, status.Deleted, req.Status.Desired, req.Driver)
				require.True(t, req.Status.IsProcessing)
				require.Equal(t, "req-1", req.ReqId)
				require.Equal(t, "node1", req.Hostname)
			}
		})
	}
}

// Custom models left out of the upload are kept: the batch only carries the
// uploaded drivers, so nothing is issued for ModelC.
func TestMarkBatchStorageModelReqOptsKeepsModelsMissingFromTheUpload(t *testing.T) {
	currents := []defstorages.Model{
		{Driver: "cinder.volume.drivers.rt694_1659.ModelA"},
		{Driver: "cinder.volume.drivers.rt694_1659.ModelC"},
	}
	uploaded := []defstorages.ModelReqOpts{
		modelReq("cinder.volume.drivers.rt694_1659.ModelA", "model-a-v3"),
	}

	got := markBatchStorageModelReqOpts(uploaded, currents, "req-1", "node1")

	require.Len(t, got, 1)
	require.Equal(t, "cinder.volume.drivers.rt694_1659.ModelA", got[0].Driver)
	require.Equal(t, "model-a-v3", got[0].Type)
}

// Swaps in a fresh queue so the test sees only what it enqueued.
func stubModelReqQueue(t *testing.T) workqueue.TypedInterface[*defstorages.ModelReqOpts] {
	t.Helper()

	orig := modelReqQueue
	queue := workqueue.NewTyped[*defstorages.ModelReqOpts]()
	modelReqQueue = queue

	t.Cleanup(func() {
		queue.ShutDown()
		modelReqQueue = orig
	})

	return queue
}

// Every model of a replaced list reaches the operator as its own request,
// even though the batch reuses h.modelReqOpts for each one.
func TestEnqueueStorageModelReqQueuesEveryModelOfTheBatch(t *testing.T) {
	queue := stubModelReqQueue(t)

	batch := markBatchStorageModelReqOpts(
		[]defstorages.ModelReqOpts{
			modelReq("cinder.volume.drivers.rt694_1659.ModelA", "model-a-v3"),
			modelReq("cinder.volume.drivers.rt694_1659.ModelB", "model-b-v1"),
		},
		[]defstorages.Model{{Driver: "cinder.volume.drivers.rt694_1659.ModelA"}},
		"req-1",
		"node1",
	)

	h := &helper{}
	for _, reqOpts := range batch {
		h.modelReqOpts = reqOpts
		h.enqueueStorageModelReq()
	}

	require.Equal(t, 2, queue.Len())

	got := map[string]string{}
	for range 2 {
		req, shutdown := queue.Get()
		require.False(t, shutdown)
		got[req.Driver] = req.Type + "/" + req.Status.Desired
		queue.Done(req)
	}

	require.Equal(t, map[string]string{
		"cinder.volume.drivers.rt694_1659.ModelA": "model-a-v3/" + status.Updated,
		"cinder.volume.drivers.rt694_1659.ModelB": "model-b-v1/" + status.Created,
	}, got)
}
