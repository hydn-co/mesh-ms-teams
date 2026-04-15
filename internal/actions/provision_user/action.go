package provision_user

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

// Options configures the user-provisioning action.
type Options struct{}

func (o *Options) GetDiscriminator() string  { return "mesh://ms-teams/actions/provision-user/options" }
func (o *Options) GetSpaces() []spaces.Space { return []spaces.Space{spaces.Accounts} }
func (o *Options) GetRequirements() []string { return []string{"users"} }

// Payload carries the input data for the user-provisioning action.
type Payload struct {
	DisplayName       string `json:"display_name" description:"Display name for the new user"`
	UserPrincipalName string `json:"user_principal_name" description:"User principal name (UPN) for the new account"`
	MailNickname      string `json:"mail_nickname" description:"Mail nickname for the new user"`
	Password          string `json:"password" description:"Initial password for the new user account"`
}

func (p *Payload) GetDiscriminator() string {
	return "mesh://ms-teams/actions/provision-user/payload"
}

// Action provisions a new user in Microsoft Teams / Entra ID.
type Action struct {
	ctx *connector.TypedFeatureContext[*Options]
}

// New creates a new provision-user Action from the provided context options.
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
