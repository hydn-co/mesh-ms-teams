package users

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hydn-co/mesh-ms-teams/internal/endpoints"
	"github.com/hydn-co/mesh-ms-teams/internal/msgraph_api"
)

// GraphUser represents a Microsoft Teams / Entra ID user from the Microsoft Graph API.
type GraphUser struct {
	ID                string `json:"id"`
	DisplayName       string `json:"displayName"`
	GivenName         string `json:"givenName"`
	Surname           string `json:"surname"`
	UserPrincipalName string `json:"userPrincipalName"`
	Mail              string `json:"mail"`
	AccountEnabled    bool   `json:"accountEnabled"`
	UserType          string `json:"userType"` // "Member" or "Guest"
}

// ListUsersResult wraps the list of users returned from the Microsoft Graph API.
type ListUsersResult struct {
	Value         []GraphUser `json:"value"`
	OdataNextLink string      `json:"@odata.nextLink"`
}

// ProvisionUserRequest represents the payload for creating a new user via the Microsoft Graph API.
type ProvisionUserRequest struct {
	AccountEnabled    bool            `json:"accountEnabled"`
	DisplayName       string          `json:"displayName"`
	MailNickname      string          `json:"mailNickname"`
	UserPrincipalName string          `json:"userPrincipalName"`
	PasswordProfile   PasswordProfile `json:"passwordProfile"`
}

// PasswordProfile holds the initial password configuration for a new user.
type PasswordProfile struct {
	ForceChangePasswordNextSignIn bool   `json:"forceChangePasswordNextSignIn"`
	Password                      string `json:"password"`
}

// ListUsers retrieves all users from the Microsoft Graph API.
func ListUsers(ctx context.Context, token string) (*ListUsersResult, error) {
	if err := msgraph_api.EnsureContextActive(ctx); err != nil {
		return nil, err
	}

	req, err := msgraph_api.NewGraphRequest(ctx, http.MethodGet, endpoints.GraphUsersListEndpoint, token, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create users list request: %w", err)
	}

	var result ListUsersResult
	if err := msgraph_api.Do(req, &result); err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	return &result, nil
}

// ListUsersPage retrieves a page of users using the provided URL.
// The URL typically comes from the @odata.nextLink field in a previous response.
func ListUsersPage(ctx context.Context, token, pageURL string) (*ListUsersResult, error) {
	if err := msgraph_api.EnsureContextActive(ctx); err != nil {
		return nil, err
	}

	req, err := msgraph_api.NewGraphRequest(ctx, http.MethodGet, pageURL, token, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create users page request: %w", err)
	}

	var result ListUsersResult
	if err := msgraph_api.Do(req, &result); err != nil {
		return nil, fmt.Errorf("failed to get users page: %w", err)
	}

	return &result, nil
}

// ProvisionUser creates a new user in Entra ID via the Microsoft Graph API.
func ProvisionUser(ctx context.Context, token string, payload ProvisionUserRequest) error {
	if err := msgraph_api.EnsureContextActive(ctx); err != nil {
		return err
	}

	req, err := msgraph_api.NewGraphRequest(ctx, http.MethodPost, endpoints.GraphProvisionUserEndpoint, token, payload)
	if err != nil {
		return fmt.Errorf("failed to create provision user request: %w", err)
	}

	if err := msgraph_api.Do(req, nil); err != nil {
		return fmt.Errorf("failed to provision user %q: %w", payload.UserPrincipalName, err)
	}

	return nil
}
