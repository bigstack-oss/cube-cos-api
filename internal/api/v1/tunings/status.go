package tunings

import definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"

func setBatchPendingDeletion(tunings []definition.Tuning) {
	for i := range tunings {
		tunings[i].Status.SetDesiredToDelete()
		tunings[i].Status.SetCurrentToPending()
	}
}

func setBatchPendingUpdate(tunings []definition.Tuning) {
	for i := range tunings {
		tunings[i].Status.SetDesiredToUpdate()
		tunings[i].Status.SetCurrentToPending()
	}
}
