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

// TeamsOptionsCore contains common fields shared by all Teams feature option types.
type TeamsOptionsCore struct {
	TenantID string `json:"tenant_id" title:"Tenant ID" description:"Microsoft Entra tenant ID used to authenticate Microsoft Graph requests for this connector." binding:"required"`
}

func (o *TeamsOptionsCore) GetTenantID() string {
	if o == nil {
		return ""
	}
	return o.TenantID
}

// TeamsCollectorOptions configures the teams collector.
type TeamsCollectorOptions struct {
	TeamsOptionsCore `json:",inline"`
	// IncludeArchived determines whether to include archived teams in the collection.
	IncludeArchived bool `json:"include_archived" title:"Include Archived Teams" description:"Include archived Microsoft Teams teams in the collector output."`
}

func (o *TeamsCollectorOptions) GetDiscriminator() string {
	return "mesh://ms-teams/collectors/ms_teams_teams_collector_options"
}

func (o *TeamsCollectorOptions) GetSpaces() []spaces.Space {
	return []spaces.Space{spaces.Groups}
}

func (o *TeamsCollectorOptions) GetRequirements() []string {
	return []string{"teams"}
}

// ChannelsCollectorOptions configures the channels collector.
// Channels are collected across all teams accessible to the service principal.
type ChannelsCollectorOptions struct {
	TeamsOptionsCore `json:",inline"`
}

func (o *ChannelsCollectorOptions) GetDiscriminator() string {
	return "mesh://ms-teams/collectors/ms_teams_channels_collector_options"
}

func (o *ChannelsCollectorOptions) GetSpaces() []spaces.Space {
	return []spaces.Space{spaces.Channels}
}

func (o *ChannelsCollectorOptions) GetRequirements() []string {
	return []string{"teams"}
}

// SendMessageActionOptions configures the send-message action.
type SendMessageActionOptions struct {
	TeamsOptionsCore `json:",inline"`

	// TeamID is the ID of the team containing the target channel.
	TeamID string `json:"team_id" binding:"required" title:"Team" description:"Microsoft Teams team that contains the target channel. Select a collected team from the Teams collector." x-lookup:"{\"entity-type\": \"groups\", \"display-key\": \"name\", \"submit-key\": \"group_ref\", \"form-input-type\": \"select\"}"`

	// ChannelID is the ID of the channel to post the message to.
	ChannelID string `json:"channel_id" binding:"required" title:"Channel" description:"Microsoft Teams channel that will receive the message. Select a collected channel from the Channels collector." x-lookup:"{\"entity-type\": \"channels\", \"display-key\": \"name\", \"submit-key\": \"channel_ref\", \"form-input-type\": \"select\"}"`
}

func (o *SendMessageActionOptions) GetDiscriminator() string {
	return "mesh://ms-teams/actions/ms_teams_send_message_action_options"
}

func (o *SendMessageActionOptions) GetSpaces() []spaces.Space {
	return []spaces.Space{spaces.Channels}
}

func (o *SendMessageActionOptions) GetRequirements() []string {
	return []string{"teams"}
}
