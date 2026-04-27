package teams

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hydn-co/mesh-ms-teams/internal/endpoints"
	"github.com/hydn-co/mesh-ms-teams/internal/msgraph_api"
)

// GraphTeam represents a Microsoft Teams team resource from the Microsoft Graph API.
type GraphTeam struct {
	ID           string `json:"id"`
	DisplayName  string `json:"displayName"`
	Description  string `json:"description"`
	MailNickname string `json:"mailNickname"`
}

// ListTeamsResult wraps the list of teams returned from the Microsoft Graph API.
type ListTeamsResult struct {
	Value         []GraphTeam `json:"value"`
	OdataNextLink string      `json:"@odata.nextLink"`
}

// ListTeams retrieves the first page of teams accessible to the service principal
// from Microsoft Graph. Use ListTeamsPage with the returned OdataNextLink to paginate.
func ListTeams(ctx context.Context, token string) (*ListTeamsResult, error) {
	if err := msgraph_api.EnsureContextActive(ctx); err != nil {
		return nil, err
	}

	req, err := msgraph_api.NewGraphRequest(ctx, http.MethodGet, endpoints.GraphTeamsListEndpoint, token, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create teams list request: %w", err)
	}

	var result ListTeamsResult
	if err := msgraph_api.Do(req, &result); err != nil {
		return nil, fmt.Errorf("failed to list teams: %w", err)
	}

	return &result, nil
}

// ListTeamsPage retrieves a page of teams using the provided URL.
// The URL typically comes from the @odata.nextLink field in a previous response.
func ListTeamsPage(ctx context.Context, token, pageURL string) (*ListTeamsResult, error) {
	if err := msgraph_api.EnsureContextActive(ctx); err != nil {
		return nil, err
	}

	req, err := msgraph_api.NewGraphRequest(ctx, http.MethodGet, pageURL, token, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create teams page request: %w", err)
	}

	var result ListTeamsResult
	if err := msgraph_api.Do(req, &result); err != nil {
		return nil, fmt.Errorf("failed to get teams page: %w", err)
	}

	return &result, nil
}
