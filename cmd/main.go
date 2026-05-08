package main

import (
	"log"

	"github.com/hydn-co/mesh-ms-teams/internal/actions"
	"github.com/hydn-co/mesh-ms-teams/internal/collectors"
	"github.com/hydn-co/mesh-ms-teams/internal/options"
	"github.com/hydn-co/mesh-ms-teams/internal/payloads"
	"github.com/hydn-co/mesh-sdk/pkg/connector"
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
		"Mesh integration with Microsoft Teams",
	)

	// Register Teams Collector
	manifest.MustRegisterFeature(
		"teams_collector",
		"Teams Collector",
		"Collects teams from Microsoft Teams and emits them as catalog entities.",
		runner.FeatureSchedulable,
		runner.FeatureTypeCollector,
		new(options.TeamsCollectorOptions),
		(*connector.NoPayload)(nil),
		runner.FeatureResumeBehaviorNone,
		runner.GrantCredential,
		runner.Factory(collectors.NewTeamsCollector),
	)

	// Register Channels Collector
	manifest.MustRegisterFeature(
		"channels_collector",
		"Channels Collector",
		"Collects channels from a specified Microsoft Teams team and emits them as catalog entities.",
		runner.FeatureSchedulable,
		runner.FeatureTypeCollector,
		new(options.ChannelsCollectorOptions),
		(*connector.NoPayload)(nil),
		runner.FeatureResumeBehaviorNone,
		runner.GrantCredential,
		runner.Factory(collectors.NewChannelsCollector),
	)

	// Register Send Message Action
	manifest.MustRegisterFeature(
		"send_message_action",
		"Send Message Action",
		"Posts messages to Microsoft Teams channels.",
		runner.FeatureUnschedulable,
		runner.FeatureTypeAction,
		new(options.SendMessageActionOptions),
		new(payloads.SendMessagePayload),
		runner.FeatureResumeBehaviorNone,
		runner.GrantCredential,
		runner.Factory(actions.NewSendMessageAction),
	)

	if err := manifest.Validate(); err != nil {
		log.Fatal(err)
	}

	return manifest
}
