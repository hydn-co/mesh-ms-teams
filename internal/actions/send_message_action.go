package actions

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/hydn-co/mesh-sdk/pkg/connector"
	"github.com/hydn-co/mesh-sdk/pkg/connectorutil"
	"github.com/hydn-co/mesh-sdk/pkg/runner"

	"github.com/hydn-co/mesh-ms-teams/internal/channels"
	"github.com/hydn-co/mesh-ms-teams/internal/credentials"
	"github.com/hydn-co/mesh-ms-teams/internal/msgraph_api"
	"github.com/hydn-co/mesh-ms-teams/internal/options"
	"github.com/hydn-co/mesh-ms-teams/internal/payloads"
)

// SendMessageAction posts messages to Microsoft Teams channels.
type SendMessageAction struct {
	*connector.TypedFeatureContext[*options.SendMessageActionOptions, *payloads.SendMessagePayload]
	token     string
	teamID    string
	channelID string
	message   string
	state     connectorutil.FeatureState
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
	if err := connectorutil.Validate(opts, "feature options"); err != nil {
		connectorutil.LogFeature(ctx, a.TypedFeatureContext, slog.LevelError, err.Error())
		return err
	}

	payload := a.GetPayload()
	if err := connectorutil.Validate(payload, "send message payload"); err != nil {
		connectorutil.LogFeature(ctx, a.TypedFeatureContext, slog.LevelError, err.Error())
		return err
	}

	creds, err := credentials.ParseCredentials(a.GetCredentials(), opts.TenantID)
	if err != nil {
		connectorutil.LogFeature(
			ctx,
			a.TypedFeatureContext,
			slog.LevelError,
			"failed to parse credentials",
			"error",
			err,
		)
		return fmt.Errorf("failed to parse credentials: %w", err)
	}

	token, err := creds.GetAccessToken(ctx)
	if err != nil {
		connectorutil.LogFeature(
			ctx,
			a.TypedFeatureContext,
			slog.LevelError,
			"failed to acquire access token",
			"error",
			err,
		)
		return fmt.Errorf("failed to acquire access token: %w", err)
	}

	a.token = token
	a.teamID = opts.TeamID
	a.channelID = opts.ChannelID
	a.message = payload.Message
	a.state.MarkReady()

	return nil
}

// Start begins processing and posts the message to the configured channel.
func (a *SendMessageAction) Start(ctx context.Context) error {
	if err := msgraph_api.EnsureContextActive(ctx); err != nil {
		return err
	}

	if err := a.state.RequireReady(); err != nil {
		return err
	}

	if err := channels.SendMessage(ctx, a.token, a.teamID, a.channelID, a.message); err != nil {
		connectorutil.LogFeature(ctx, a.TypedFeatureContext, slog.LevelError, "failed to send message", "error", err)
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

// Stop halts message processing and releases resources.
func (a *SendMessageAction) Stop(ctx context.Context) error {
	if err := msgraph_api.EnsureContextActive(ctx); err != nil {
		return err
	}

	if err := a.state.RequireReady(); err != nil {
		return err
	}

	a.state.Reset()
	a.token = ""
	a.teamID = ""
	a.channelID = ""
	a.message = ""

	return nil
}
