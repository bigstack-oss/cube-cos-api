package cubecos

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/firmwares"
)

func resetLastGoodProgress(t *testing.T) {
	t.Helper()
	storeLastGoodProgress(nil)
	t.Cleanup(func() { storeLastGoodProgress(nil) })
}

func TestWriteUpgradeProgressIsAtomic(t *testing.T) {
	resetLastGoodProgress(t)
	dir := t.TempDir()
	path := filepath.Join(dir, "progress.json")
	err := os.WriteFile(path, []byte("old"), 0o600)
	if err != nil {
		t.Fatal(err)
	}

	want := &firmwares.Upgrade{Version: "3.1.20", Progresses: []firmwares.Progress{{Host: "cube46"}}}
	err = writeUpgradeProgress(path, want)
	if err != nil {
		t.Fatal(err)
	}

	got, err := readUpgradeProgress(path)
	if err != nil || got.Version != "3.1.20" || len(got.Progresses) != 1 {
		t.Fatalf("got %+v err=%v", got, err)
	}

	info, err := os.Stat(path)
	if err != nil || info.Mode().Perm() != 0o644 {
		t.Fatalf("got mode %v err=%v, want 0644", info.Mode().Perm(), err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) != 1 {
		t.Fatalf("got %d entries err=%v, want only progress.json (no temp left)", len(entries), err)
	}
}

func TestReadCorruptProgressWithoutLastGood(t *testing.T) {
	for name, content := range map[string][]byte{
		"nul bytes": make([]byte, 64),
		"empty":     {},
		"truncated": []byte(`{"version":"3.1.20","progr`),
	} {
		resetLastGoodProgress(t)
		path := filepath.Join(t.TempDir(), "progress.json")
		err := os.WriteFile(path, content, 0o644)
		if err != nil {
			t.Fatal(err)
		}

		got, err := readUpgradeProgress(path)
		if !errors.Is(err, ErrCorruptProgress) || got != nil {
			t.Errorf("%s: got %+v err=%v, want ErrCorruptProgress", name, got, err)
		}
	}
}

func TestReadCorruptProgressFallsBackToLastGood(t *testing.T) {
	resetLastGoodProgress(t)
	path := filepath.Join(t.TempDir(), "progress.json")
	err := writeUpgradeProgress(path, &firmwares.Upgrade{Version: "3.1.20", Progresses: []firmwares.Progress{{Host: "cube46"}}})
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(path, make([]byte, 64), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	got, err := readUpgradeProgress(path)
	if err != nil || got.Version != "3.1.20" {
		t.Fatalf("got %+v err=%v, want the last good copy", got, err)
	}

	// callers mutate what they read; the cached copy must not change
	got.Progresses[0].Host = "mutated"
	again, err := readUpgradeProgress(path)
	if err != nil || again.Progresses[0].Host != "cube46" {
		t.Fatalf("got %+v err=%v, want an unshared copy", again, err)
	}
}
