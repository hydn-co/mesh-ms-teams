package send_message

import (
	"context"

	"github.com/fgrzl/json/polymorphic"
	"github.com/hydn-co/mesh-sdk/pkg/catalog/spaces"
	"github.com/hydn-co/mesh-sdk/pkg/connector"
	"github.com/hydn-co/mesh-sdk/pkg/runner"
)

func init() {
	polymorphic.RegisterType[Options]()
	polymorphic.RegisterType[Payload]()
}

// Options configures the send-message action.
type Options struct{}

func (o *Options) GetDiscriminator() string  { return "mesh://ms-teams/actions/send-message/options" }
func (o *Options) GetSpaces() []spaces.Space { return []spaces.Space{spaces.Channels} }
func (o *Options) GetRequirements() []string { return []string{"channels"} }

// Payload carries the input data for the send-message action.
type Payload struct {
	TeamID    string `json:"team_id" description:"ID of the team containing the target channel"`
	ChannelID string `json:"channel_id" description:"ID of the channel to post the message to"`
	Message   string `json:"message" description:"Content of the message to send"`
}

func (p *Payload) GetDiscriminator() string { return "mesh://ms-teams/actions/send-message/payload" }

// Action sends a message to a Microsoft Teams channel.
type Action struct {
	ctx *connector.TypedFeatureContext[*Options]
}

// New creates a new send-message Action from the provided context options.
func New(opts ...connector.FeatureContextOption) runner.Feature {
	return &Action{
		ctx: connector.NewTypedFeatureContext[*Options](
			connector.NewFeatureContext(opts...),
		),
	}
}

func (a *Action) Init(_ context.Context) error  { return nil }
func (a *Action) Start(_ context.Context) error { return nil }
func (a *Action) Stop(_ context.Context) error  { return nil }
