package observability_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/xcreativs/gigmann/internal/observability"
)

func TestRecordAICall(t *testing.T) {
	require.NotNil(t, observability.AIRegistry())
	// Records without panicking for success + error + zero-token paths.
	observability.RecordAICall("brief", 120, 340, 1500*time.Millisecond, nil)
	observability.RecordAICall("ask", 0, 0, 200*time.Millisecond, errors.New("boom"))

	mfs, err := observability.AIRegistry().Gather()
	require.NoError(t, err)
	assert.NotEmpty(t, mfs, "AI metrics are registered and gatherable")
}
