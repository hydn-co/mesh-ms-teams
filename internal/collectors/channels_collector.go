package collectors

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/hydn-co/mesh-sdk/pkg/catalog/entities"
	"github.com/hydn-co/mesh-sdk/pkg/connector"
	"github.com/hydn-co/mesh-sdk/pkg/connectorutil"
	"github.com/hydn-co/mesh-sdk/pkg/runner"

	"github.com/hydn-co/mesh-ms-teams/internal/channels"
	"github.com/hydn-co/mesh-ms-teams/internal/credentials"
	"github.com/hydn-co/mesh-ms-teams/internal/msgraph_api"
	"github.com/hydn-co/mesh-ms-teams/internal/options"
	"github.com/hydn-co/mesh-ms-teams/internal/teams"
)

// channelsGraphClient is the collector-local view of the Microsoft Graph API
// used by the channels collector. Contract tests inject a fake through newClient.
type channelsGraphClient interface {
	ListTeams(ctx context.Context) (*teams.ListTeamsResult, error)
	ListTeamsPage(ctx context.Context, pageURL string) (*teams.ListTeamsResult, error)
	ListChannels(ctx context.Context, teamID string) (*channels.ListChannelsResult, error)
	ListChannelsPage(ctx context.Context, pageURL string) (*channels.ListChannelsResult, error)
}

// channelsGraphClientFactory builds a channelsGraphClient from parsed credentials.
type channelsGraphClientFactory func(ctx context.Context, creds *credentials.AzureADCredentials) (channelsGraphClient, error)

// defaultChannelsGraphClientFactory exchanges the parsed credentials for an
// access token and returns a token-bound Microsoft Graph client.
func defaultChannelsGraphClientFactory(
	ctx context.Context,
	creds *credentials.AzureADCredentials,
) (channelsGraphClient, error) {
	token, err := creds.GetAccessToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire access token: %w", err)
	}
	return &channelsGraphAPI{token: token}, nil
}

// channelsGraphAPI adapts the package-level Graph helpers to channelsGraphClient.
type channelsGraphAPI struct {
	token string
}

func (a *channelsGraphAPI) ListTeams(ctx context.Context) (*teams.ListTeamsResult, error) {
	return teams.ListTeams(ctx, a.token)
}

func (a *channelsGraphAPI) ListTeamsPage(ctx context.Context, pageURL string) (*teams.ListTeamsResult, error) {
	return teams.ListTeamsPage(ctx, a.token, pageURL)
}

func (a *channelsGraphAPI) ListChannels(ctx context.Context, teamID string) (*channels.ListChannelsResult, error) {
	return channels.ListChannels(ctx, a.token, teamID)
}

func (a *channelsGraphAPI) ListChannelsPage(
	ctx context.Context,
	pageURL string,
) (*channels.ListChannelsResult, error) {
	return channels.ListChannelsPage(ctx, a.token, pageURL)
}

// ChannelsCollector collects channels across all teams and emits them as catalog entities.
type ChannelsCollector struct {
	*connector.TypedFeatureContext[*options.ChannelsCollectorOptions, *connector.NoPayload]
	client    channelsGraphClient
	newClient channelsGraphClientFactory
	state     connectorutil.FeatureState
}

// NewChannelsCollector constructs a ChannelsCollector.
func NewChannelsCollector(
	ctx *connector.TypedFeatureContext[*options.ChannelsCollectorOptions, *connector.NoPayload],
) runner.Feature {
	return &ChannelsCollector{
		TypedFeatureContext: ctx,
		newClient:           defaultChannelsGraphClientFactory,
	}
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

	client, err := c.newClient(ctx, creds)
	if err != nil {
		connectorutil.LogFeature(
			ctx,
			c.TypedFeatureContext,
			slog.LevelError,
			"failed to create Microsoft Graph client",
			"error",
			err,
		)
		return fmt.Errorf("failed to create Microsoft Graph client: %w", err)
	}

	c.client = client
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
			teamResult, err = c.client.ListTeams(ctx)
		} else {
			teamResult, err = c.client.ListTeamsPage(ctx, teamPageURL)
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
			result, err = c.client.ListChannels(ctx, teamID)
		} else {
			result, err = c.client.ListChannelsPage(ctx, channelPageURL)
		}

		if err != nil {
			connectorutil.LogFeature(ctx, c.TypedFeatureContext, slog.LevelError, "failed to list channels",
				"team_id", teamID, "error", err)
			return fmt.Errorf("failed to list channels for team %s: %w", teamID, err)
		}

		for _, channel := range result.Value {
			channelEntity := &entities.Channel{
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
	c.client = nil
	return nil
}
