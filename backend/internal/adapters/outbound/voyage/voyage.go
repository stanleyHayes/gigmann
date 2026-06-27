// Package voyage is a ports.Embedder backed by the Voyage AI embeddings REST API
// (https://docs.voyageai.com/reference/embeddings-api). Voyage has no official Go
// SDK, so this calls the REST endpoint directly. The API key is read from config
// (env) and never logged.
package voyage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"time"

	"github.com/xcreativs/gigmann/internal/ports"
)

// Dim is the requested output dimension; it must match the vector(Dim) DB column.
const Dim = 1024

const defaultModel = "voyage-3.5-lite"

// Embedder calls the Voyage AI embeddings API.
type Embedder struct {
	apiKey  string
	model   string
	dim     int
	baseURL string
	httpc   *http.Client
}

var _ ports.Embedder = (*Embedder)(nil)

// NewEmbedder builds a Voyage embedder. An empty model uses the default.
func NewEmbedder(apiKey, model string) *Embedder {
	if model == "" {
		model = defaultModel
	}
	return &Embedder{
		apiKey:  apiKey,
		model:   model,
		dim:     Dim,
		baseURL: "https://api.voyageai.com",
		httpc:   &http.Client{Timeout: 30 * time.Second},
	}
}

// Dimensions returns the requested embedding dimension.
func (e *Embedder) Dimensions() int { return e.dim }

type embedRequest struct {
	Input           []string `json:"input"`
	Model           string   `json:"model"`
	InputType       string   `json:"input_type,omitempty"`
	OutputDimension int      `json:"output_dimension"`
}

type embedResponse struct {
	Data []struct {
		Embedding []float32 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
}

// Embed returns one vector per input text, preserving input order.
func (e *Embedder) Embed(ctx context.Context, texts []string, kind ports.EmbedKind) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, nil
	}
	body, err := json.Marshal(embedRequest{
		Input:           texts,
		Model:           e.model,
		InputType:       string(kind),
		OutputDimension: e.dim,
	})
	if err != nil {
		return nil, fmt.Errorf("voyage: marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, e.baseURL+"/v1/embeddings", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("voyage: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+e.apiKey)

	resp, err := e.httpc.Do(req)
	if err != nil {
		return nil, fmt.Errorf("voyage: do request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		snippet, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("voyage: status %d: %s", resp.StatusCode, bytes.TrimSpace(snippet))
	}

	var parsed embedResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, fmt.Errorf("voyage: decode response: %w", err)
	}
	if len(parsed.Data) != len(texts) {
		return nil, fmt.Errorf("voyage: expected %d embeddings, got %d", len(texts), len(parsed.Data))
	}

	// Order by the API's index field so output aligns with input order.
	sort.Slice(parsed.Data, func(i, j int) bool { return parsed.Data[i].Index < parsed.Data[j].Index })
	out := make([][]float32, len(parsed.Data))
	for i, d := range parsed.Data {
		if len(d.Embedding) != e.dim {
			return nil, fmt.Errorf("voyage: embedding %d has dimension %d, want %d", i, len(d.Embedding), e.dim)
		}
		out[i] = d.Embedding
	}
	return out, nil
}
