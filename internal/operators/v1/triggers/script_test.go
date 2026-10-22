package triggers

import (
	"errors"
	"io/fs"
	"testing"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/triggers"
)

// A rejected name must fail before any file system call is made.
func TestApplyScriptOnFsRejectsPath(t *testing.T) {
	o := NewOperator()
	for _, name := range []string{".", "..", "../x.sh", "../../etc/x", "a/b.sh", "/etc/x"} {
		err := o.applyScriptOnFs(triggers.Script{Name: name})
		if err == nil {
			t.Fatalf("applyScriptOnFs(%q) succeeded", name)
		}

		var pathErr *fs.PathError
		if errors.As(err, &pathErr) {
			t.Fatalf("applyScriptOnFs(%q) reached the file system: %v", name, err)
		}
	}
}
