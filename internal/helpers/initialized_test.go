package helpers_test

import (
	"testing"

	"github.com/hydn-co/mesh-ms-teams/internal/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckInitialized_WhenTrue_ReturnsNil(t *testing.T) {
	// Arrange
	initiated := true

	// Act
	err := helpers.CheckInitialized(initiated)

	// Assert
	require.NoError(t, err)
}

func TestCheckInitialized_WhenFalse_ReturnsError(t *testing.T) {
	// Arrange
	initiated := false

	// Act
	err := helpers.CheckInitialized(initiated)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")
}
