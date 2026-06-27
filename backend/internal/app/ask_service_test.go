package app_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/xcreativs/gigmann/internal/app"
	"github.com/xcreativs/gigmann/internal/core/signal"
	"github.com/xcreativs/gigmann/internal/intel"
	"github.com/xcreativs/gigmann/internal/ports/mocks"
)

func TestAskServiceAnswer(t *testing.T) {
	ctrl := gomock.NewController(t)
	answerer := mocks.NewMockAnswerer(ctrl)
	answerer.EXPECT().Answer(gomock.Any(), "Why is Tafo critical?", gomock.Any()).
		Return(intel.Answer{Text: "Claims are unsubmitted.", Citations: []string{"tafo-maternity"}}, nil)

	svc := app.NewAskService(signal.Default(signal.DefaultThresholds()), answerer, briefInput(time.Now().UTC()), 0)
	a, err := svc.Answer(context.Background(), "Why is Tafo critical?")
	require.NoError(t, err)
	assert.Equal(t, "Claims are unsubmitted.", a.Text)
	assert.Contains(t, a.Citations, "tafo-maternity")
}

func TestAskServiceBlankQuestion(t *testing.T) {
	ctrl := gomock.NewController(t)
	answerer := mocks.NewMockAnswerer(ctrl) // no call expected

	svc := app.NewAskService(signal.Default(signal.DefaultThresholds()), answerer, briefInput(time.Now().UTC()), 0)
	a, err := svc.Answer(context.Background(), "   ")
	require.NoError(t, err)
	assert.Contains(t, a.Text, "Please ask")
}

func TestAskServiceAnswererError(t *testing.T) {
	ctrl := gomock.NewController(t)
	answerer := mocks.NewMockAnswerer(ctrl)
	answerer.EXPECT().Answer(gomock.Any(), gomock.Any(), gomock.Any()).Return(intel.Answer{}, errors.New("api down"))

	svc := app.NewAskService(signal.Default(signal.DefaultThresholds()), answerer, briefInput(time.Now().UTC()), 0)
	_, err := svc.Answer(context.Background(), "anything")
	require.Error(t, err)
}

func TestAskServiceCapsQuestionLength(t *testing.T) {
	ctrl := gomock.NewController(t)
	answerer := mocks.NewMockAnswerer(ctrl)
	var gotLen int
	answerer.EXPECT().Answer(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, q string, _ intel.Context) (intel.Answer, error) {
			gotLen = utf8.RuneCountInString(q)
			return intel.Answer{Text: "ok"}, nil
		})

	svc := app.NewAskService(signal.Default(signal.DefaultThresholds()), answerer, briefInput(time.Now().UTC()), 0)
	_, err := svc.Answer(context.Background(), strings.Repeat("é", 1500)) // multi-byte → rune-based cap
	require.NoError(t, err)
	assert.Equal(t, 1000, gotLen, "question capped to 1000 runes")
}
