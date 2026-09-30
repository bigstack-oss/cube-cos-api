package cubecos

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetRollFromMissingJob(t *testing.T) {
	dir := t.TempDir()

	// not cephfs: the job's absence proves nothing
	roll, err := getRollFrom(filepath.Join(dir, "job.json"), dir)
	if err == nil || roll != nil {
		t.Fatalf("got roll=%v err=%v, want an error", roll, err)
	}
}

func TestGetRollFromExistingJob(t *testing.T) {
	dir := t.TempDir()
	job := filepath.Join(dir, "job.json")
	err := os.WriteFile(job, []byte(`{"kind":"upgrade","state":"running"}`), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	roll, err := getRollFrom(job, dir)
	if err != nil || roll == nil || !roll.IsInFlight() {
		t.Fatalf("got roll=%+v err=%v, want an in-flight roll", roll, err)
	}
}
