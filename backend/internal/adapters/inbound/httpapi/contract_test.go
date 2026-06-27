package httpapi_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/xcreativs/gigmann/internal/adapters/inbound/httpapi"
)

// TestOpenAPISpecIsValid guards the published API contract: the embedded spec must
// be a well-formed OpenAPI document (this plus the CI codegen-drift job — which
// regenerates the server/client from the spec — keeps contract and code in lock-step).
func TestOpenAPISpecIsValid(t *testing.T) {
	spec, err := httpapi.GetSwagger()
	require.NoError(t, err)
	require.NotNil(t, spec)
	require.NoError(t, spec.Validate(context.Background()))
}

// TestOpenAPISpecCoversCoreRoutes asserts the contract declares the hero surfaces.
func TestOpenAPISpecCoversCoreRoutes(t *testing.T) {
	spec, err := httpapi.GetSwagger()
	require.NoError(t, err)
	for _, path := range []string{
		"/api/v1/brief", "/api/v1/metrics", "/api/v1/facilities", "/api/v1/auth/login",
		"/api/v1/approvals", "/api/v1/tasks", "/api/v1/ask", "/api/v1/alerts",
		"/api/v1/facilities/search", "/api/v1/me/preferences",
	} {
		assert.NotNilf(t, spec.Paths.Find(path), "spec must declare %s", path)
	}
}
