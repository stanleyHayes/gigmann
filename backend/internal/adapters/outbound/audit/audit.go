// Package audit implements ports.AuditLogger over slog, emitting one structured
// "audit" line per security-relevant event (auth, approval decisions).
package audit

import (
	"context"
	"log/slog"

	"github.com/xcreativs/gigmann/internal/ports"
)

// Logger records audit events to a structured slog logger.
type Logger struct {
	log *slog.Logger
}

var _ ports.AuditLogger = Logger{}

// New builds an audit Logger over the given slog logger.
func New(log *slog.Logger) Logger { return Logger{log: log} }

// Record emits the audit event.
func (l Logger) Record(ctx context.Context, e ports.AuditEvent) {
	l.log.LogAttrs(ctx, slog.LevelInfo, "audit",
		slog.String("actor", e.Actor),
		slog.String("action", e.Action),
		slog.String("target", e.Target),
		slog.String("outcome", e.Outcome),
	)
}
