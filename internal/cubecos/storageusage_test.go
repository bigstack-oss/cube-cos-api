package cubecos

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func withStorageUsageInterval(t *testing.T, raw string, err error) {
	t.Helper()
	orig := readStorageUsageInterval
	readStorageUsageInterval = func() (string, error) { return raw, err }
	t.Cleanup(func() { readStorageUsageInterval = orig })
}

func TestStorageUsageLookbackFollowsInterval(t *testing.T) {
	cases := []struct {
		name          string
		raw           string
		err           error
		rank, history int
	}{
		{"default", "15", nil, 30, 60},
		{"30m", "30", nil, 60, 120},
		{"hourly", "60", nil, 120, 240},
		{"daily", "1440", nil, 2880, 5760},
		{"below min clamps", "1", nil, 10, 60},
		{"above max clamps", "9999", nil, 2880, 5760},
		{"unreadable falls back", "", errors.New("no tuning"), 30, 60},
		{"garbage falls back", "abc", nil, 30, 60},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			withStorageUsageInterval(t, c.raw, c.err)
			require.Equal(t, c.rank, StorageUsageRankMinutes())
			require.Equal(t, c.history, StorageUsageHistoryMinutes())
		})
	}
}
