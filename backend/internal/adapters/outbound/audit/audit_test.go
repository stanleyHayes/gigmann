package audit_test

import (
	"bytes"
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/xcreativs/gigmann/internal/adapters/outbound/audit"
	"github.com/xcreativs/gigmann/internal/ports"
)

func TestRecord(t *testing.T) {
	var buf bytes.Buffer
	a := audit.New(slog.New(slog.NewJSONHandler(&buf, nil)))
	a.Record(context.Background(), ports.AuditEvent{Actor: "u1", Action: "auth.login", Outcome: "success"})

	out := buf.String()
	assert.Contains(t, out, `"msg":"audit"`)
	assert.Contains(t, out, `"action":"auth.login"`)
	assert.Contains(t, out, `"actor":"u1"`)
	assert.Contains(t, out, `"outcome":"success"`)
}
