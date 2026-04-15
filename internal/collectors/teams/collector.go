package teams

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

// Options configures the teams collector.
type Options struct {
	IncludeArchived bool `json:"include_archived" description:"Whether to include archived teams"`
}

func (o *Options) GetDiscriminator() string  { return "mesh://ms-teams/collectors/teams/options" }
func (o *Options) GetSpaces() []spaces.Space { return []spaces.Space{spaces.Groups} }
func (o *Options) GetRequirements() []string { return []string{"teams"} }

// Collector collects Microsoft Teams teams (groups).
type Collector struct {
	ctx *connector.TypedFeatureContext[*Options]
}

// New creates a new teams Collector from the provided context options.
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
