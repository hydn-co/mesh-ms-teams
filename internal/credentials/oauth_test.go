package credentials_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hydn-co/mesh-sdk/pkg/connectorutil"

	"github.com/hydn-co/mesh-ms-teams/internal/credentials"
)

func TestShouldParseCredentialsWhenValidJSON(t *testing.T) {
	// Arrange
	raw := json.RawMessage(`{"client_id":"def456","client_secret":"s3cr3t"}`)

	// Act
	creds, err := credentials.ParseCredentials(
		map[string]json.RawMessage{connectorutil.DefaultCredentialName: raw}, "abc123")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "abc123", creds.TenantID)
	assert.Equal(t, "def456", creds.ClientID)
	assert.Equal(t, "s3cr3t", creds.ClientSecret)
}

func TestShouldReturnErrorWhenCredentialsEmpty(t *testing.T) {
	// Arrange
	raw := json.RawMessage{}

	// Act
	_, err := credentials.ParseCredentials(
		map[string]json.RawMessage{connectorutil.DefaultCredentialName: raw}, "abc123")

	// Assert
	assert.ErrorContains(t, err, "no credentials")
}

func TestShouldReturnErrorWhenJSONInvalid(t *testing.T) {
	// Arrange
	raw := json.RawMessage(`{not valid json`)

	// Act
	_, err := credentials.ParseCredentials(
		map[string]json.RawMessage{connectorutil.DefaultCredentialName: raw}, "abc123")

	// Assert
	assert.Error(t, err)
}

func TestShouldReturnErrorWhenTenantIDMissing(t *testing.T) {
	// Arrange
	raw := json.RawMessage(`{"client_id":"id","client_secret":"s"}`)

	// Act
	_, err := credentials.ParseCredentials(
		map[string]json.RawMessage{connectorutil.DefaultCredentialName: raw}, "")

	// Assert
	assert.ErrorContains(t, err, "tenant_id")
}

func TestShouldReturnErrorWhenClientIDMissing(t *testing.T) {
	// Arrange
	raw := json.RawMessage(`{"client_secret":"s"}`)

	// Act
	_, err := credentials.ParseCredentials(
		map[string]json.RawMessage{connectorutil.DefaultCredentialName: raw}, "t1")

	// Assert
	assert.ErrorContains(t, err, "client_id")
}

func TestShouldReturnErrorWhenClientSecretMissing(t *testing.T) {
	// Arrange
	raw := json.RawMessage(`{"client_id":"id"}`)

	// Act
	_, err := credentials.ParseCredentials(
		map[string]json.RawMessage{connectorutil.DefaultCredentialName: raw}, "t1")

	// Assert
	assert.ErrorContains(t, err, "client_secret")
}

func TestShouldValidateSuccessfullyWhenAllFieldsPresent(t *testing.T) {
	// Arrange
	creds := &credentials.AzureADCredentials{TenantID: "t1", ClientID: "c1", ClientSecret: "s1"}

	// Act
	err := creds.Validate()

	// Assert
	assert.NoError(t, err)
}

func TestShouldReturnAccessTokenWhenOAuthSucceeds(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, r.ParseForm())
		assert.Equal(t, "client_credentials", r.FormValue("grant_type"))
		assert.Equal(t, "https://graph.microsoft.com/.default", r.FormValue("scope"))
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"access_token": "test-token", "expires_in": 3600})
	}))
	defer server.Close()
	creds := &credentials.AzureADCredentials{TenantID: "t", ClientID: "c", ClientSecret: "s"}

	// Act
	token, err := creds.GetAccessTokenFromURL(context.Background(), server.URL)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "test-token", token)
}

func TestShouldReturnErrorWhenOAuthReturnsAuthError(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error":             "invalid_client",
			"error_description": "The client secret is invalid.",
		})
	}))
	defer server.Close()
	creds := &credentials.AzureADCredentials{TenantID: "t", ClientID: "c", ClientSecret: "wrong"}

	// Act
	_, err := creds.GetAccessTokenFromURL(context.Background(), server.URL)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid_client")
}

func TestShouldReturnErrorWhenOAuthResponseHasNoToken(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"expires_in": 3600})
	}))
	defer server.Close()
	creds := &credentials.AzureADCredentials{TenantID: "t", ClientID: "c", ClientSecret: "s"}

	// Act
	_, err := creds.GetAccessTokenFromURL(context.Background(), server.URL)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no access token")
}

func TestShouldReturnErrorWhenContextCanceledDuringOAuth(t *testing.T) {
	// Arrange
	creds := &credentials.AzureADCredentials{TenantID: "t", ClientID: "c", ClientSecret: "s"}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Act
	_, err := creds.GetAccessTokenFromURL(ctx, "http://"+strings.Repeat("x", 200)+".invalid")

	// Assert
	assert.Error(t, err)
}
