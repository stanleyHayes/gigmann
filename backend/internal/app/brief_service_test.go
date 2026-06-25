package app_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/xcreativs/gigmann/internal/app"
	"github.com/xcreativs/gigmann/internal/core/brief"
	"github.com/xcreativs/gigmann/internal/core/severity"
	"github.com/xcreativs/gigmann/internal/core/signal"
	"github.com/xcreativs/gigmann/internal/intel"
	"github.com/xcreativs/gigmann/internal/ports/mocks"
	"github.com/xcreativs/gigmann/internal/seed"
)

func briefInput(asOf time.Time) signal.Input {
	net := seed.Generate(7, asOf, 14)
	return signal.Input{
		AsOf: asOf, Facilities: net.Facilities, Metrics: net.Metrics,
		Inventory: net.Inventory, Staff: net.Staff,
	}
}

func validBrief(asOf time.Time) brief.Brief {
	return brief.Brief{
		ID: "b1", Date: asOf, Prose: "Good morning, Sammy.",
		Items: []brief.Item{{Severity: severity.Critical, FacilityID: "tafo-maternity", Headline: "Tafo first"}},
		Model: "claude-sonnet-4-6",
	}
}

func TestBriefServiceGenerate(t *testing.T) {
	asOf := time.Date(2026, 6, 9, 0, 0, 0, 0, time.UTC)
	ctrl := gomock.NewController(t)
	narrator := mocks.NewMockNarrator(ctrl)
	narrator.EXPECT().NarrateBrief(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, c intel.Context) (brief.Brief, error) {
			require.NotEmpty(t, c.Items, "engine should have surfaced signals into the context")
			return validBrief(asOf), nil
		})

	svc := app.NewBriefService(signal.Default(signal.DefaultThresholds()), narrator, 5)
	b, err := svc.Generate(context.Background(), briefInput(asOf))

	require.NoError(t, err)
	require.Len(t, b.Items, 1)
	assert.Equal(t, severity.Critical, b.Items[0].Severity)
}

func TestBriefServiceNarratorError(t *testing.T) {
	ctrl := gomock.NewController(t)
	narrator := mocks.NewMockNarrator(ctrl)
	narrator.EXPECT().NarrateBrief(gomock.Any(), gomock.Any()).Return(brief.Brief{}, errors.New("api down"))

	svc := app.NewBriefService(signal.Default(signal.DefaultThresholds()), narrator, 5)
	_, err := svc.Generate(context.Background(), briefInput(time.Now().UTC()))
	require.Error(t, err)
}

func TestBriefServiceInvalidNarratedBrief(t *testing.T) {
	ctrl := gomock.NewController(t)
	narrator := mocks.NewMockNarrator(ctrl)
	narrator.EXPECT().NarrateBrief(gomock.Any(), gomock.Any()).Return(brief.Brief{ID: ""}, nil)

	svc := app.NewBriefService(signal.Default(signal.DefaultThresholds()), narrator, 5)
	_, err := svc.Generate(context.Background(), briefInput(time.Now().UTC()))
	require.Error(t, err)
}
