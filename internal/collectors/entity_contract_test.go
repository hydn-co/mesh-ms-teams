package collectors

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/hydn-co/mesh-sdk/pkg/catalog/entities"
	"github.com/hydn-co/mesh-sdk/pkg/catalog/spaces"
	"github.com/hydn-co/mesh-sdk/pkg/connector"
	"github.com/hydn-co/mesh-sdk/pkg/connectorutil"
	"github.com/hydn-co/substrate/json/polymorphic"
	"github.com/stretchr/testify/require"

	"github.com/hydn-co/mesh-ms-teams/internal/channels"
	"github.com/hydn-co/mesh-ms-teams/internal/credentials"
	"github.com/hydn-co/mesh-ms-teams/internal/options"
	"github.com/hydn-co/mesh-ms-teams/internal/teams"
)

type captureEntityEmitter struct {
	emitted []any
}

func (e *captureEntityEmitter) Emit(_ context.Context, entity any) error {
	e.emitted = append(e.emitted, entity)
	return nil
}

const fakeTeamsPageURL = "https://graph.microsoft.com/v1.0/teams?$skiptoken=page-2"

type fakeTeamsGraphClient struct{}

func (fakeTeamsGraphClient) ListTeams(_ context.Context) (*teams.ListTeamsResult, error) {
	return &teams.ListTeamsResult{
		OdataNextLink: fakeTeamsPageURL,
		Value: []teams.GraphTeam{
			{ID: "team-1", DisplayName: "Team One", Description: "First team"},
		},
	}, nil
}

func (fakeTeamsGraphClient) ListTeamsPage(_ context.Context, pageURL string) (*teams.ListTeamsResult, error) {
	if pageURL != fakeTeamsPageURL {
		panic("unexpected teams page URL: " + pageURL)
	}
	return &teams.ListTeamsResult{
		Value: []teams.GraphTeam{
			{ID: "team-2", DisplayName: "Team Two", Description: "Second team"},
		},
	}, nil
}

const fakeChannelsPageURL = "https://graph.microsoft.com/v1.0/teams/team-1/channels?$skiptoken=page-2"

type fakeChannelsGraphClient struct {
	fakeTeamsGraphClient
}

func (fakeChannelsGraphClient) ListChannels(_ context.Context, teamID string) (*channels.ListChannelsResult, error) {
	if teamID == "" {
		panic("team ID unexpectedly empty")
	}
	result := &channels.ListChannelsResult{
		Value: []channels.GraphChannel{
			{ID: "channel-" + teamID + "-1", DisplayName: "General", Description: "General channel"},
		},
	}
	if teamID == "team-1" {
		result.OdataNextLink = fakeChannelsPageURL
	}
	return result, nil
}

func (fakeChannelsGraphClient) ListChannelsPage(
	_ context.Context,
	pageURL string,
) (*channels.ListChannelsResult, error) {
	if pageURL != fakeChannelsPageURL {
		panic("unexpected channels page URL: " + pageURL)
	}
	return &channels.ListChannelsResult{
		Value: []channels.GraphChannel{
			{ID: "channel-team-1-2", DisplayName: "Announcements", Description: "Announcements channel"},
		},
	}, nil
}

func TestShouldOnlyEmitDeclaredEntityTypesWhenTeamsCollectorRunsWithInjectedClient(t *testing.T) {
	// Arrange
	emitter := &captureEntityEmitter{}
	collector := &TeamsCollector{
		TypedFeatureContext: newTeamsTestFeatureContext(t, emitter, &options.TeamsCollectorOptions{
			TeamsOptionsCore: options.TeamsOptionsCore{TenantID: "tenant-id"},
		}),
		newClient: func(_ context.Context, _ *credentials.AzureADCredentials) (teamsGraphClient, error) {
			return fakeTeamsGraphClient{}, nil
		},
	}

	// Act
	require.NoError(t, collector.Init(t.Context()))
	require.NoError(t, collector.Start(t.Context()))
	require.NoError(t, collector.Stop(t.Context()))

	// Assert
	assertEmittedEntityContract(t, emitter.emitted, []any{
		&entities.Group{},
	}, (&options.TeamsCollectorOptions{}).GetSpaces())
}

