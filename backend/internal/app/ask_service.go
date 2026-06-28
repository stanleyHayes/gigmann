package app

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/xcreativs/gigmann/internal/core/signal"
	"github.com/xcreativs/gigmann/internal/intel"
	"github.com/xcreativs/gigmann/internal/ports"
)

// AskService is the "Ask" use case: answer a natural-language question grounded
// in the freshly computed network context (same deterministic signals as the brief).
type AskService struct {
	engine   *signal.Engine
	answerer ports.Answerer
	input    signal.Input
	topN     int
}

// maxQuestionLen bounds a question (runes) — input validation + AI-cost control.
const maxQuestionLen = 1000

var _ ports.QuestionAnswerer = (*AskService)(nil)

// NewAskService wires the Ask use case. topN bounds the context items (0 = all).
func NewAskService(engine *signal.Engine, answerer ports.Answerer, input signal.Input, topN int) *AskService {
	return &AskService{engine: engine, answerer: answerer, input: input, topN: topN}
}

// Answer computes the current context and has the answerer respond, grounded.
func (s *AskService) Answer(ctx context.Context, question string) (intel.Answer, error) {
	question = strings.TrimSpace(question)
	if question == "" {
		return intel.Answer{Text: "Please ask a question about the network."}, nil
	}
	if utf8.RuneCountInString(question) > maxQuestionLen {
		question = string([]rune(question)[:maxQuestionLen])
	}
	signals := s.engine.Run(s.input)
	pulse := signal.NetworkPulse(s.input.Facilities, signals)
	c := intel.BuildContext(s.input.AsOf, s.input.Facilities, signals, pulse, s.topN)

	answer, err := s.answerer.Answer(ctx, question, c)
	if err != nil {
		return intel.Answer{}, fmt.Errorf("app: answer question: %w", err)
	}
	// Grounding guardrail: drop any citation the model invented (not a real facility).
	answer.Citations = groundCitations(answer.Citations, knownFacilityIDs(s.input.Facilities))
	return answer, nil
}
