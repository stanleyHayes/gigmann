package app

import (
	"context"

	"github.com/xcreativs/gigmann/internal/core/brief"
	"github.com/xcreativs/gigmann/internal/core/signal"
	"github.com/xcreativs/gigmann/internal/ports"
)

// StaticBrief adapts a BriefService + a captured signal input into a
// ports.BriefGenerator (used while metric/inventory repos don't yet exist).
type StaticBrief struct {
	svc *BriefService
	in  signal.Input
}

// Compile-time guarantee that StaticBrief satisfies the port.
var _ ports.BriefGenerator = (*StaticBrief)(nil)

// NewStaticBrief wires a BriefService over a fixed signal input.
func NewStaticBrief(svc *BriefService, in signal.Input) *StaticBrief {
	return &StaticBrief{svc: svc, in: in}
}

// Generate produces the Daily Brief for the captured input.
func (s *StaticBrief) Generate(ctx context.Context) (brief.Brief, error) {
	return s.svc.Generate(ctx, s.in)
}
