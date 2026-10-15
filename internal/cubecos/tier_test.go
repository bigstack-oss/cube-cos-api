package cubecos

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/storages"
)

func TestParseStorageTiers(t *testing.T) {
	got, err := parseStorageTiers("registered|gold\nregistry-only|silver\n\nunmanaged|computehdd\n")
	if err != nil {
		t.Fatal(err)
	}
	want := []storages.Tier{
		{Name: "gold", State: storages.TierRegistered},
		{Name: "silver", State: storages.TierRegistryOnly},
		{Name: "computehdd", State: storages.TierUnmanaged},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}

	if _, err := parseStorageTiers("garbage"); err == nil {
		t.Fatal("a line without a state was accepted")
	}
}

func TestCreateStorageTierPassesOutput(t *testing.T) {
	var gotArgs []string
	runHexSdk = func(ctx context.Context, args ...string) ([]byte, error) {
		gotArgs = args
		return []byte("Error: ceph is recovering\n"), errors.New("exit status 1")
	}

	out, err := CreateStorageTier("gold", []string{"0", "osd.1"})
	if err == nil || err.Error() != "Error: ceph is recovering" || out != err.Error() {
		t.Fatalf("got %q %v", out, err)
	}
	if !reflect.DeepEqual(gotArgs, []string{"ceph_device_tier_create", "gold", "0", "osd.1"}) {
		t.Fatalf("args %v", gotArgs)
	}
}
