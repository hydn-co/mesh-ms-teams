package collectors

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/hydn-co/mesh-ms-teams/internal/credentials"
	"github.com/hydn-co/mesh-ms-teams/internal/helpers"
	"github.com/hydn-co/mesh-ms-teams/internal/msgraph_api"
	"github.com/hydn-co/mesh-ms-teams/internal/options"
	"github.com/hydn-co/mesh-ms-teams/internal/users"
	"github.com/hydn-co/mesh-sdk/pkg/catalog/entities"
	"github.com/hydn-co/mesh-sdk/pkg/catalog/spaces"
	"github.com/hydn-co/mesh-sdk/pkg/catalog/types"
	"github.com/hydn-co/mesh-sdk/pkg/connector"
	"github.com/hydn-co/mesh-sdk/pkg/runner"
)

// UsersCollector collects users from Microsoft Teams and emits them as catalog entities.
type UsersCollector struct {
	*connector.TypedFeatureContext[*options.UsersCollectorOptions, *connector.NoPayload]
	token       string
	initialized bool
}

// NewUsersCollector constructs a UsersCollector.
func NewUsersCollector(
	ctx *connector.TypedFeatureContext[*options.UsersCollectorOptions, *connector.NoPayload],
) runner.Feature {
	return &UsersCollector{TypedFeatureContext: ctx}
}

// Init prepares the collector for operation by validating credentials.
func (c *UsersCollector) Init(ctx context.Context) error {
	if err := msgraph_api.EnsureContextActive(ctx); err != nil {
		return err
	}

	creds, err := credentials.ParseCredentials(c.GetCredentials())
	if err != nil {
		logCollector(ctx, c.TypedFeatureContext, slog.LevelError, "failed to parse credentials", "error", err)
		return fmt.Errorf("failed to parse credentials: %w", err)
	}

	token, err := creds.GetAccessToken(ctx)
	if err != nil {
		logCollector(ctx, c.TypedFeatureContext, slog.LevelError, "failed to acquire access token", "error", err)
		return fmt.Errorf("failed to acquire access token: %w", err)
	}

	c.token = token
	c.initialized = true
	return nil
}

// Start begins collecting users from Microsoft Teams.
func (c *UsersCollector) Start(ctx context.Context) error {
	if err := msgraph_api.EnsureContextActive(ctx); err != nil {
		return err
	}

	if err := helpers.CheckInitialized(c.initialized); err != nil {
		return err
	}

	opts := c.GetOptions()
	includeGuests := opts != nil && opts.IncludeGuests

	pageURL := ""
	for {
		if err := msgraph_api.EnsureContextActive(ctx); err != nil {
			return err
		}

		var result *users.ListUsersResult
		var err error

		if pageURL == "" {
			result, err = users.ListUsers(ctx, c.token)
		} else {
			result, err = users.ListUsersPage(ctx, c.token, pageURL)
		}

		if err != nil {
			logCollector(ctx, c.TypedFeatureContext, slog.LevelError, "failed to list users", "error", err)
			return fmt.Errorf("failed to list users: %w", err)
		}

		for _, user := range result.Value {
			if !includeGuests && user.UserType == "Guest" {
				continue
			}

			accountType := types.AccountTypeUser
			if user.UserType == "Guest" {
				accountType = types.AccountTypeGuest
			}

			accountEntity := &entities.Account{
				Metadata: types.EntityMetadata{Space: spaces.Accounts},
				// AccountRef is the Microsoft Graph / Entra ID object GUID for this user.
				AccountRef:  user.ID,
				AccountType: accountType,
				Name:        user.UserPrincipalName,
				DisplayName: user.DisplayName,
				FirstName:   user.GivenName,
				LastName:    user.Surname,
				Enabled:     user.AccountEnabled,
			}

			if user.Mail != "" {
				accountEntity.PrimaryEmail = &types.Email{Address: user.Mail}
			}

			if err := c.Emit(ctx, accountEntity); err != nil {
				logCollector(ctx, c.TypedFeatureContext, slog.LevelError,
					"failed to emit user", "user_id", user.ID, "error", err)
				return fmt.Errorf("failed to emit user %s: %w", user.ID, err)
			}
		}

		pageURL = result.OdataNextLink
		if pageURL == "" {
			break
		}
	}

	return nil
}

// Stop halts user collection and releases resources.
func (c *UsersCollector) Stop(ctx context.Context) error {
	if err := msgraph_api.EnsureContextActive(ctx); err != nil {
		return err
	}

	if err := helpers.CheckInitialized(c.initialized); err != nil {
		return err
	}

	c.initialized = false
	c.token = ""
	return nil
}
