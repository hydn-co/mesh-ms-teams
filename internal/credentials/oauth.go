package credentials

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// AzureADCredentials holds the OAuth 2.0 client credentials for the Microsoft Graph API.
type AzureADCredentials struct {
	TenantID     string `json:"tenant_id"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	Error       string `json:"error"`
	Description string `json:"error_description"`
}

// ParseCredentials deserializes client_id and client_secret from raw JSON credentials.
// The tenantID is supplied from feature options, not from the credentials secret,
// aligning with the GrantCredential template which only stores client_id and client_secret.
func ParseCredentials(raw json.RawMessage, tenantID string) (*AzureADCredentials, error) {
	if len(raw) == 0 {
		return nil, fmt.Errorf("no credentials provided")
	}

	var creds AzureADCredentials
	if err := json.Unmarshal(raw, &creds); err != nil {
		return nil, fmt.Errorf("invalid credentials JSON format: %w", err)
	}

	creds.TenantID = tenantID

	if err := creds.Validate(); err != nil {
		return nil, err
	}

	return &creds, nil
}

// Validate ensures all required credential fields are present.
func (c *AzureADCredentials) Validate() error {
	if c.TenantID == "" {
		return fmt.Errorf("missing tenant_id: ensure tenant_id is set in feature options")
	}
	if c.ClientID == "" {
		return fmt.Errorf("missing client_id credential field")
	}
	if c.ClientSecret == "" {
		return fmt.Errorf("missing client_secret credential field")
	}
	return nil
}

// GetAccessToken exchanges client credentials for a Microsoft Graph access token
// using the OAuth 2.0 client credentials flow.
func (c *AzureADCredentials) GetAccessToken(ctx context.Context) (string, error) {
	tokenURL := fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/token", c.TenantID)
	return c.GetAccessTokenFromURL(ctx, tokenURL)
}

// GetAccessTokenFromURL is the testable core of GetAccessToken. It accepts an
// explicit token endpoint URL, allowing tests to point at a local httptest server.
func (c *AzureADCredentials) GetAccessTokenFromURL(ctx context.Context, tokenURL string) (string, error) {

	body := url.Values{
		"grant_type":    {"client_credentials"},
		"client_id":     {c.ClientID},
		"client_secret": {c.ClientSecret},
		"scope":         {"https://graph.microsoft.com/.default"},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(body.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to create token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("token request failed: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read token response: %w", err)
	}

	var result tokenResponse
	if err := json.Unmarshal(raw, &result); err != nil {
		return "", fmt.Errorf("failed to parse token response: %w", err)
	}

	if result.Error != "" {
		return "", fmt.Errorf("token request denied [%s]: %s", result.Error, result.Description)
	}

	if result.AccessToken == "" {
		return "", fmt.Errorf("token response contained no access token")
	}

	return result.AccessToken, nil
}
