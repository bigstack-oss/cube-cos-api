package triggers

import (
	"encoding/base64"
	"testing"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/triggers"
	"github.com/stretchr/testify/require"
)

func scriptHelper(name, content string) *helper {
	h := &helper{}
	h.reqOpts.Name = "t1"
	h.reqOpts.Response.Script = triggers.Script{Name: name, Content: content}
	return h
}

func TestCheckScriptAcceptsFileName(t *testing.T) {
	content := base64.StdEncoding.EncodeToString([]byte("#!/bin/bash\necho hi\n"))
	require.NoError(t, scriptHelper("notify.sh", content).checkScript())
	require.NoError(t, scriptHelper("notify.sh", "").checkScript())
	require.NoError(t, scriptHelper("", "").checkScript())
}

// The name is checked even when the request carries no script content.
func TestCheckScriptRejectsPath(t *testing.T) {
	content := base64.StdEncoding.EncodeToString([]byte("#!/bin/bash\necho hi\n"))
	for _, name := range []string{".", "..", "../x.sh", "../../etc/x", "a/b.sh", "/etc/x"} {
		require.Error(t, scriptHelper(name, content).checkScript(), name)
		require.Error(t, scriptHelper(name, "").checkScript(), name)
	}
}
