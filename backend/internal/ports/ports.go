// Package ports declares the interfaces the application layer depends on.
// Outbound adapters (Postgres, Redis, in-memory, Anthropic) implement these.
package ports

import (
	"context"

	"github.com/xcreativs/gigmann/internal/core/brief"
	"github.com/xcreativs/gigmann/internal/core/facility"
	"github.com/xcreativs/gigmann/internal/intel"
)

//go:generate go tool mockgen -destination=mocks/mocks.go -package=mocks github.com/xcreativs/gigmann/internal/ports FacilityRepository,Narrator,BriefGenerator

// FacilityRepository is a driven port for reading/writing facilities.
type FacilityRepository interface {
	List(ctx context.Context) ([]facility.Facility, error)
}

// Narrator turns a computed brief context into a narrated Daily Brief.
// Implementations must narrate only the supplied figures and never fabricate numbers.
type Narrator interface {
	NarrateBrief(ctx context.Context, c intel.Context) (brief.Brief, error)
}

// BriefGenerator produces the current Daily Brief for the network (inbound use case).
type BriefGenerator interface {
	Generate(ctx context.Context) (brief.Brief, error)
}
