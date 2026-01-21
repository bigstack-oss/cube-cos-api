package fixpacks

import (
	"fmt"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/fixpacks"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"
	log "go-micro.dev/v5/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (h *helper) setPkgAs(status string) error {
	return h.mongo.UpdateOne(
		fixpacks.Db,
		fixpacks.UploadCollection,
		bson.M{},
		bson.M{"$set": bson.M{"status": status}},
		options.Update().SetUpsert(true),
	)
}

func (h *helper) checkIfHasProcessingPkg() error {
	err := h.checkPkgBy(status.Uploading)
	if err != nil {
		return err
	}

	err = h.checkPkgBy(status.Verifying)
	if err != nil {
		return err
	}

	return nil
}

func (h *helper) checkPkgBy(status string) error {
	count, err := h.mongo.GetCount(
		fixpacks.Db,
		fixpacks.UploadCollection,
		bson.M{"status": status},
	)
	if err != nil {
		log.Errorf("fixpacks(%s): failed to check %s status(%v)", h.reqId, status, err)
		return fmt.Errorf("failed to check %s status", status)
	}

	if count > 0 {
		return fmt.Errorf(
			"there is a fixpack in %s status, please try again later",
			status,
		)
	}

	return nil
}

func (h *helper) clearPkgBy(status string) error {
	err := h.mongo.DeleteAll(
		fixpacks.Db,
		fixpacks.UploadCollection,
		bson.M{"status": status},
	)
	if err != nil {
		log.Errorf("fixpacks(%s): failed to clear %s status(%v)", h.reqId, status, err)
		return err
	}

	return nil
}

func (h *helper) syncRequestingRecord(list *[]fixpacks.Fixpack) {
	for i, fixpack := range *list {
		filter := bson.M{"version": fixpack.Version, "status.current": status.Installing}
		if h.hasInprogressUpdate(filter) {
			(*list)[i].Status.Current = status.Installing
			(*list)[i].Status.IsProcessing = true
			continue
		}

		filter["status.current"] = status.RollingBack
		if h.hasInprogressUpdate(filter) {
			(*list)[i].Status.Current = status.RollingBack
			(*list)[i].Status.IsProcessing = true
		}
	}
}

func (h *helper) syncStatusByNodeProgresses(list *[]fixpacks.Fixpack) {
	for _, fixpack := range *list {
		update, err := h.getFixpackUpdateProgress(fixpack.Version)
		if err != nil {
			log.Errorf("fixpacks(%s): failed to get fixpack update progress (%v)", h.reqId, err)
			return
		}

		finalStatus := status.Available
		val, found := h.foundFailureOrInprogessStatus(update.Progresses)
		if found {
			finalStatus = val
		}

		h.polishRebootingStatus(&fixpack, finalStatus, update.Operation)
		h.polishInstallationStatus(&fixpack, update)
	}
}

func (h *helper) foundFailureOrInprogessStatus(progresses []progress) (string, bool) {
	for _, progress := range progresses {
		if progress.Status.Current == status.Installing {
			return status.Installing, true
		}

		if progress.Status.Current == status.InstallFailed {
			return status.InstallFailed, true
		}

		if progress.Status.Current == status.RollingBack {
			return status.RollingBack, true
		}

		if progress.Status.Current == status.RollbackFailed {
			return status.RollbackFailed, true

		}

		if progress.Status.Current == status.WaitingReboot {
			return status.WaitingReboot, true
		}

		if progress.Status.Current == status.Rebooting {
			return status.Rebooting, true
		}
	}

	return "", false
}

func (h *helper) polishRebootingStatus(fixpack *fixpacks.Fixpack, finalStatus string, operation string) {
	if finalStatus == status.WaitingReboot || finalStatus == status.Rebooting {
		fixpack.Status.Current = fmt.Sprintf("%s from %s", finalStatus, operation)
	} else {
		fixpack.Status.Current = finalStatus
	}
}

func (h *helper) polishInstallationStatus(fixpack *fixpacks.Fixpack, update *update) {
	installCount := 0
	for _, progress := range update.Progresses {
		if progress.Status.Current == status.Installed {
			installCount++
		}
	}

	if installCount == len(update.Progresses) {
		fixpack.Status.Current = status.Installed
	}
}

func (h *helper) hasInprogressUpdate(filter bson.M) bool {
	count, err := h.mongo.GetCount(
		fixpacks.Db,
		fixpacks.ReqCollection,
		filter,
	)
	if err != nil {
		log.Errorf("fixpacks(%s): failed to count in-progress record(%v)", h.reqId, err)
		return false
	}

	return count > 0
}

func (h *helper) addReqRecord(node string) {
	h.reqOpts.Hostname = node
	err := h.mongo.UpdateOne(
		fixpacks.Db,
		fixpacks.ReqCollection,
		bson.M{"hostname": h.reqOpts.Hostname},
		bson.M{"$set": h.reqOpts},
		options.Update().SetUpsert(true),
	)
	if err != nil {
		log.Errorf(
			"fixpacks(%s): failed to add request record(%v)",
			h.reqId, err,
		)
	}
}

func (h *helper) changeNodeFixpackStatus(status string) error {
	return h.mongo.UpdateOne(
		fixpacks.Db,
		fixpacks.ReqCollection,
		bson.M{"hostname": h.reqOpts.Hostname, "version": h.reqOpts.Version},
		bson.M{"$set": bson.M{"status.current": status}},
	)
}

func (h *helper) deleteReqRecord() error {
	err := h.mongo.DeleteAll(
		fixpacks.Db,
		fixpacks.ReqCollection,
		bson.M{"version": h.reqOpts.Version},
	)
	if err == nil {
		return nil
	}

	log.Errorf(
		"fixpacks(%s): failed to delete request record(%v)",
		h.reqId, err,
	)
	return err
}

func (h *helper) markReqRecordAsCompleted() error {
	return h.mongo.UpdateMany(
		fixpacks.Db,
		fixpacks.ReqCollection,
		bson.M{"version": h.reqOpts.Version},
		bson.M{
			"$set": bson.M{
				"status.current":      h.reqOpts.Status.Current,
				"status.isProcessing": h.reqOpts.Status.IsProcessing,
				"status.description":  h.reqOpts.Status.Description,
			},
		},
	)
}

func (h *helper) markReqRecordAsFailed(list []nodes.Node) error {
	for _, node := range list {
		err := h.mongo.UpdateOne(
			fixpacks.Db,
			fixpacks.ReqCollection,
			bson.M{"hostname": node.Hostname, "version": h.reqOpts.Version},
			bson.M{
				"$set": bson.M{
					"status.current":      h.reqOpts.Status.Current,
					"status.isProcessing": h.reqOpts.Status.IsProcessing,
					"status.description":  h.reqOpts.Status.Description,
				},
			},
		)
		if err == nil {
			return nil
		}

		log.Errorf(
			"fixpacks(%s): failed mark req record as failed(%v)",
			h.reqId, err,
		)
	}

	return nil
}
