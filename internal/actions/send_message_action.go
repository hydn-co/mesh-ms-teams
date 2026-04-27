package actions

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/hydn-co/mesh-ms-teams/internal/channels"
	"github.com/hydn-co/mesh-ms-teams/internal/credentials"
	"github.com/hydn-co/mesh-ms-teams/internal/helpers"
	"github.com/hydn-co/mesh-ms-teams/internal/msgraph_api"
	"github.com/hydn-co/mesh-ms-teams/internal/options"
	"github.com/hydn-co/mesh-ms-teams/internal/payloads"
	"github.com/hydn-co/mesh-sdk/pkg/connector"
	"github.com/hydn-co/mesh-sdk/pkg/runner"
)

// SendMessageAction posts messages to Microsoft Teams channels.
type SendMessageAction struct {
	*connector.TypedFeatureContext[*options.SendMessageActionOptions, *payloads.SendMessagePayload]
	token       string
	teamID      string
	channelID   string
	initialized bool
}

// NewSendMessageAction constructs a SendMessageAction.
func NewSendMessageAction(
	ctx *connector.TypedFeatureContext[*options.SendMessageActionOptions, *payloads.SendMessagePayload],
) runner.Feature {
	return &SendMessageAction{TypedFeatureContext: ctx}
}

// Init prepares the action for operation by validating credentials, options, and payload.
func (a *SendMessageAction) Init(ctx context.Context) error {
	if err := msgraph_api.EnsureContextActive(ctx); err != nil {
		return err
	}

	opts := a.GetOptions()
	if opts == nil {
		logAction(ctx, a.TypedFeatureContext, slog.LevelError, "options are required")
		return fmt.Errorf("options are required")
	}
	if opts.TeamID == "" {
		logAction(ctx, a.TypedFeatureContext, slog.LevelError, "team_id is required in options")
		return fmt.Errorf("team_id is required in options")
	}
	if opts.ChannelID == "" {
		logAction(ctx, a.TypedFeatureContext, slog.LevelError, "channel_id is required in options")
		return fmt.Errorf("channel_id is required in options")
	}

	payload := a.GetPayload()
	if payload == nil {
		logAction(ctx, a.TypedFeatureContext, slog.LevelError, "message payload is required")
		return fmt.Errorf("message payload is required")
	}

	if err := payload.Validate(); err != nil {
		logAction(ctx, a.TypedFeatureContext, slog.LevelError, "invalid message payload", "error", err)
		return fmt.Errorf("invalid message payload: %w", err)
	}

	creds, err := credentials.ParseCredentials(a.GetCredentials(), opts.TenantID)
	if err != nil {
		logAction(ctx, a.TypedFeatureContext, slog.LevelError, "failed to parse credentials", "error", err)
		return fmt.Errorf("failed to parse credentials: %w", err)
	}

	token, err := creds.GetAccessToken(ctx)
	if err != nil {
		logAction(ctx, a.TypedFeatureContext, slog.LevelError, "failed to acquire access token", "error", err)
		return fmt.Errorf("failed to acquire access token: %w", err)
	}

	a.token = token
	a.teamID = opts.TeamID
	a.channelID = opts.ChannelID
	a.initialized = true

	return nil
}

// Start begins processing and posts the message to the configured channel.
func (a *SendMessageAction) Start(ctx context.Context) error {
	if err := msgraph_api.EnsureContextActive(ctx); err != nil {
		return err
	}

	if err := helpers.CheckInitialized(a.initialized); err != nil {
		return err
	}

	payload := a.GetPayload()
	if payload == nil {
		logAction(ctx, a.TypedFeatureContext, slog.LevelError, "message payload is required")
		return fmt.Errorf("message payload is required")
	}

	// Validate message again at runtime
	if err := payload.Validate(); err != nil {
		logAction(ctx, a.TypedFeatureContext, slog.LevelError, "invalid message", "error", err)
		return fmt.Errorf("invalid message: %w", err)
	}

	message := strings.TrimSpace(payload.Message)

	if err := channels.SendMessage(ctx, a.token, a.teamID, a.channelID, message); err != nil {
		logAction(ctx, a.TypedFeatureContext, slog.LevelError, "failed to send message", "error", err)
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

// Stop halts message processing and releases resources.
func (a *SendMessageAction) Stop(ctx context.Context) error {
	if err := msgraph_api.EnsureContextActive(ctx); err != nil {
		return err
	}

	if err := helpers.CheckInitialized(a.initialized); err != nil {
		return err
	}

	a.initialized = false
	a.token = ""
	a.teamID = ""
	a.channelID = ""

	return nil
}
