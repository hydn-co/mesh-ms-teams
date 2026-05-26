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

	"github.com/hydn-co/mesh-ms-teams/internal/channels"
	"github.com/hydn-co/mesh-ms-teams/internal/credentials"
	"github.com/hydn-co/mesh-ms-teams/internal/msgraph_api"
	"github.com/hydn-co/mesh-ms-teams/internal/options"
	"github.com/hydn-co/mesh-ms-teams/internal/teams"
)

// ChannelsCollector collects channels across all teams and emits them as catalog entities.
type ChannelsCollector struct {
	*connector.TypedFeatureContext[*options.ChannelsCollectorOptions, *connector.NoPayload]
	token string
	state connectorutil.FeatureState
}

// NewChannelsCollector constructs a ChannelsCollector.
func NewChannelsCollector(
	ctx *connector.TypedFeatureContext[*options.ChannelsCollectorOptions, *connector.NoPayload],
) runner.Feature {
	return &ChannelsCollector{TypedFeatureContext: ctx}
}

// Init prepares the collector for operation by validating credentials.
func (c *ChannelsCollector) Init(ctx context.Context) error {
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

// Start collects channels across all teams accessible to the service principal.
func (c *ChannelsCollector) Start(ctx context.Context) error {
	if err := msgraph_api.EnsureContextActive(ctx); err != nil {
		return err
	}

	if err := c.state.RequireReady(); err != nil {
		return err
	}

	teamPageURL := ""
	for {
		if err := msgraph_api.EnsureContextActive(ctx); err != nil {
			return err
		}

		var teamResult *teams.ListTeamsResult
		var err error

		if teamPageURL == "" {
			teamResult, err = teams.ListTeams(ctx, c.token)
		} else {
			teamResult, err = teams.ListTeamsPage(ctx, c.token, teamPageURL)
		}

		if err != nil {
			connectorutil.LogFeature(ctx, c.TypedFeatureContext, slog.LevelError, "failed to list teams", "error", err)
			return fmt.Errorf("failed to list teams: %w", err)
		}

		for _, team := range teamResult.Value {
			if err := c.collectChannelsForTeam(ctx, team.ID); err != nil {
				return err
			}
		}

		teamPageURL = teamResult.OdataNextLink
		if teamPageURL == "" {
			break
		}
	}

	return nil
}

// collectChannelsForTeam fetches and emits all channels for a single team.
func (c *ChannelsCollector) collectChannelsForTeam(ctx context.Context, teamID string) error {
	channelPageURL := ""
	for {
		if err := msgraph_api.EnsureContextActive(ctx); err != nil {
			return err
		}

		var result *channels.ListChannelsResult
		var err error

		if channelPageURL == "" {
			result, err = channels.ListChannels(ctx, c.token, teamID)
		} else {
			result, err = channels.ListChannelsPage(ctx, c.token, channelPageURL)
		}

		if err != nil {
			connectorutil.LogFeature(ctx, c.TypedFeatureContext, slog.LevelError, "failed to list channels",
				"team_id", teamID, "error", err)
			return fmt.Errorf("failed to list channels for team %s: %w", teamID, err)
		}

		for _, channel := range result.Value {
			channelEntity := &entities.Channel{
				Metadata:    types.EntityMetadata{Space: spaces.Channels},
				ChannelRef:  channel.ID,
				Name:        channel.DisplayName,
				Description: channel.Description,
			}

			if err := c.Emit(ctx, channelEntity); err != nil {
				connectorutil.LogFeature(ctx, c.TypedFeatureContext, slog.LevelError,
					"failed to emit channel", "team_id", teamID, "channel_id", channel.ID, "error", err)
				return fmt.Errorf("failed to emit channel %s: %w", channel.ID, err)
			}
		}

		channelPageURL = result.OdataNextLink
		if channelPageURL == "" {
			break
		}
	}

	return nil
}

// Stop halts channel collection and releases resources.
func (c *ChannelsCollector) Stop(ctx context.Context) error {
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
