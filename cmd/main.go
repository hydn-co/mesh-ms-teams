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
		"Mesh integration with Microsoft Teams",
	)

	// Register Teams Collector
	if err := manifest.RegisterFeature(
		"teams_collector",
		"Teams Collector",
		"Collects teams from Microsoft Teams and emits them as catalog entities.",
		true,
		runner.FeatureTypeCollector,
		new(options.TeamsCollectorOptions),
		nil,
		runner.FeatureResumeBehaviorNone,
		runner.GrantCredential,
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
		runner.FeatureResumeBehaviorNone,
		runner.GrantCredential,
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
		runner.GrantCredential,
		runner.Factory(actions.NewSendMessageAction),
	); err != nil {
		log.Fatal(err)
	}

	if err := manifest.Validate(); err != nil {
		log.Fatal(err)
	}

	return manifest
}
