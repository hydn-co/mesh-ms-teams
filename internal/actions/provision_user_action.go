package actions

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/hydn-co/mesh-ms-teams/internal/credentials"
	"github.com/hydn-co/mesh-ms-teams/internal/helpers"
	"github.com/hydn-co/mesh-ms-teams/internal/msgraph_api"
	"github.com/hydn-co/mesh-ms-teams/internal/options"
	"github.com/hydn-co/mesh-ms-teams/internal/payloads"
	"github.com/hydn-co/mesh-ms-teams/internal/users"
	"github.com/hydn-co/mesh-sdk/pkg/connector"
	"github.com/hydn-co/mesh-sdk/pkg/runner"
)

// ProvisionUserAction creates a new user in Entra ID / Microsoft Teams.
type ProvisionUserAction struct {
	*connector.TypedFeatureContext[*options.ProvisionUserActionOptions, *payloads.ProvisionUserPayload]
	token       string
	initialized bool
}

// NewProvisionUserAction constructs a ProvisionUserAction.
func NewProvisionUserAction(
	ctx *connector.TypedFeatureContext[*options.ProvisionUserActionOptions, *payloads.ProvisionUserPayload],
) runner.Feature {
	return &ProvisionUserAction{TypedFeatureContext: ctx}
}

// Init prepares the action for operation by validating credentials and payload.
func (a *ProvisionUserAction) Init(ctx context.Context) error {
	if err := msgraph_api.EnsureContextActive(ctx); err != nil {
		return err
	}

	payload := a.GetPayload()
	if payload == nil {
		logAction(ctx, a.TypedFeatureContext, slog.LevelError, "provision-user payload is required")
		return fmt.Errorf("provision-user payload is required")
	}

	if err := payload.Validate(); err != nil {
		logAction(ctx, a.TypedFeatureContext, slog.LevelError, "invalid provision-user payload", "error", err)
		return fmt.Errorf("invalid provision-user payload: %w", err)
	}

	creds, err := credentials.ParseCredentials(a.GetCredentials())
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
	a.initialized = true
	return nil
}

// Start provisions the user described by the action payload.
func (a *ProvisionUserAction) Start(ctx context.Context) error {
	if err := msgraph_api.EnsureContextActive(ctx); err != nil {
		return err
	}

	if err := helpers.CheckInitialized(a.initialized); err != nil {
		return err
	}

	payload := a.GetPayload()
	if payload == nil {
		logAction(ctx, a.TypedFeatureContext, slog.LevelError, "provision-user payload is required")
		return fmt.Errorf("provision-user payload is required")
	}

	if err := payload.Validate(); err != nil {
		logAction(ctx, a.TypedFeatureContext, slog.LevelError, "invalid provision-user payload", "error", err)
		return fmt.Errorf("invalid provision-user payload: %w", err)
	}

	req := users.ProvisionUserRequest{
		AccountEnabled:    true,
		DisplayName:       payload.DisplayName,
		MailNickname:      payload.MailNickname,
		UserPrincipalName: payload.UserPrincipalName,
		PasswordProfile: users.PasswordProfile{
			ForceChangePasswordNextSignIn: true,
			Password:                      payload.Password,
		},
	}

	if err := users.ProvisionUser(ctx, a.token, req); err != nil {
		logAction(ctx, a.TypedFeatureContext, slog.LevelError,
			"failed to provision user", "upn", payload.UserPrincipalName, "error", err)
		return fmt.Errorf("failed to provision user %q: %w", payload.UserPrincipalName, err)
	}

	return nil
}

// Stop halts the action and releases resources.
func (a *ProvisionUserAction) Stop(ctx context.Context) error {
	if err := msgraph_api.EnsureContextActive(ctx); err != nil {
		return err
	}

	if err := helpers.CheckInitialized(a.initialized); err != nil {
		return err
	}

	a.initialized = false
	a.token = ""
	return nil
}
