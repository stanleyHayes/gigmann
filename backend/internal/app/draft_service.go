package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/xcreativs/gigmann/internal/ports"
)

// DraftService generates AI-drafted messages/summaries grounded in the network's
// computed figures. Drafts are **read-only**: they are returned for the executive
// to review and send — the AI never sends anything (CLAUDE.md §7).
type DraftService struct {
	answerer ports.QuestionAnswerer
}

// NewDraftService wires the draft use case over the grounded answerer.
func NewDraftService(answerer ports.QuestionAnswerer) *DraftService {
	return &DraftService{answerer: answerer}
}

// Draft returns a grounded draft for the given kind/instruction (and optional
// facility). An empty instruction yields an empty draft.
func (s *DraftService) Draft(ctx context.Context, kind, facilityID, instruction string) (string, error) {
	instruction = strings.TrimSpace(instruction)
	if instruction == "" {
		return "", nil
	}
	answer, err := s.answerer.Answer(ctx, buildDraftPrompt(kind, facilityID, instruction))
	if err != nil {
		return "", fmt.Errorf("app: generate draft: %w", err)
	}
	return answer.Text, nil
}

func buildDraftPrompt(kind, facilityID, instruction string) string {
	var b strings.Builder
	if kind == "summary" {
		b.WriteString("Write a brief executive summary")
	} else {
		b.WriteString("Draft a short, professional message")
	}
	if facilityID != "" {
		b.WriteString(" concerning the ")
		b.WriteString(facilityID)
		b.WriteString(" facility")
	}
	b.WriteString(" regarding: ")
	b.WriteString(instruction)
	b.WriteString(". Use only the network's known figures; never invent numbers.")
	return b.String()
}
