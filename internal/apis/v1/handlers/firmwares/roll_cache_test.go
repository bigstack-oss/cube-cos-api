package firmwares

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/firmwares"
	"github.com/gin-gonic/gin"
)

var errUnmounted = fmt.Errorf("%w: /mnt/cephfs is not mounted", cubecos.ErrRollStateUnavailable)

func resetRollCache(t *testing.T) {
	t.Helper()
	origNow := rollCacheNow
	wipe := func() {
		lastRoll.Lock()
		lastRoll.upgrade, lastRoll.at = nil, time.Time{}
		lastRoll.Unlock()
	}
	wipe()
	t.Cleanup(func() { wipe(); rollCacheNow = origNow })
}

func stubRollErr(t *testing.T, err error) {
	t.Helper()
	origRoll := getRoll
	getRoll = func() (*firmwares.Roll, error) { return nil, err }
	t.Cleanup(func() { getRoll = origRoll })
}

func callProgressHandler(t *testing.T) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/firmwares/upgradeProgress", nil)
	getFirmwareUpgradeProgress(c)
	return w
}

func TestRollingProgressServesCachedRollWhileCephfsUnmounted(t *testing.T) {
	resetRollCache(t)
	stubRoll(t,
		&firmwares.Roll{Kind: firmwares.RollKindRestart, State: firmwares.RollStateRunning},
		drainingStatus(firmwares.RollStateRunning),
	)

	h := &helper{reqId: "test"}
	if _, err := h.getRollingProgress(); err != nil {
		t.Fatal(err)
	}

	stubRollErr(t, errUnmounted)
	got, err := h.getRollingProgress()
	if err != nil {
		t.Fatalf("got err %v, want the cached projection", err)
	}
	if got.Kind != firmwares.RollKindRestart || len(got.Progresses) != 2 || got.Progresses[0].Host != "cube46" {
		t.Fatalf("got %+v, want the cached restart projection", got)
	}

	if w := callProgressHandler(t); w.Code != http.StatusOK {
		t.Fatalf("handler got %d %s, want 200 from the cache", w.Code, w.Body)
	}
}

func TestRollingProgressUnavailableWithoutCacheIs503(t *testing.T) {
	resetRollCache(t)
	stubRollErr(t, errUnmounted)

	h := &helper{reqId: "test"}
	if _, err := h.getRollingProgress(); !errors.Is(err, cubecos.ErrRollStateUnavailable) {
		t.Fatalf("got err %v, want ErrRollStateUnavailable", err)
	}

	if w := callProgressHandler(t); w.Code != http.StatusServiceUnavailable {
		t.Fatalf("handler got %d %s, want 503", w.Code, w.Body)
	}
}

func TestRollingProgressEmptyRollStatusIsNotAnEmptyRoll(t *testing.T) {
	resetRollCache(t)
	inFlight := &firmwares.Roll{Kind: firmwares.RollKindRestart, State: firmwares.RollStateRunning}
	// hex_sdk answers state "none" with no progresses once the job is gone.
	gone := &firmwares.RollStatus{IsRollingApplied: true, State: "none"}
	stubRoll(t, inFlight, gone)

	h := &helper{reqId: "test"}
	if got, err := h.getRollingProgress(); !errors.Is(err, cubecos.ErrRollStateUnavailable) {
		t.Fatalf("got %+v err %v, want ErrRollStateUnavailable", got, err)
	}

	stubRoll(t, inFlight, drainingStatus(firmwares.RollStateRunning))
	if _, err := h.getRollingProgress(); err != nil {
		t.Fatal(err)
	}

	stubRoll(t, inFlight, gone)
	got, err := h.getRollingProgress()
	if err != nil || len(got.Progresses) != 2 {
		t.Fatalf("got %+v err %v, want the cached projection", got, err)
	}
}

func TestRollingProgressCacheExpires(t *testing.T) {
	resetRollCache(t)
	now := time.Now()
	rollCacheNow = func() time.Time { return now }
	stubRoll(t,
		&firmwares.Roll{Kind: firmwares.RollKindUpgrade, State: firmwares.RollStateRunning},
		drainingStatus(firmwares.RollStateRunning),
	)

	h := &helper{reqId: "test"}
	if _, err := h.getRollingProgress(); err != nil {
		t.Fatal(err)
	}

	stubRollErr(t, errUnmounted)
	now = now.Add(rollCacheTTL + time.Second)
	if _, err := h.getRollingProgress(); !errors.Is(err, cubecos.ErrRollStateUnavailable) {
		t.Fatalf("got err %v, want ErrRollStateUnavailable after the TTL", err)
	}
}

func TestRollCacheIgnoresOtherErrors(t *testing.T) {
	resetRollCache(t)
	stubRoll(t,
		&firmwares.Roll{Kind: firmwares.RollKindRestart, State: firmwares.RollStateRunning},
		drainingStatus(firmwares.RollStateRunning),
	)

	h := &helper{reqId: "test"}
	if _, err := h.getRollingProgress(); err != nil {
		t.Fatal(err)
	}

	stubRollErr(t, errors.New("failed to unmarshal roll job"))
	if _, err := h.getRollingProgress(); err == nil || errors.Is(err, cubecos.ErrRollStateUnavailable) {
		t.Fatalf("got err %v, want the original error", err)
	}
}

func TestRollCacheClearedWhenNoRoll(t *testing.T) {
	resetRollCache(t)
	h := &helper{reqId: "test"}
	if _, err := h.fallBackToCachedRoll(&firmwares.Upgrade{Kind: firmwares.RollKindRestart}, nil); err != nil {
		t.Fatal(err)
	}

	// a local-record answer (no roll in flight) drops the cached roll
	if _, err := h.fallBackToCachedRoll(&firmwares.Upgrade{}, nil); err != nil {
		t.Fatal(err)
	}

	if _, err := h.fallBackToCachedRoll(nil, errUnmounted); !errors.Is(err, cubecos.ErrRollStateUnavailable) {
		t.Fatalf("got err %v, want no stale roll served", err)
	}
}
