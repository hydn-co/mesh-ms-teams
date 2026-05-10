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
		"ms_teams_teams_collector",
		"Collect Teams",
		"Collects teams and emits them as group catalog entities.",
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
		"ms_teams_channels_collector",
		"Collect Channels",
		"Collects channels across all accessible teams and emits them as channel catalog entities.",
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
		"ms_teams_send_message_action",
		"Send Message",
		"Posts a message to a Microsoft Teams channel.",
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
