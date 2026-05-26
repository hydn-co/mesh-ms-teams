package msgraph_api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/hydn-co/mesh-sdk/pkg/connectorutil"
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
	return newGraphRequest(ctx, method, endpoint, token, payload, false)
}

// NewGraphAdvancedRequest creates a Graph request with the ConsistencyLevel: eventual header,
// required for advanced OData queries (lambda operators, $count on directory objects, etc.).
func NewGraphAdvancedRequest(ctx context.Context, method, endpoint, token string, payload any) (*http.Request, error) {
	return newGraphRequest(ctx, method, endpoint, token, payload, true)
}

func newGraphRequest(
	ctx context.Context,
	method, endpoint, token string,
	payload any,
	advanced bool,
) (*http.Request, error) {
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
	if advanced {
		// Required for advanced OData filters on directory objects (lambda operators, $search, etc.)
		req.Header.Set("ConsistencyLevel", "eventual")
	}

	return req, nil
}

// Do executes an HTTP request and decodes the response into the provided interface.
// It handles rate limiting with exponential backoff and retry logic.
func Do(req *http.Request, response any) error {
	if err := EnsureContextActive(req.Context()); err != nil {
		return err
	}

	return connectorutil.RetryOperation(req.Context(), connectorutil.RetryPolicy{
		ShouldRetry: isRateLimitError,
		BaseDelay:   1 * time.Second,
		MaxDelay:    60 * time.Second,
		MaxRetries:  maxRateLimitRetries,
	}, func(stepCtx context.Context) (connectorutil.RetryOperationResult, error) {
		resp, err := doRequestAttempt(req)
		if err != nil {
			if cerr := req.Context().Err(); cerr != nil {
				return connectorutil.RetryOperationResult{}, fmt.Errorf("operation canceled: %w", cerr)
			}
			return connectorutil.RetryOperationResult{}, fmt.Errorf("API request failed: %w", err)
		}

		body, closeErr, readErr := readResponseBody(stepCtx, resp)

		if closeErr != nil {
			return connectorutil.RetryOperationResult{}, fmt.Errorf("failed to close response body: %w", closeErr)
		}

		if readErr != nil {
			return connectorutil.RetryOperationResult{}, fmt.Errorf("failed to read response body: %w", readErr)
		}

		// Handle successful response
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			if response != nil && len(body) > 0 {
				if err := json.Unmarshal(body, response); err != nil {
					return connectorutil.RetryOperationResult{}, fmt.Errorf("failed to parse response: %w", err)
				}
			}
			return connectorutil.RetryOperationResult{}, nil
		}

		if resp.StatusCode == http.StatusTooManyRequests {
			return connectorutil.RetryOperationResult{
					RetryAfter: getRetryAfter(resp.Header, 0),
				}, connectorutil.NewRetryableStatusError(
					resp.StatusCode,
					fmt.Sprintf("API request failed with status %d", resp.StatusCode),
					nil,
				)
		}

		// Handle error response
		return connectorutil.RetryOperationResult{}, parseErrorResponse(resp.StatusCode, body)
	})
}

// doRequestAttempt sends a single HTTP request with client timeout.
func doRequestAttempt(req *http.Request) (*http.Response, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	return client.Do(req)
}

// readResponseBody reads the response body and closes it.
func readResponseBody(ctx context.Context, resp *http.Response) ([]byte, error, error) {
	defer func() {
		_ = resp.Body.Close()
	}()

	if err := EnsureContextActive(ctx); err != nil {
		return nil, nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read body: %w", err)
	}

	return body, nil, nil
}

// getRetryAfter extracts the retry-after duration from response headers.
// Falls back to exponential backoff based on attempt: 1s → 2s → 4s → 8s → 16s.
func getRetryAfter(headers http.Header, attempt int) time.Duration {
	if retryAfter := headers.Get("Retry-After"); retryAfter != "" {
		if seconds, err := strconv.Atoi(retryAfter); err == nil {
			return time.Duration(seconds) * time.Second
		}
	}
	return time.Duration(1<<uint(attempt)) * time.Second
}

func isRateLimitError(err error) bool {
	if err == nil {
		return false
	}

	var statusErr interface{ StatusCode() int }
	if errors.As(err, &statusErr) {
		return statusErr.StatusCode() == http.StatusTooManyRequests
	}

	return false
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

// BuildURL constructs a Microsoft Graph API URL with properly URL-encoded query parameters.
func BuildURL(endpoint string, params map[string]string) string {
	if len(params) == 0 {
		return endpoint
	}

	query := url.Values{}
	for key, value := range params {
		query.Set("$"+key, value)
	}

	return endpoint + "?" + query.Encode()
}
