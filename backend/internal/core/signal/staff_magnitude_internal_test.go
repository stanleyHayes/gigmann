package signal

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestLicenceExpiryMagnitude exercises the urgency-normalisation directly,
// including the zero-window guard and the negative-magnitude clamp that the
// public StaffDetector.Detect path cannot reach (it only flags within-window).
func TestLicenceExpiryMagnitude(t *testing.T) {
	const window = 30
	day := 24 * time.Hour

	assert.InDelta(t, 1.0, licenceExpiryMagnitude(0, window), 1e-9, "expiring today is most urgent")
	assert.InDelta(t, 0.5, licenceExpiryMagnitude(15*day, window), 1e-9, "mid-window")
	assert.InDelta(t, 0.0, licenceExpiryMagnitude(time.Duration(window)*day, window), 1e-9, "just entering the window")
	assert.Greater(t, licenceExpiryMagnitude(-10*day, window), 1.0, "an already-expired licence ranks above any in-window one")
	assert.InDelta(t, 1.0, licenceExpiryMagnitude(5*day, 0), 1e-9, "zero-window guard returns a constant")
	assert.InDelta(t, 0.0, licenceExpiryMagnitude(60*day, window), 1e-9, "beyond the window clamps to 0")
}
