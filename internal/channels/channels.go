package channels

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/hydn-co/mesh-ms-teams/internal/endpoints"
	"github.com/hydn-co/mesh-ms-teams/internal/msgraph_api"
)

// GraphChannel represents a Microsoft Teams channel resource from the Microsoft Graph API.
type GraphChannel struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
	Description string `json:"description"`
}

// ListChannelsResult wraps the list of channels returned from the Microsoft Graph API.
type ListChannelsResult struct {
	OdataNextLink string         `json:"@odata.nextLink"`
	Value         []GraphChannel `json:"value"`
}

// SendMessagePayload represents the payload for sending a message to a channel.
type SendMessagePayload struct {
	Body struct {
		Content string `json:"content"`
	} `json:"body"`
}

// ListChannels retrieves all channels from a specific team.
func ListChannels(ctx context.Context, token, teamID string) (*ListChannelsResult, error) {
	if err := msgraph_api.EnsureContextActive(ctx); err != nil {
		return nil, err
	}

	if teamID == "" {
		return nil, fmt.Errorf("team_id cannot be empty")
	}

	endpoint := strings.ReplaceAll(endpoints.GraphChannelsListEndpoint, "{teamId}", teamID)

	req, err := msgraph_api.NewGraphRequest(ctx, http.MethodGet, endpoint, token, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create channels list request: %w", err)
	}

	var result ListChannelsResult
	if err := msgraph_api.Do(req, &result); err != nil {
		return nil, fmt.Errorf("failed to list channels for team %s: %w", teamID, err)
	}

	return &result, nil
}

// ListChannelsPage retrieves a page of channels using the provided URL.
func ListChannelsPage(ctx context.Context, token, pageURL string) (*ListChannelsResult, error) {
	if err := msgraph_api.EnsureContextActive(ctx); err != nil {
		return nil, err
	}

	req, err := msgraph_api.NewGraphRequest(ctx, http.MethodGet, pageURL, token, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create channels page request: %w", err)
	}

	var result ListChannelsResult
	if err := msgraph_api.Do(req, &result); err != nil {
		return nil, fmt.Errorf("failed to get channels page: %w", err)
	}

	return &result, nil
}

// SendMessage posts a message to a Teams channel.
func SendMessage(ctx context.Context, token, teamID, channelID, message string) error {
	if err := msgraph_api.EnsureContextActive(ctx); err != nil {
		return err
	}

	if teamID == "" {
		return fmt.Errorf("team_id cannot be empty")
	}
	if channelID == "" {
		return fmt.Errorf("channel_id cannot be empty")
	}
	if message == "" {
		return fmt.Errorf("message cannot be empty")
	}

	endpoint := strings.NewReplacer(
		"{teamId}", teamID,
		"{channelId}", channelID,
	).Replace(endpoints.GraphSendMessageEndpoint)

	payload := SendMessagePayload{}
	payload.Body.Content = message

	req, err := msgraph_api.NewGraphRequest(ctx, http.MethodPost, endpoint, token, payload)
	if err != nil {
		return fmt.Errorf("failed to create send message request: %w", err)
	}

	if err := msgraph_api.Do(req, nil); err != nil {
		return fmt.Errorf("failed to send message to team %s channel %s: %w", teamID, channelID, err)
	}

	return nil
}
