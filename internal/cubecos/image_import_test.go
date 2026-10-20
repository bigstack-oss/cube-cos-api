package cubecos

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/images"
	"github.com/stretchr/testify/require"
)

func validImportReqOpts() images.ReqOpts {
	return images.ReqOpts{
		File:        "ubuntu-22.04.qcow2",
		Name:        "new-img",
		Os:          "Ubuntu",
		Destination: "cube-storage",
		Domain:      "Default",
		Project:     "alpha",
		Visibility:  "private",
	}
}

func validCreateOpts() *images.CreateOpts {
	return &images.CreateOpts{
		Dir:         images.GlanceDir,
		File:        "ubuntu-22.04.qcow2",
		Name:        "new-img",
		Destination: "cube-storage",
		Os:          "ubuntu",
		Domain:      "Default",
		Project:     "alpha",
		PoolType:    "glance-images",
		Visibility:  "private",
	}
}

func TestValidateImportOptsAcceptsCommonValues(t *testing.T) {
	require.NoError(t, ValidateImportOpts(validImportReqOpts()))

	opts := validImportReqOpts()
	opts.Name = "映像檔_1"
	opts.File = "c++.img"
	opts.Project = "proj.v2"
	require.NoError(t, ValidateImportOpts(opts))

	opts = validImportReqOpts()
	opts.File = "amphora-x64-haproxy.qcow2"
	opts.Name = "amphora"
	require.NoError(t, ValidateImportOpts(opts))
}

func TestValidateImportOptsAccepts255Bytes(t *testing.T) {
	opts := validImportReqOpts()
	opts.Name = strings.Repeat("a", 255)
	require.NoError(t, ValidateImportOpts(opts))
}

func TestValidateImportOptsRejectsUnsafeValues(t *testing.T) {
	values := []string{
		"",
		"a b",
		"a/b",
		"..",
		".hidden",
		"-x",
		"_x",
		"a;b",
		"a$b",
		"a`b",
		`a"b`,
		"a'b",
		"a\nb",
		"a\n",
		"a\\b",
		strings.Repeat("a", 256),
	}

	fields := map[string]func(*images.ReqOpts, string){
		"file":        func(o *images.ReqOpts, v string) { o.File = v },
		"name":        func(o *images.ReqOpts, v string) { o.Name = v },
		"os":          func(o *images.ReqOpts, v string) { o.Os = v },
		"destination": func(o *images.ReqOpts, v string) { o.Destination = v },
		"domain":      func(o *images.ReqOpts, v string) { o.Domain = v },
		"project":     func(o *images.ReqOpts, v string) { o.Project = v },
		"visibility":  func(o *images.ReqOpts, v string) { o.Visibility = v },
	}

	for field, set := range fields {
		for _, value := range values {
			opts := validImportReqOpts()
			set(&opts, value)

			err := ValidateImportOpts(opts)
			require.Error(t, err, "%s=%q", field, value)
			require.Contains(t, err.Error(), "invalid "+field, "%s=%q", field, value)
		}
	}
}

func TestGenCreateOptsRejectsInvalidReqOpts(t *testing.T) {
	reqOpts := validImportReqOpts()
	reqOpts.File = "../x"

	opts, err := GenCreateOptsByReqOpts(reqOpts)
	require.Error(t, err)
	require.Nil(t, opts)
}

func TestGenCreateOptsKeepsFileInsideGlanceDir(t *testing.T) {
	opts, err := GenCreateOptsByReqOpts(validImportReqOpts())
	require.NoError(t, err)
	require.Equal(t, images.GlanceDir, filepath.Dir(filepath.Join(opts.Dir, opts.File)))
}

func TestGenImageArgsPassesFlagsAsOneArgument(t *testing.T) {
	args := genImageArgs(validCreateOpts())
	require.Equal(t, []string{
		"os_image_import",
		images.GlanceDir,
		"ubuntu-22.04.qcow2",
		"new-img",
		"--os-project-domain-name Default --os-project-name alpha --visibility private",
		"glance-images",
		"cube-storage",
		"ubuntu",
	}, args)
}

func TestGenImageArgsReservedTypes(t *testing.T) {
	opts := validCreateOpts()
	opts.ReservedType = "lb"
	require.Equal(t, []string{"os_octavia_image_import", images.GlanceDir, "ubuntu-22.04.qcow2"}, genImageArgs(opts))

	opts.ReservedType = "fs"
	require.Equal(t, []string{"os_manila_image_import", images.GlanceDir, "ubuntu-22.04.qcow2"}, genImageArgs(opts))
}

func TestGenCmdRunsHexSdkDirectly(t *testing.T) {
	t.Chdir(t.TempDir())
	opts := validCreateOpts()

	cmd := genCmd(context.Background(), opts)
	require.Equal(t, "hex_sdk", cmd.Args[0])
	require.Equal(t, genImageArgs(opts), cmd.Args[1:])

	scripts, err := filepath.Glob("image-import-*.sh")
	require.NoError(t, err)
	require.Empty(t, scripts)
}
