package app

import (
	"context"
	"fmt"
	"strings"

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

var _ ports.QuestionAnswerer = (*AskService)(nil)

// NewAskService wires the Ask use case. topN bounds the context items (0 = all).
func NewAskService(engine *signal.Engine, answerer ports.Answerer, input signal.Input, topN int) *AskService {
	return &AskService{engine: engine, answerer: answerer, input: input, topN: topN}
}

// Answer computes the current context and has the answerer respond, grounded.
func (s *AskService) Answer(ctx context.Context, question string) (intel.Answer, error) {
	if strings.TrimSpace(question) == "" {
		return intel.Answer{Text: "Please ask a question about the network."}, nil
	}
	signals := s.engine.Run(s.input)
	pulse := signal.NetworkPulse(s.input.Facilities, signals)
	c := intel.BuildContext(s.input.AsOf, s.input.Facilities, signals, pulse, s.topN)

	answer, err := s.answerer.Answer(ctx, question, c)
	if err != nil {
		return intel.Answer{}, fmt.Errorf("app: answer question: %w", err)
	}
	return answer, nil
}
