package supportfiles

import (
	"context"

	bsmongo "github.com/bigstack-oss/bigstack-dependency-go/pkg/mongo"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/support"
	log "go-micro.dev/v5/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func (h *helper) syncCreatingFiles(files *[]support.File) {
	mongo := bsmongo.GetGlobalHelper()
	c, err := mongo.GetQueryCursor(
		support.FileDB,
		support.FileReqCollection,
		bson.M{"status.current": "creating"},
	)
	if err != nil {
		log.Errorf("supportFiles(%s): failed to get creating file set(%v)", h.reqId, err)
		return
	}

	ctx, cancel := context.WithTimeout(wait.CtxSeconds(10))
	defer cancel()
	defer c.Close(ctx)
	h.setCreatingFile(files, c)
}

func (h *helper) setCreatingFile(files *[]support.File, c *mongo.Cursor) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(10))
	defer cancel()
	for c.Next(ctx) {
		file := support.File{}
		err := c.Decode(&file)
		if err != nil {
			log.Errorf("supportFiles(%s): failed to decode creating file set(%v)", h.reqId, err)
			continue
		}

		*files = append(*files, file)
	}
	if c.Err() != nil {
		log.Errorf(
			"supportFiles(%s): failed to iterate support file cursor(%s)",
			h.reqId,
			c.Err().Error(),
		)
	}
}
