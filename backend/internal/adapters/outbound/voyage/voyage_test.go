package voyage_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/xcreativs/gigmann/internal/adapters/outbound/voyage"
	"github.com/xcreativs/gigmann/internal/ports"
)

func TestEmbedSendsRequestAndParsesOrderedResponse(t *testing.T) {
	var gotAuth, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		// Return out-of-order indices to prove we re-sort by index.
		mk := func(v float32) []float32 {
			out := make([]float32, voyage.Dim)
			out[0] = v
			return out
		}
		resp := map[string]any{"data": []map[string]any{
			{"embedding": mk(2.0), "index": 1},
			{"embedding": mk(1.0), "index": 0},
		}}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	e := voyage.NewEmbedder("secret-key", "voyage-3.5-lite")
	// Point at the test server via the unexported baseURL through the exported hook.
	voyage.SetBaseURLForTest(e, srv.URL)

	vecs, err := e.Embed(context.Background(), []string{"alpha", "beta"}, ports.EmbedQuery)
	require.NoError(t, err)
	require.Len(t, vecs, 2)
	assert.InDelta(t, 1.0, vecs[0][0], 1e-6, "index 0 first after re-sort")
	assert.InDelta(t, 2.0, vecs[1][0], 1e-6, "index 1 second")

	assert.Equal(t, "Bearer secret-key", gotAuth)
	assert.Contains(t, gotBody, `"input_type":"query"`)
	assert.Contains(t, gotBody, `"output_dimension":1024`)
	assert.Contains(t, gotBody, `"voyage-3.5-lite"`)
}

func TestEmbedNonOKStatusErrors(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"detail":"invalid key"}`))
	}))
	defer srv.Close()

	e := voyage.NewEmbedder("bad", "")
	voyage.SetBaseURLForTest(e, srv.URL)
	_, err := e.Embed(context.Background(), []string{"x"}, ports.EmbedDocument)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "401")
}
