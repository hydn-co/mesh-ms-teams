package main

import (
	"github.com/hydn-co/mesh-ms-teams/internal/actions/provision_user"
	"github.com/hydn-co/mesh-ms-teams/internal/actions/send_message"
	"github.com/hydn-co/mesh-ms-teams/internal/collectors/channels"
	"github.com/hydn-co/mesh-ms-teams/internal/collectors/teams"
	"github.com/hydn-co/mesh-ms-teams/internal/collectors/users"
	"github.com/hydn-co/mesh-sdk/pkg/connector"
	"github.com/hydn-co/mesh-sdk/pkg/runner"
)

const (
	manifestName        = "ms-teams"
	manifestDisplayName = "Microsoft Teams"
	manifestDescription = "Collects and acts on Microsoft Teams data via the Microsoft Graph API."
)

func main() {
	runner.Run(buildManifest())
}

func buildManifest() *runner.Manifest {
	m := runner.CreateManifest(manifestName, "", manifestDisplayName, manifestDescription)

	mustRegister(m, "collect-users", "Collect Users", "Collects user accounts from Microsoft Teams.",
		true, runner.FeatureTypeCollector, &users.Options{}, nil,
		runner.FeatureResumeBehaviorLastActivity, runner.ClientCredential, users.New)

	mustRegister(m, "collect-channels", "Collect Channels", "Collects channels from Microsoft Teams.",
		true, runner.FeatureTypeCollector, &channels.Options{}, nil,
		runner.FeatureResumeBehaviorLastActivity, runner.ClientCredential, channels.New)

	mustRegister(m, "collect-teams", "Collect Teams", "Collects teams (groups) from Microsoft Teams.",
		true, runner.FeatureTypeCollector, &teams.Options{}, nil,
		runner.FeatureResumeBehaviorLastActivity, runner.ClientCredential, teams.New)

	mustRegister(m, "send-message", "Send Message", "Sends a message to a Microsoft Teams channel.",
		false, runner.FeatureTypeAction, &send_message.Options{}, &send_message.Payload{},
		runner.FeatureResumeBehaviorNone, runner.ClientCredential, send_message.New)

	mustRegister(m, "provision-user", "Provision User", "Provisions a new user in Microsoft Teams / Entra ID.",
		false, runner.FeatureTypeAction, &provision_user.Options{}, &provision_user.Payload{},
		runner.FeatureResumeBehaviorNone, runner.ClientCredential, provision_user.New)

	return m
}

func mustRegister(
	m *runner.Manifest,
	name, displayName, description string,
	schedulable bool,
	featureType runner.FeatureType,
	options connector.FeatureOptions,
	payload connector.FeaturePayload,
	resumeBehavior runner.FeatureResumeBehavior,
	secretTemplateName string,
	factory func(...connector.FeatureContextOption) runner.Feature,
) {
	if err := m.RegisterFeature(
		name, displayName, description,
		schedulable, featureType,
		options, payload,
		resumeBehavior, secretTemplateName,
		factory,
	); err != nil {
		panic("failed to register feature " + name + ": " + err.Error())
	}
}
