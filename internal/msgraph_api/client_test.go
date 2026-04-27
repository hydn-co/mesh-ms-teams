package msgraph_api_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hydn-co/mesh-ms-teams/internal/msgraph_api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShouldReturnNilWhenContextIsActive(t *testing.T) {
	// Arrange
	ctx := context.Background()

	// Act
	err := msgraph_api.EnsureContextActive(ctx)

	// Assert
	assert.NoError(t, err)
}

func TestShouldReturnErrorWhenContextIsCanceled(t *testing.T) {
	// Arrange
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Act
	err := msgraph_api.EnsureContextActive(ctx)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "canceled")
}

func TestShouldSetAuthorizationHeaderWhenNewGraphRequest(t *testing.T) {
	// Arrange
	ctx := context.Background()

	// Act
	req, err := msgraph_api.NewGraphRequest(ctx, http.MethodGet, "https://example.com", "mytoken", nil)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "Bearer mytoken", req.Header.Get("Authorization"))
}

func TestShouldSetJSONContentTypeWhenNewGraphRequest(t *testing.T) {
	// Arrange
	ctx := context.Background()

	// Act
	req, err := msgraph_api.NewGraphRequest(ctx, http.MethodGet, "https://example.com", "token", nil)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
}

func TestShouldHaveBodyWhenNewGraphRequestWithPayload(t *testing.T) {
	// Arrange
	payload := map[string]string{"key": "value"}

	// Act
	req, err := msgraph_api.NewGraphRequest(
		context.Background(),
		http.MethodPost,
		"https://example.com",
		"token",
		payload,
	)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, req.Body)
}

func TestShouldDecodeBodyWhenDoSucceeds(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"key": "value"})
	}))
	defer server.Close()
	req, err := msgraph_api.NewGraphRequest(context.Background(), http.MethodGet, server.URL, "token", nil)
	require.NoError(t, err)

	// Act
	var result map[string]string
	err = msgraph_api.Do(req, &result)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "value", result["key"])
}

func TestShouldReturnErrorWhenDo401Response(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]string{
				"code":    "invalid_grant",
				"message": "The token is invalid",
			},
		})
	}))
	defer server.Close()
	req, err := msgraph_api.NewGraphRequest(context.Background(), http.MethodGet, server.URL, "badtoken", nil)
	require.NoError(t, err)

	// Act
	err = msgraph_api.Do(req, nil)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid_grant")
}

func TestShouldReturnErrorWhenDo404Response(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]string{
				"code":    "itemNotFound",
				"message": "The item was not found",
			},
		})
	}))
	defer server.Close()
	req, err := msgraph_api.NewGraphRequest(context.Background(), http.MethodGet, server.URL, "token", nil)
	require.NoError(t, err)

	// Act
	err = msgraph_api.Do(req, nil)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "itemNotFound")
}
