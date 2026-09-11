package cubecos

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"
)

// writeLicense builds a .license archive at dir/<filename> whose members carry
// memberStem, which need not match filename.
func writeLicense(t *testing.T, dir, filename, memberStem, dat string) string {
	t.Helper()

	path := filepath.Join(dir, filename)
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}

	defer f.Close()
	zw := zip.NewWriter(f)
	for ext, body := range map[string]string{"dat": dat, "sig": "signature-bytes"} {
		w, err := zw.Create(memberStem + "." + ext)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write([]byte(body)); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}

	return path
}

const testDat = "license.name=NYCU ADFP3.0\n" +
	"license.type=enterprise\n" +
	"issue.hardware=GRRYSF4,DVRYSF4\n" +
	"product=CubeCOS\n"

// Member names need not match the uploaded filename.
func TestParseLicenseDatIgnoresUploadedFilename(t *testing.T) {
	dir := t.TempDir()
	path := writeLicense(t, dir, "nycu-adfp3-0 (1).license", "nycu-adfp3-0", testDat)

	license, err := parseLicenseDat(path)
	if err != nil {
		t.Fatalf("parseLicenseDat: %v", err)
	}
	if license.Name != "NYCU ADFP3.0" {
		t.Errorf("Name = %q, want %q", license.Name, "NYCU ADFP3.0")
	}
	if license.Issue.Hardware != "GRRYSF4,DVRYSF4" {
		t.Errorf("Hardware = %q", license.Issue.Hardware)
	}
}

// A stale pair beside the upload must not be picked up.
func TestParseLicenseDatIgnoresStaleMembers(t *testing.T) {
	dir := t.TempDir()
	for _, ext := range []string{"dat", "sig"} {
		stale := filepath.Join(dir, "someone-elses."+ext)
		if err := os.WriteFile(stale, []byte("license.name=someone else\n"), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	path := writeLicense(t, dir, "mine.license", "mine", testDat)
	license, err := parseLicenseDat(path)
	if err != nil {
		t.Fatalf("parseLicenseDat: %v", err)
	}
	if license.Name != "NYCU ADFP3.0" {
		t.Errorf("Name = %q, want the uploaded license, not the stale pair", license.Name)
	}
}

// An archive without a .dat/.sig pair is genuinely malformed.
func TestParseLicenseDatRejectsArchiveWithoutPair(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "broken.license")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)
	w, err := zw.Create("readme.txt")
	if err != nil {
		t.Fatal(err)
	}
	w.Write([]byte("not a license"))
	zw.Close()
	f.Close()

	if _, err := parseLicenseDat(path); err == nil {
		t.Fatal("parseLicenseDat accepted an archive with no .dat/.sig pair")
	}
}

// parseLicenseDat must not leave extracted members behind.
func TestParseLicenseDatCleansUpMembers(t *testing.T) {
	dir := t.TempDir()
	path := writeLicense(t, dir, "tidy.license", "tidy", testDat)

	if _, err := parseLicenseDat(path); err != nil {
		t.Fatal(err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if e.Name() != "tidy.license" {
			t.Errorf("left behind %q in the verify dir", e.Name())
		}
	}
}
