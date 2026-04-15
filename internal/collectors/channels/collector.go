package channels

import (
	"context"

	"github.com/fgrzl/json/polymorphic"
	"github.com/hydn-co/mesh-sdk/pkg/catalog/spaces"
	"github.com/hydn-co/mesh-sdk/pkg/connector"
	"github.com/hydn-co/mesh-sdk/pkg/runner"
)

func init() {
	polymorphic.RegisterType[Options]()
}

// Options configures the channels collector.
type Options struct {
	IncludePrivate bool `json:"include_private" description:"Whether to include private channels"`
}

func (o *Options) GetDiscriminator() string  { return "mesh://ms-teams/collectors/channels/options" }
func (o *Options) GetSpaces() []spaces.Space { return []spaces.Space{spaces.Channels} }
func (o *Options) GetRequirements() []string { return []string{"channels"} }

// Collector collects Microsoft Teams channels.
type Collector struct {
	ctx *connector.TypedFeatureContext[*Options]
}

// New creates a new channels Collector from the provided context options.
func New(opts ...connector.FeatureContextOption) runner.Feature {
	return &Collector{
		ctx: connector.NewTypedFeatureContext[*Options](
			connector.NewFeatureContext(opts...),
		),
	}
}

func (c *Collector) Init(_ context.Context) error  { return nil }
func (c *Collector) Start(_ context.Context) error { return nil }
func (c *Collector) Stop(_ context.Context) error  { return nil }
