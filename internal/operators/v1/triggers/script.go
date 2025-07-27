package triggers

import (
	"encoding/base64"
	"os"

	log "go-micro.dev/v5/logger"

	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/triggers"
)

func (o *Operator) syncScripts(apiTrigger triggers.ApiSchema, cosTrigger triggers.CosSchema) error {
	err := o.applyScriptOnFs(apiTrigger)
	if err != nil {
		log.Errorf("triggers: failed to apply script on fs %s(%v)", apiTrigger.FilePath, err)
		return err
	}

	err = cubecos.ApplyScript(cosTrigger.WriteResponses.Execs)
	if err != nil {
		log.Errorf("triggers: failed to apply script %s(%v)", apiTrigger.FilePath, err)
		return err
	}

	return nil
}

func (o *Operator) applyScriptOnFs(trigger triggers.ApiSchema) error {
	if trigger.FilePath == "" {
		return nil
	}

	bytes, err := base64.StdEncoding.DecodeString(trigger.Content)
	if err != nil {
		log.Errorf("triggers: failed to decode script content %s(%v)", trigger.Content, err)
		return err
	}

	err = os.WriteFile(trigger.FilePath, bytes, 0755)
	if err != nil {
		log.Errorf("triggers: failed to write script file %s(%v)", trigger.FilePath, err)
		return err
	}

	return nil
}
