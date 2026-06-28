package app_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/xcreativs/gigmann/internal/app"
	"github.com/xcreativs/gigmann/internal/intel"
	"github.com/xcreativs/gigmann/internal/ports/mocks"
)

func TestDraftServiceBuildsGroundedPrompt(t *testing.T) {
	ctrl := gomock.NewController(t)
	answerer := mocks.NewMockQuestionAnswerer(ctrl)
	var prompt string
	answerer.EXPECT().Answer(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, q string) (intel.Answer, error) {
			prompt = q
			return intel.Answer{Text: "Dear manager, regarding the denial spike..."}, nil
		})

	svc := app.NewDraftService(answerer)
	out, err := svc.Draft(context.Background(), execPrincipal(), "message", "kasoa", "the NHIS denial spike")
	require.NoError(t, err)
	assert.Equal(t, "Dear manager, regarding the denial spike...", out)
	assert.Contains(t, prompt, "kasoa")
	assert.Contains(t, prompt, "denial spike")
	assert.Contains(t, prompt, "never invent numbers", "grounding instruction in the prompt")
}

func TestDraftServiceEmptyInstruction(t *testing.T) {
	ctrl := gomock.NewController(t)
	answerer := mocks.NewMockQuestionAnswerer(ctrl) // no call expected
	svc := app.NewDraftService(answerer)
	out, err := svc.Draft(context.Background(), execPrincipal(), "message", "", "   ")
	require.NoError(t, err)
	assert.Empty(t, out)
}

func TestDraftServiceScopesManager(t *testing.T) {
	ctrl := gomock.NewController(t)
	answerer := mocks.NewMockQuestionAnswerer(ctrl) // no Answer expected: authz fails first
	svc := app.NewDraftService(answerer)

	_, err := svc.Draft(context.Background(), managerPrincipal("kasoa"), "message", "nima", "x")
	require.ErrorIs(t, err, app.ErrForbidden, "manager cannot draft for another facility")

	_, err = svc.Draft(context.Background(), managerPrincipal("kasoa"), "summary", "", "x")
	require.ErrorIs(t, err, app.ErrForbidden, "manager cannot make a network-wide draft")
}
