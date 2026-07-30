//go:build linux && privileged

// Regression tests for parseFixpackInfo, which mounts a fixpack image to read
// fixpack.info out of it.
//
// A plain `mount <file> <dir>` is a read-write mount. On a writable filesystem
// image it rewrites the superblock (mount count, last write time), so reading
// the metadata used to change the artifact's md5 on every single call. See
// mountFixpackReadOnly.
//
// These tests need root and a free loop device, so they sit behind the
// `privileged` build tag and never run in `task test`. Run them with:
//
//	task testFixpackMount
package cubecos

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

const testFixpackInfo = "FIXPACK_ID=\"3.2.0\"\n" +
	"FIXPACK_NAME=\"regression-fixture\"\n" +
	"ROLLBACK=\"yes\"\n" +
	"REBOOT_REQUIRED=\"control compute\"\n"

// ext4 and ext4Dirty are the shapes a real fixpack takes: hex's makehotfix and
// makefixpack both build the image with `CreateFsImage ext4`, and
// hex_fixpack_install mounts it with an explicit `-t ext4`. ext4Dirty is the one
// that a plain `-o ro` refuses outright, because a read-only loop device cannot
// replay a journal -- an interrupted mount by the old read-write code is enough
// to leave a real fixpack in that state. squashfs and ext2 are defensive cover
// in case the build ever changes format; nothing ships them today.
var fixpackImageKinds = []string{"squashfs", "ext2", "ext4", "ext4Dirty"}

func run(t *testing.T, name string, args ...string) {
	t.Helper()

	out, err := exec.Command(name, args...).CombinedOutput()
	require.NoErrorf(t, err, "%s %v failed: %s", name, args, string(out))
}

func requireTools(t *testing.T, names ...string) {
	t.Helper()

	for _, name := range names {
		_, err := exec.LookPath(name)
		if err != nil {
			t.Skipf("%s is not installed; run this suite via `task testFixpackMount`", name)
		}
	}
}

// Builds an artifact shaped like a real fixpack: a filesystem image with
// fixpack.info at its root. Returns the image path.
func buildFixpackImage(t *testing.T, kind string) string {
	t.Helper()
	requireTools(t, "mount", "umount")

	dir := t.TempDir()
	img := filepath.Join(dir, "fixture.fixpack")

	if kind == "squashfs" {
		requireTools(t, "mksquashfs")

		src := filepath.Join(dir, "src")
		require.NoError(t, os.MkdirAll(src, 0755))
		require.NoError(t, os.WriteFile(filepath.Join(src, "fixpack.info"), []byte(testFixpackInfo), 0644))
		run(t, "mksquashfs", src, img, "-quiet", "-no-progress")

		return img
	}

	mkfs := "mkfs.ext4"
	if kind == "ext2" {
		mkfs = "mkfs.ext2"
	}
	requireTools(t, mkfs)

	run(t, "dd", "if=/dev/zero", "of="+img, "bs=1M", "count=16", "status=none")
	run(t, mkfs, "-q", "-F", img)

	mnt := filepath.Join(dir, "mnt")
	require.NoError(t, os.MkdirAll(mnt, 0755))
	run(t, "mount", img, mnt)
	require.NoError(t, os.WriteFile(filepath.Join(mnt, "fixpack.info"), []byte(testFixpackInfo), 0644))
	run(t, "umount", mnt)
	run(t, "sync")

	if kind == "ext4Dirty" {
		requireTools(t, "debugfs")

		// Mark the journal as needing recovery, the state an interrupted mount
		// leaves behind. A read-only mount cannot replay it.
		run(t, "debugfs", "-w", "-R", "ssv state 0", img)
		run(t, "debugfs", "-w", "-R", "feature needs_recovery", img)
	}

	return img
}

func md5OfFile(t *testing.T, path string) string {
	t.Helper()

	f, err := os.Open(path)
	require.NoError(t, err)
	defer f.Close()

	h := md5.New()
	_, err = io.Copy(h, f)
	require.NoError(t, err)

	return hex.EncodeToString(h.Sum(nil))
}

// The bug: reading the metadata rewrote the artifact. Mounting must leave the
// file byte-for-byte identical, whatever the image format.
func TestParseFixpackInfoLeavesTheImageUntouched(t *testing.T) {
	for _, kind := range fixpackImageKinds {
		t.Run(kind, func(t *testing.T) {
			img := buildFixpackImage(t, kind)
			before := md5OfFile(t, img)

			info, err := parseFixpackInfo(img)

			require.NoError(t, err)
			require.Contains(t, string(info), `FIXPACK_ID="3.2.0"`)
			require.Equal(t, before, md5OfFile(t, img), "mounting the fixpack rewrote the artifact")
		})
	}
}

// The reported symptom, stated directly: the md5 changed on every read. The
// list endpoints call this once per .fixpack in /var/fixpack, so a single
// request used to churn every artifact on the node.
func TestParseFixpackInfoKeepsTheSameMd5AcrossRepeatedReads(t *testing.T) {
	img := buildFixpackImage(t, "ext4")
	want := md5OfFile(t, img)

	for range 5 {
		_, err := parseFixpackInfo(img)

		require.NoError(t, err)
		require.Equal(t, want, md5OfFile(t, img))
	}
}

// Pins the noload fallback. Without it, `-o ro` fails on this image with
// "cannot mount /dev/loop0 read-only" and the fixpack becomes unreadable --
// which the old read-write code could itself cause by being interrupted.
func TestParseFixpackInfoStillReadsAnExt4ImageWithADirtyJournal(t *testing.T) {
	img := buildFixpackImage(t, "ext4Dirty")

	info, err := parseFixpackInfo(img)

	require.NoError(t, err)
	require.Contains(t, string(info), `FIXPACK_ID="3.2.0"`)
}

// The parsing on top of the mount still works, so the read-only options did not
// cost us access to any part of the image.
func TestGetFixpackInfoParsesAReadOnlyMountedImage(t *testing.T) {
	img := buildFixpackImage(t, "ext4")

	raw, err := GetFixpackInfo(img)

	require.NoError(t, err)
	require.Equal(t, "3.2.0", raw.Id)
	require.Equal(t, "regression-fixture", raw.Name)
	require.True(t, raw.Rollbackable)
	require.Contains(t, raw.RebootRequired, "control")
}

// A non-image file must fail, not hang and not leave a mount behind.
func TestParseFixpackInfoRejectsAFileThatIsNotAFilesystemImage(t *testing.T) {
	img := filepath.Join(t.TempDir(), "garbage.fixpack")
	require.NoError(t, os.WriteFile(img, []byte("not a filesystem"), 0644))

	_, err := parseFixpackInfo(img)

	require.ErrorContains(t, err, "failed to mount fixpack")
}
