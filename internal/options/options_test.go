package options_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hydn-co/mesh-ms-teams/internal/options"
)

func TestShouldReturnExpectedDiscriminators(t *testing.T) {
	assert.Equal(
		t,
		"mesh://ms-teams/collectors/ms_teams_teams_collector_options",
		(&options.TeamsCollectorOptions{}).GetDiscriminator(),
	)
	assert.Equal(
		t,
		"mesh://ms-teams/collectors/ms_teams_channels_collector_options",
		(&options.ChannelsCollectorOptions{}).GetDiscriminator(),
	)
	assert.Equal(
		t,
		"mesh://ms-teams/actions/ms_teams_send_message_action_options",
		(&options.SendMessageActionOptions{}).GetDiscriminator(),
	)
}

func TestShouldReturnGroupsSpaceWhenTeamsCollectorOptionsGetSpaces(t *testing.T) {
	// Arrange
	opts := &options.TeamsCollectorOptions{}

	// Act
	spaces := opts.GetSpaces()

	// Assert
	assert.Contains(t, spaces, "groups")
}

func TestShouldReturnChannelsSpaceWhenChannelsCollectorOptionsGetSpaces(t *testing.T) {
	// Arrange
	opts := &options.ChannelsCollectorOptions{}

	// Act
	spaces := opts.GetSpaces()

	// Assert
	assert.Contains(t, spaces, "channels")
}

func TestShouldReturnChannelsSpaceWhenSendMessageActionOptionsGetSpaces(t *testing.T) {
	// Arrange
	opts := &options.SendMessageActionOptions{}

	// Act
	spaces := opts.GetSpaces()

	// Assert
	assert.Contains(t, spaces, "channels")
}

func TestShouldReturnRequirementsWhenTeamsCollectorOptionsGetRequirements(t *testing.T) {
	// Arrange
	opts := &options.TeamsCollectorOptions{}

	// Act
	reqs := opts.GetRequirements()

	// Assert
	assert.NotEmpty(t, reqs)
}

func TestShouldReturnRequirementsWhenChannelsCollectorOptionsGetRequirements(t *testing.T) {
	// Arrange
	opts := &options.ChannelsCollectorOptions{}

	// Act
	reqs := opts.GetRequirements()

	// Assert
	assert.NotEmpty(t, reqs)
}

func TestShouldReturnRequirementsWhenSendMessageActionOptionsGetRequirements(t *testing.T) {
	// Arrange
	opts := &options.SendMessageActionOptions{}

	// Act
	reqs := opts.GetRequirements()

	// Assert
	assert.NotEmpty(t, reqs)
}

func TestShouldValidateTeamsCollectorOptionsWhenTenantIDPresent(t *testing.T) {
	opts := &options.TeamsCollectorOptions{
		TeamsOptionsCore: options.TeamsOptionsCore{TenantID: "  tenant-id  "},
	}

	require.NoError(t, opts.Validate())
	assert.Equal(t, "tenant-id", opts.TenantID)
}

func TestShouldReturnErrorWhenTeamsCollectorOptionsMissingTenantID(t *testing.T) {
	err := (&options.TeamsCollectorOptions{}).Validate()

	require.Error(t, err)
	assert.EqualError(t, err, "tenant_id is required")
}

func TestShouldValidateAndTrimSendMessageActionOptions(t *testing.T) {
	opts := &options.SendMessageActionOptions{
		TeamsOptionsCore: options.TeamsOptionsCore{TenantID: " tenant-id "},
		TeamID:           " team-id ",
		ChannelID:        " channel-id ",
	}

	require.NoError(t, opts.Validate())
	assert.Equal(t, "tenant-id", opts.TenantID)
	assert.Equal(t, "team-id", opts.TeamID)
	assert.Equal(t, "channel-id", opts.ChannelID)
	assert.Contains(t, opts.GetRequirements(), "teams")
}

func TestShouldReturnErrorWhenSendMessageActionOptionsMissingChannelID(t *testing.T) {
	opts := &options.SendMessageActionOptions{
		TeamsOptionsCore: options.TeamsOptionsCore{TenantID: "tenant-id"},
		TeamID:           "team-id",
	}

	err := opts.Validate()

	require.Error(t, err)
	assert.EqualError(t, err, "channel_id is required")
}
