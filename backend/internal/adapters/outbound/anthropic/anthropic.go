// Package anthropic is the outbound adapter that narrates the Daily Brief via
// Claude. The deterministic figures come from the signal engine; Claude only
// prioritises and writes them in plain language (spec §6.1). Network-backed, so
// it is excluded from the unit coverage gate; the pure parse path is unit-tested.
package anthropic

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"

	"github.com/xcreativs/gigmann/internal/core/brief"
	"github.com/xcreativs/gigmann/internal/core/severity"
	"github.com/xcreativs/gigmann/internal/intel"
	"github.com/xcreativs/gigmann/internal/ports"
)

const (
	systemPrompt = "You are Sammy Adjei's chief of staff for Gigmann Medicals, a hospital network in Ghana. " +
		"Write his Daily Brief from the supplied signals. Speak in plain English, in cedis, about NHIS, MoMo, and his facilities. " +
		"Lead with the worst item first and connect cause to effect only where the figures support it. " +
		"CRITICAL: use only the figures provided in the context — never invent or estimate numbers. " +
		"Return the brief via the emit_brief tool."
	maxTokens = 4096
	toolName  = "emit_brief"
)

// Narrator narrates briefs using the Anthropic Messages API.
type Narrator struct {
	client anthropic.Client
	model  string
}

// Compile-time guarantee that Narrator satisfies the port.
var _ ports.Narrator = (*Narrator)(nil)

// NewNarrator builds a Narrator. model defaults to claude-sonnet-4-6 when empty.
func NewNarrator(apiKey, model string) *Narrator {
	if model == "" {
		model = anthropic.ModelClaudeSonnet4_6
	}
	return &Narrator{
		client: anthropic.NewClient(option.WithAPIKey(apiKey)),
		model:  model,
	}
}

// NarrateBrief calls Claude with the computed context and a strict emit_brief tool.
func (n *Narrator) NarrateBrief(ctx context.Context, c intel.Context) (brief.Brief, error) {
	ctxJSON, err := json.Marshal(c)
	if err != nil {
		return brief.Brief{}, fmt.Errorf("anthropic: marshal context: %w", err)
	}

	tool := anthropic.ToolParam{
		Name:        toolName,
		Description: anthropic.String("Emit the prioritised Daily Brief."),
		InputSchema: anthropic.ToolInputSchemaParam{
			Properties: briefSchemaProperties(),
			ExtraFields: map[string]any{
				"required":             []string{"prose", "items"},
				"additionalProperties": false,
			},
		},
		Strict: anthropic.Bool(true),
	}

	resp, err := n.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:      n.model,
		MaxTokens:  maxTokens,
		System:     []anthropic.TextBlockParam{{Text: systemPrompt}},
		Tools:      []anthropic.ToolUnionParam{{OfTool: &tool}},
		ToolChoice: anthropic.ToolChoiceParamOfTool(toolName),
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(
				"Today's computed network context (JSON):\n" + string(ctxJSON))),
		},
	})
	if err != nil {
		return brief.Brief{}, fmt.Errorf("anthropic: messages: %w", err)
	}

	meta := briefMeta{
		id:          "brief-" + c.Date.Format(time.DateOnly),
		date:        c.Date,
		generatedAt: time.Now().UTC(),
		model:       n.model,
	}
	for _, block := range resp.Content {
		if tu, ok := block.AsAny().(anthropic.ToolUseBlock); ok && tu.Name == toolName {
			return parseBrief(meta, []byte(tu.JSON.Input.Raw()))
		}
	}
	return brief.Brief{}, fmt.Errorf("anthropic: response had no %s tool call", toolName)
}

func briefSchemaProperties() map[string]any {
	return map[string]any{
		"prose": map[string]any{
			"type":        "string",
			"description": "Short prose brief that greets Sammy by name.",
		},
		"items": map[string]any{
			"type": "array",
			"items": map[string]any{
				"type":                 "object",
				"additionalProperties": false,
				"required":             []string{"severity", "facility_id", "headline"},
				"properties": map[string]any{
					"severity":          map[string]any{"type": "string", "enum": []string{"good", "watch", "critical"}},
					"facility_id":       map[string]any{"type": "string"},
					"headline":          map[string]any{"type": "string"},
					"explanation":       map[string]any{"type": "string"},
					"suggested_actions": map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
				},
			},
		},
	}
}

type briefMeta struct {
	id          string
	model       string
	date        time.Time
	generatedAt time.Time
}

type briefDTO struct {
	Prose string `json:"prose"`
	Items []struct {
		Severity         string   `json:"severity"`
		FacilityID       string   `json:"facility_id"`
		Headline         string   `json:"headline"`
		Explanation      string   `json:"explanation"`
		SuggestedActions []string `json:"suggested_actions"`
	} `json:"items"`
}

// parseBrief turns the model's tool-call JSON into a validated domain Brief.
func parseBrief(meta briefMeta, raw []byte) (brief.Brief, error) {
	var dto briefDTO
	if err := json.Unmarshal(raw, &dto); err != nil {
		return brief.Brief{}, fmt.Errorf("anthropic: parse brief json: %w", err)
	}
	items := make([]brief.Item, 0, len(dto.Items))
	for _, it := range dto.Items {
		items = append(items, brief.Item{
			Severity:         severity.Severity(it.Severity),
			FacilityID:       it.FacilityID,
			Headline:         it.Headline,
			Explanation:      it.Explanation,
			SuggestedActions: it.SuggestedActions,
		})
	}
	return brief.New(brief.Brief{
		ID: meta.id, Date: meta.date, Prose: dto.Prose, Items: items,
		GeneratedAt: meta.generatedAt, Model: meta.model,
	})
}
