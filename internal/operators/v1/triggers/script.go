package triggers

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"

	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/settings"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/triggers"
	log "go-micro.dev/v5/logger"
)

func (o *Operator) syncScripts(script triggers.Script) error {
	err := o.applyScriptOnFs(script)
	if err != nil {
		log.Errorf("triggers: failed to apply script on fs %s(%v)", script.Name, err)
		return err
	}

	err = cubecos.ApplyScript([]string{script.Name})
	if err != nil {
		log.Errorf("triggers: failed to apply script %s(%v)", script.Name, err)
		return err
	}

	return nil
}

func (o *Operator) applyScriptOnFs(script triggers.Script) error {
	if script.Name == "" {
		return nil
	}

	bytes, err := base64.StdEncoding.DecodeString(script.Content)
	if err != nil {
		log.Errorf("triggers: failed to decode script content %s(%v)", script.Content, err)
		return err
	}

	path := filepath.Join(settings.ScriptDir, fmt.Sprintf("%s.shell", script.Name))
	err = os.WriteFile(path, bytes, 0755)
	if err != nil {
		log.Errorf("triggers: failed to write script file %s(%v)", script.Name, err)
		return err
	}

	return nil
}
