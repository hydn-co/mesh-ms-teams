package collectors

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/hydn-co/mesh-sdk/pkg/catalog/entities"
	"github.com/hydn-co/mesh-sdk/pkg/catalog/spaces"
	"github.com/hydn-co/mesh-sdk/pkg/catalog/types"
	"github.com/hydn-co/mesh-sdk/pkg/connector"
	"github.com/hydn-co/mesh-sdk/pkg/connectorutil"
	"github.com/hydn-co/mesh-sdk/pkg/runner"

	"github.com/hydn-co/mesh-ms-teams/internal/credentials"
	"github.com/hydn-co/mesh-ms-teams/internal/msgraph_api"
	"github.com/hydn-co/mesh-ms-teams/internal/options"
	"github.com/hydn-co/mesh-ms-teams/internal/teams"
)

// TeamsCollector collects teams from Microsoft Teams and emits them as catalog entities.
type TeamsCollector struct {
	*connector.TypedFeatureContext[*options.TeamsCollectorOptions, *connector.NoPayload]
	token string
	state connectorutil.FeatureState
}

// NewTeamsCollector constructs a TeamsCollector.
func NewTeamsCollector(
	ctx *connector.TypedFeatureContext[*options.TeamsCollectorOptions, *connector.NoPayload],
) runner.Feature {
	return &TeamsCollector{TypedFeatureContext: ctx}
}

// Init prepares the collector for operation by validating credentials.
func (c *TeamsCollector) Init(ctx context.Context) error {
	if err := msgraph_api.EnsureContextActive(ctx); err != nil {
		return err
	}

	opts := c.GetOptions()
	if err := connectorutil.Validate(opts, "feature options"); err != nil {
		connectorutil.LogFeature(ctx, c.TypedFeatureContext, slog.LevelError, err.Error())
		return err
	}

	creds, err := credentials.ParseCredentials(c.GetCredentials(), opts.TenantID)
	if err != nil {
		connectorutil.LogFeature(
			ctx,
			c.TypedFeatureContext,
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
			c.TypedFeatureContext,
			slog.LevelError,
			"failed to acquire access token",
			"error",
			err,
		)
		return fmt.Errorf("failed to acquire access token: %w", err)
	}

	c.token = token
	c.state.MarkReady()
	return nil
}

// Start begins collecting teams from Microsoft Teams.
func (c *TeamsCollector) Start(ctx context.Context) error {
	if err := msgraph_api.EnsureContextActive(ctx); err != nil {
		return err
	}

	if err := c.state.RequireReady(); err != nil {
		return err
	}

	pageURL := ""
	for {
		if err := msgraph_api.EnsureContextActive(ctx); err != nil {
			return err
		}

		var result *teams.ListTeamsResult
		var err error

		if pageURL == "" {
			result, err = teams.ListTeams(ctx, c.token)
		} else {
			result, err = teams.ListTeamsPage(ctx, c.token, pageURL)
		}

		if err != nil {
			connectorutil.LogFeature(ctx, c.TypedFeatureContext, slog.LevelError, "failed to list teams", "error", err)
			return fmt.Errorf("failed to list teams: %w", err)
		}

		for _, team := range result.Value {
			groupEntity := &entities.Group{
				Metadata:    types.EntityMetadata{Space: spaces.Groups},
				GroupRef:    team.ID,
				Name:        team.DisplayName,
				Description: team.Description,
			}

			if err := c.Emit(ctx, groupEntity); err != nil {
				connectorutil.LogFeature(ctx, c.TypedFeatureContext, slog.LevelError,
					"failed to emit team", "team_id", team.ID, "error", err)
				return fmt.Errorf("failed to emit team %s: %w", team.ID, err)
			}
		}

		pageURL = result.OdataNextLink
		if pageURL == "" {
			break
		}
	}

	return nil
}

// Stop halts team collection and releases resources.
func (c *TeamsCollector) Stop(ctx context.Context) error {
	if err := msgraph_api.EnsureContextActive(ctx); err != nil {
		return err
	}

	if err := c.state.RequireReady(); err != nil {
		return err
	}

	c.state.Reset()
	c.token = ""
	return nil
}
