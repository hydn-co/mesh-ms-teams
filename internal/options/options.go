package options

import (
	"github.com/fgrzl/json/polymorphic"
	"github.com/hydn-co/mesh-sdk/pkg/catalog/spaces"
)

func init() {
	polymorphic.RegisterType[TeamsCollectorOptions]()
	polymorphic.RegisterType[ChannelsCollectorOptions]()
	polymorphic.RegisterType[SendMessageActionOptions]()
}

// TeamsCollectorOptions configures the teams collector.
type TeamsCollectorOptions struct {
	// IncludeArchived determines whether to include archived teams in the collection.
	IncludeArchived bool `json:"include_archived" description:"Include archived teams in collection"`
}

func (o *TeamsCollectorOptions) GetDiscriminator() string {
	return "mesh://ms-teams/collectors/teams/options"
}

func (o *TeamsCollectorOptions) GetSpaces() []spaces.Space {
	return []spaces.Space{spaces.Groups}
}

func (o *TeamsCollectorOptions) GetRequirements() []string {
	return []string{"teams"}
}

// ChannelsCollectorOptions configures the channels collector.
type ChannelsCollectorOptions struct {
	// TeamID is the ID of the team to collect channels from.
	TeamID string `json:"team_id" binding:"required" description:"Team ID to collect channels from"`
}

func (o *ChannelsCollectorOptions) GetDiscriminator() string {
	return "mesh://ms-teams/collectors/channels/options"
}

func (o *ChannelsCollectorOptions) GetSpaces() []spaces.Space {
	return []spaces.Space{spaces.Channels}
}

func (o *ChannelsCollectorOptions) GetRequirements() []string {
	return []string{"teams"}
}

// SendMessageActionOptions configures the send-message action.
type SendMessageActionOptions struct {
	// TeamID is the ID of the team containing the target channel.
	TeamID string `json:"team_id" binding:"required" description:"Team ID containing the target channel"`

	// ChannelID is the ID of the channel to post the message to.
	ChannelID string `json:"channel_id" binding:"required" description:"Channel ID to post the message to"`
}

func (o *SendMessageActionOptions) GetDiscriminator() string {
	return "mesh://ms-teams/actions/send-message/options"
}

func (o *SendMessageActionOptions) GetSpaces() []spaces.Space {
	return []spaces.Space{spaces.Channels}
}

func (o *SendMessageActionOptions) GetRequirements() []string {
	return []string{"teams"}
}
