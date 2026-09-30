package cubecos

import (
	"reflect"
	"testing"
)

func TestParseFixpackNodeStatus(t *testing.T) {
	out := "n1|v1 v2|ok\nn2|v1|missing v2\nn3|-|unreachable\ngarbage\n"
	want := []NodeFixpackStatus{
		{Name: "n1", Installed: []string{"v1", "v2"}, Status: "ok"},
		{Name: "n2", Installed: []string{"v1"}, Status: "missing v2"},
		{Name: "n3", Installed: []string{}, Status: "unreachable"},
	}

	got := parseFixpackNodeStatus(out)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}
