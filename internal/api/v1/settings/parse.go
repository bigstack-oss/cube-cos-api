package settings

import (
	"context"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	log "go-micro.dev/v5/logger"
	"go.mongodb.org/mongo-driver/mongo"
)

func parseTitlePrefix(cursor *mongo.Cursor) (string, error) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(5))
	defer cancel()

	for cursor.Next(ctx) {
		titlePrefix := v1.TitlePrefix{}
		err := cursor.Decode(&titlePrefix)
		if err != nil {
			continue
		}

		return titlePrefix.Value, nil
	}
	if cursor.Err() != nil {
		log.Errorf("failed to iterate email sender(%s)", cursor.Err().Error())
	}

	return "", nil
}