func TestShouldOnlyEmitDeclaredEntityTypesWhenChannelsCollectorRunsWithInjectedClient(t *testing.T) {
	// Arrange
	emitter := &captureEntityEmitter{}
	collector := &ChannelsCollector{
		TypedFeatureContext: newTeamsTestFeatureContext(t, emitter, &options.ChannelsCollectorOptions{
			TeamsOptionsCore: options.TeamsOptionsCore{TenantID: "tenant-id"},
		}),
		newClient: func(_ context.Context, _ *credentials.AzureADCredentials) (channelsGraphClient, error) {
			return fakeChannelsGraphClient{}, nil
		},
	}

	// Act
	require.NoError(t, collector.Init(t.Context()))
	require.NoError(t, collector.Start(t.Context()))
	require.NoError(t, collector.Stop(t.Context()))

	// Assert
	assertEmittedEntityContract(t, emitter.emitted, []any{
		&entities.Channel{},
	}, (&options.ChannelsCollectorOptions{}).GetSpaces())
}

func newTeamsTestFeatureContext[T connector.FeatureOptions](
	t *testing.T,
	emitter *captureEntityEmitter,
	featureOptions T,
) *connector.TypedFeatureContext[T, *connector.NoPayload] {
	t.Helper()

	grantCredential, err := json.Marshal(map[string]string{
		"client_id":     "client-id",
		"client_secret": "client-secret",
	})
	require.NoError(t, err)

	return connector.NewTypedFeatureContext[T, *connector.NoPayload](
		connector.NewFeatureContext(
			connector.WithConfiguration(&connector.Configuration{
				TenantID:    uuid.MustParse("11111111-1111-1111-1111-111111111111"),
				ConnectorID: uuid.MustParse("22222222-2222-2222-2222-222222222222"),
				Options:     polymorphic.NewEnvelope(featureOptions),
				Credentials: map[string]json.RawMessage{
					connectorutil.DefaultCredentialName: grantCredential,
				},
			}),
			connector.WithEmitter(emitter),
		),
	)
}

// assertEmittedEntityContract verifies the collector's emitted surface: only the
// allowed concrete entity types were emitted, and the observed spaces — derived
// from each entity's own GetSpace() accessor — exactly match the collector
// options' declared GetSpaces().
func assertEmittedEntityContract(t *testing.T, emitted []any, allowedTypes []any, expectedSpaces []spaces.Space) {
	t.Helper()

	require.NotEmpty(t, emitted, "expected at least one emitted entity")

	allowedTypeNames := make([]string, 0, len(allowedTypes))
	allowedTypeSet := map[string]struct{}{}
	for _, item := range allowedTypes {
		typeName := reflect.TypeOf(item).String()
		allowedTypeNames = append(allowedTypeNames, typeName)
		allowedTypeSet[typeName] = struct{}{}
	}

	observedTypeSet := map[string]struct{}{}
	observedSpaceSet := map[spaces.Space]struct{}{}
	for _, item := range emitted {
		typeName := reflect.TypeOf(item).String()
		if _, ok := allowedTypeSet[typeName]; !ok {
			t.Fatalf("unexpected emitted entity type %s", typeName)
		}
		observedTypeSet[typeName] = struct{}{}

		entity, ok := item.(entities.MeshEntity)
		if !ok {
			t.Fatalf("emitted entity type %s does not implement entities.MeshEntity", typeName)
		}
		observedSpaceSet[entity.GetSpace()] = struct{}{}
	}

	observedTypeNames := make([]string, 0, len(observedTypeSet))
	for typeName := range observedTypeSet {
		observedTypeNames = append(observedTypeNames, typeName)
	}

	observedSpaces := make([]spaces.Space, 0, len(observedSpaceSet))
	for space := range observedSpaceSet {
		observedSpaces = append(observedSpaces, space)
	}

	require.ElementsMatch(t, allowedTypeNames, observedTypeNames)
	require.ElementsMatch(t, expectedSpaces, observedSpaces)
}
