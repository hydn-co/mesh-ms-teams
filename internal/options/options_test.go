package options_test

import (
	"testing"

	"github.com/hydn-co/mesh-ms-teams/internal/options"
	"github.com/stretchr/testify/assert"
)

func TestShouldHaveUniqueDiscriminators(t *testing.T) {
	// Arrange
	discriminators := []string{
		(&options.TeamsCollectorOptions{}).GetDiscriminator(),
		(&options.ChannelsCollectorOptions{}).GetDiscriminator(),
		(&options.SendMessageActionOptions{}).GetDiscriminator(),
	}

	// Act
	seen := make(map[string]bool)
	for _, d := range discriminators {
		// Assert
		assert.NotEmpty(t, d)
		assert.False(t, seen[d], "duplicate discriminator: %s", d)
		seen[d] = true
	}
}

func TestShouldReturnSpacesWhenTeamsCollectorOptionsGetSpaces(t *testing.T) {
	// Arrange
	opts := &options.TeamsCollectorOptions{}

	// Act
	spaces := opts.GetSpaces()

	// Assert
	assert.NotEmpty(t, spaces)
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
