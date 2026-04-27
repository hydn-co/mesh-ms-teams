package msgraph_api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var graphErrorDescriptions = map[string]string{
	"Authorization_RequestDenied": "insufficient permissions for this operation",
	"invalid_grant":               "authentication failed; verify credentials",
	"invalid_client":              "invalid client credentials",
	"invalid_scope":               "requested scope is invalid or requires additional permissions",
	"AADSTS50058":                 "silent sign-in request failed; user may need to reauthenticate",
	"AADSTS65001":                 "user or admin has not consented to use the application",
	"itemNotFound":                "resource not found",
	"notAllowedOnUserObject":      "operation not supported on this user object",
	"quota_exceeded":              "request quota exceeded",
}

type ResponseEnvelope struct {
	Error ErrorResponse `json:"error,omitempty"`
}

type ErrorResponse struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	InnerError struct {
		Code string `json:"code"`
	} `json:"innerError,omitempty"`
}

const maxRateLimitRetries = 5

// EnsureContextActive returns early when the provided context has been canceled.
func EnsureContextActive(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("operation canceled: %w", err)
	}
	return nil
}

// NewGraphRequest creates a new HTTP request for Microsoft Graph API calls.
// The request includes the Authorization header with the bearer token.
func NewGraphRequest(ctx context.Context, method, endpoint, token string, payload any) (*http.Request, error) {
	if err := EnsureContextActive(ctx); err != nil {
		return nil, err
	}

	var body io.Reader
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal payload: %w", err)
		}
		body = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	return req, nil
}

// Do executes an HTTP request and decodes the response into the provided interface.
// It handles rate limiting with exponential backoff and retry logic.
func Do(req *http.Request, response any) error {
	if err := EnsureContextActive(req.Context()); err != nil {
		return err
	}

	for attempt := 0; attempt <= maxRateLimitRetries; attempt++ {
		resp, err := doRequestAttempt(req)
		if err != nil {
			if cerr := req.Context().Err(); cerr != nil {
				return fmt.Errorf("operation canceled: %w", cerr)
			}
			return fmt.Errorf("API request failed: %w", err)
		}

		body, closeErr, readErr := readResponseBody(req.Context(), resp)

		if closeErr != nil {
			return fmt.Errorf("failed to close response body: %w", closeErr)
		}

		if readErr != nil {
			return fmt.Errorf("failed to read response body: %w", readErr)
		}

		// Handle rate limiting (429 Too Many Requests)
		if resp.StatusCode == http.StatusTooManyRequests {
			if attempt < maxRateLimitRetries {
				retryAfter := getRetryAfter(resp.Header)
				select {
				case <-time.After(retryAfter):
					continue
				case <-req.Context().Done():
					return fmt.Errorf("operation canceled during backoff: %w", req.Context().Err())
				}
			}
		}

		// Handle successful response
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			if response != nil && len(body) > 0 {
				if err := json.Unmarshal(body, response); err != nil {
					return fmt.Errorf("failed to parse response: %w", err)
				}
			}
			return nil
		}

		// Handle error response
		return parseErrorResponse(resp.StatusCode, body)
	}

	return fmt.Errorf("max retry attempts exceeded")
}

// doRequestAttempt sends a single HTTP request with client timeout.
func doRequestAttempt(req *http.Request) (*http.Response, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	return client.Do(req)
}

// readResponseBody reads and returns the response body.
func readResponseBody(ctx context.Context, resp *http.Response) ([]byte, error, error) {
	if err := EnsureContextActive(ctx); err != nil {
		return nil, resp.Body.Close(), err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.Body.Close(), fmt.Errorf("failed to read body: %w", err)
	}

	return body, nil, nil
}

// getRetryAfter extracts the retry-after duration from response headers.
// Defaults to exponential backoff based on attempt count.
func getRetryAfter(headers http.Header) time.Duration {
	if retryAfter := headers.Get("Retry-After"); retryAfter != "" {
		if seconds, err := strconv.Atoi(retryAfter); err == nil {
			return time.Duration(seconds) * time.Second
		}
	}
	// Exponential backoff: 1s → 2s → 4s → 8s → 16s
	return time.Duration(1<<(uint(2))) * time.Second // Default to 4s
}

// parseErrorResponse converts an API error response into a descriptive error.
func parseErrorResponse(statusCode int, body []byte) error {
	var errResp ResponseEnvelope
	if err := json.Unmarshal(body, &errResp); err != nil {
		return fmt.Errorf("API request failed with status %d (failed to parse error)", statusCode)
	}

	if errResp.Error.Code == "" {
		return fmt.Errorf("API request failed with status %d", statusCode)
	}

	description := errResp.Error.Message
	if desc, exists := graphErrorDescriptions[errResp.Error.Code]; exists {
		description = desc
	}

	return fmt.Errorf("microsoft Graph API error [%s]: %s (status %d)",
		errResp.Error.Code, description, statusCode)
}

// BuildURL constructs a Microsoft Graph API URL with query parameters.
func BuildURL(endpoint string, params map[string]string) string {
	if len(params) == 0 {
		return endpoint
	}

	var query []string
	for key, value := range params {
		query = append(query, fmt.Sprintf("$%s=%s", key, value))
	}

	return endpoint + "?" + strings.Join(query, "&")
}
