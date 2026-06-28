package app

import (
	"context"
	"fmt"

	"github.com/xcreativs/gigmann/internal/core/brief"
	"github.com/xcreativs/gigmann/internal/core/signal"
	"github.com/xcreativs/gigmann/internal/intel"
	"github.com/xcreativs/gigmann/internal/ports"
)

// BriefService orchestrates the Daily Brief pipeline (spec §6.2): run the
// deterministic signal engine, summarise the network pulse, assemble context,
// have the narrator write the brief, then validate it.
type BriefService struct {
	engine   *signal.Engine
	narrator ports.Narrator
	topN     int
}

// NewBriefService wires the brief pipeline. topN bounds how many signals are
// handed to the narrator (0 = all).
func NewBriefService(engine *signal.Engine, narrator ports.Narrator, topN int) *BriefService {
	return &BriefService{engine: engine, narrator: narrator, topN: topN}
}

// Generate runs the pipeline and returns a validated Daily Brief.
func (s *BriefService) Generate(ctx context.Context, in signal.Input) (brief.Brief, error) {
	signals := s.engine.Run(in)
	pulse := signal.NetworkPulse(in.Facilities, signals)
	c := intel.BuildContext(in.AsOf, in.Facilities, signals, pulse, s.topN)

	b, err := s.narrator.NarrateBrief(ctx, c)
	if err != nil {
		return brief.Brief{}, fmt.Errorf("app: narrate brief: %w", err)
	}

	// Grounding guardrail: drop any item the model attached to an invented facility.
	b.Items = groundBriefItems(b.Items, knownFacilityIDs(in.Facilities))

	// Re-validate the narrated brief against domain invariants.
	validated, err := brief.New(b)
	if err != nil {
		return brief.Brief{}, fmt.Errorf("app: invalid narrated brief: %w", err)
	}
	return validated, nil
}
