package users

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

// Options configures the users collector.
type Options struct {
	IncludeGuests bool `json:"include_guests" description:"Whether to include guest accounts"`
}

func (o *Options) GetDiscriminator() string  { return "mesh://ms-teams/collectors/users/options" }
func (o *Options) GetSpaces() []spaces.Space { return []spaces.Space{spaces.Accounts} }
func (o *Options) GetRequirements() []string { return []string{"users"} }

// Collector collects Microsoft Teams user accounts.
type Collector struct {
	ctx *connector.TypedFeatureContext[*Options]
}

// New creates a new users Collector from the provided context options.
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
