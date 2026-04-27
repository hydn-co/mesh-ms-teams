package main

import (
	"log"

	"github.com/hydn-co/mesh-ms-teams/internal/actions"
	"github.com/hydn-co/mesh-ms-teams/internal/collectors"
	"github.com/hydn-co/mesh-ms-teams/internal/options"
	"github.com/hydn-co/mesh-ms-teams/internal/payloads"
	"github.com/hydn-co/mesh-sdk/pkg/runner"
)

func main() {
	runner.Run(WithManifest())
}

func WithManifest() *runner.Manifest {
	manifest := runner.CreateManifest(
		"mesh-ms-teams",
		"",
		"Microsoft Teams",
		"Mesh integration with Microsoft Teams via Microsoft Graph API",
	)

	// Register Users Collector
	if err := manifest.RegisterFeature(
		"users_collector",
		"Users Collector",
		"Collects users from Microsoft Teams and emits them as catalog entities.",
		true,
		runner.FeatureTypeCollector,
		new(options.UsersCollectorOptions),
		nil,
		runner.FeatureResumeBehaviorLastActivity,
		runner.ClientCredential,
		runner.Factory(collectors.NewUsersCollector),
	); err != nil {
		log.Fatal(err)
	}

	// Register Teams Collector
	if err := manifest.RegisterFeature(
		"teams_collector",
		"Teams Collector",
		"Collects teams from Microsoft Teams and emits them as catalog entities.",
		true,
		runner.FeatureTypeCollector,
		new(options.TeamsCollectorOptions),
		nil,
		runner.FeatureResumeBehaviorLastActivity,
		runner.ClientCredential,
		runner.Factory(collectors.NewTeamsCollector),
	); err != nil {
		log.Fatal(err)
	}

	// Register Channels Collector
	if err := manifest.RegisterFeature(
		"channels_collector",
		"Channels Collector",
		"Collects channels from a specified Microsoft Teams team and emits them as catalog entities.",
		true,
		runner.FeatureTypeCollector,
		new(options.ChannelsCollectorOptions),
		nil,
		runner.FeatureResumeBehaviorLastActivity,
		runner.ClientCredential,
		runner.Factory(collectors.NewChannelsCollector),
	); err != nil {
		log.Fatal(err)
	}

	// Register Send Message Action
	if err := manifest.RegisterFeature(
		"send_message_action",
		"Send Message Action",
		"Posts messages to Microsoft Teams channels.",
		false,
		runner.FeatureTypeAction,
		new(options.SendMessageActionOptions),
		new(payloads.SendMessagePayload),
		runner.FeatureResumeBehaviorNone,
		runner.ClientCredential,
		runner.Factory(actions.NewSendMessageAction),
	); err != nil {
		log.Fatal(err)
	}

	// Register Provision User Action
	if err := manifest.RegisterFeature(
		"provision_user_action",
		"Provision User Action",
		"Provisions a new user in Entra ID for Microsoft Teams.",
		false,
		runner.FeatureTypeAction,
		new(options.ProvisionUserActionOptions),
		new(payloads.ProvisionUserPayload),
		runner.FeatureResumeBehaviorNone,
		runner.ClientCredential,
		runner.Factory(actions.NewProvisionUserAction),
	); err != nil {
		log.Fatal(err)
	}

	if err := manifest.Validate(); err != nil {
		log.Fatal(err)
	}

	return manifest
}
