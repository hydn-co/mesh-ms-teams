package collectors

import (
	"context"
	"fmt"

	"github.com/hydn-co/mesh-ms-teams/internal/channels"
	"github.com/hydn-co/mesh-ms-teams/internal/credentials"
	"github.com/hydn-co/mesh-ms-teams/internal/helpers"
	"github.com/hydn-co/mesh-ms-teams/internal/msgraph_api"
	"github.com/hydn-co/mesh-ms-teams/internal/options"
	"github.com/hydn-co/mesh-sdk/pkg/catalog/entities"
	"github.com/hydn-co/mesh-sdk/pkg/catalog/spaces"
	"github.com/hydn-co/mesh-sdk/pkg/catalog/types"
	"github.com/hydn-co/mesh-sdk/pkg/connector"
	"github.com/hydn-co/mesh-sdk/pkg/runner"
)

// ChannelsCollector collects channels from Microsoft Teams and emits them as catalog entities.
type ChannelsCollector struct {
	*connector.TypedFeatureContext[*options.ChannelsCollectorOptions, *connector.NoPayload]
	token       string
	teamID      string
	initialized bool
}

// NewChannelsCollector constructs a ChannelsCollector.
func NewChannelsCollector(ctx *connector.TypedFeatureContext[*options.ChannelsCollectorOptions, *connector.NoPayload]) runner.Feature {
	return &ChannelsCollector{TypedFeatureContext: ctx}
}

// Init prepares the collector for operation by validating credentials and options.
func (c *ChannelsCollector) Init(ctx context.Context) error {
	if err := msgraph_api.EnsureContextActive(ctx); err != nil {
		return err
	}

	opts := c.GetOptions()
	if opts == nil || opts.TeamID == "" {
		return fmt.Errorf("team_id is required in options")
	}

	creds, err := credentials.ParseCredentials(c.GetCredentials())
	if err != nil {
		return fmt.Errorf("failed to parse credentials: %w", err)
	}

	token, err := creds.GetAccessToken(ctx)
	if err != nil {
		return fmt.Errorf("failed to acquire access token: %w", err)
	}

	c.token = token
	c.teamID = opts.TeamID
	c.initialized = true
	return nil
}

// Start begins collecting channels from the specified team.
func (c *ChannelsCollector) Start(ctx context.Context) error {
	if err := msgraph_api.EnsureContextActive(ctx); err != nil {
		return err
	}

	if err := helpers.CheckInitialized(c.initialized); err != nil {
		return err
	}

	pageURL := ""
	for {
		if err := msgraph_api.EnsureContextActive(ctx); err != nil {
			return err
		}

		var result *channels.ListChannelsResult
		var err error

		if pageURL == "" {
			result, err = channels.ListChannels(ctx, c.token, c.teamID)
		} else {
			result, err = channels.ListChannelsPage(ctx, c.token, pageURL)
		}

		if err != nil {
			return fmt.Errorf("failed to list channels: %w", err)
		}

		for _, channel := range result.Value {
			channelEntity := &entities.Channel{
				Metadata:    types.EntityMetadata{Space: spaces.Channels},
				ChannelRef:  channel.ID,
				Name:        channel.DisplayName,
				Description: channel.Description,
			}

			if err := c.Emit(ctx, channelEntity); err != nil {
				return fmt.Errorf("failed to emit channel %s: %w", channel.ID, err)
			}
		}

		pageURL = result.OdataNextLink
		if pageURL == "" {
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

	if err := helpers.CheckInitialized(c.initialized); err != nil {
		return err
	}

	c.initialized = false
	c.token = ""
	c.teamID = ""
	return nil
}
