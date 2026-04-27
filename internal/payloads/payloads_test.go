package payloads_test

import (
	"testing"

	"github.com/hydn-co/mesh-ms-teams/internal/payloads"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShouldReturnDiscriminatorWhenSendMessagePayloadGetDiscriminator(t *testing.T) {
	// Arrange
	p := &payloads.SendMessagePayload{}

	// Act
	discriminator := p.GetDiscriminator()

	// Assert
	assert.NotEmpty(t, discriminator)
}

func TestShouldValidateSuccessfullyWhenMessageIsValid(t *testing.T) {
	// Arrange
	p := &payloads.SendMessagePayload{Message: "Hello, Teams!"}

	// Act
	err := p.Validate()

	// Assert
	require.NoError(t, err)
}

func TestShouldReturnErrorWhenMessageIsEmpty(t *testing.T) {
	// Arrange
	p := &payloads.SendMessagePayload{Message: ""}

	// Act
	err := p.Validate()

	// Assert
	assert.Error(t, err)
}

func TestShouldReturnErrorWhenMessageIsWhitespaceOnly(t *testing.T) {
	// Arrange
	p := &payloads.SendMessagePayload{Message: "   "}

	// Act
	err := p.Validate()

	// Assert
	assert.Error(t, err)
}

func TestShouldReturnErrorWhenMessageExceedsMaxLength(t *testing.T) {
	// Arrange
	msg := make([]byte, 4001)
	for i := range msg {
		msg[i] = 'a'
	}
	p := &payloads.SendMessagePayload{Message: string(msg)}

	// Act
	err := p.Validate()

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "exceeds maximum length")
}

func TestShouldValidateSuccessfullyWhenMessageIsAtMaxLength(t *testing.T) {
	// Arrange
	msg := make([]byte, 4000)
	for i := range msg {
		msg[i] = 'a'
	}
	p := &payloads.SendMessagePayload{Message: string(msg)}

	// Act
	err := p.Validate()

	// Assert
	require.NoError(t, err)
}

func TestShouldReturnDiscriminatorWhenProvisionUserPayloadGetDiscriminator(t *testing.T) {
	// Arrange
	p := &payloads.ProvisionUserPayload{}

	// Act
	discriminator := p.GetDiscriminator()

	// Assert
	assert.NotEmpty(t, discriminator)
}

func TestShouldValidateSuccessfullyWhenProvisionUserPayloadIsValid(t *testing.T) {
	// Arrange
	p := &payloads.ProvisionUserPayload{
		DisplayName:       "Test User",
		UserPrincipalName: "testuser@example.com",
		MailNickname:      "testuser",
		Password:          "P@ssw0rd!",
	}

	// Act
	err := p.Validate()

	// Assert
	require.NoError(t, err)
}

func TestShouldReturnErrorWhenDisplayNameIsEmpty(t *testing.T) {
	// Arrange
	p := &payloads.ProvisionUserPayload{
		UserPrincipalName: "testuser@example.com",
		MailNickname:      "testuser",
		Password:          "P@ssw0rd!",
	}

	// Act
	err := p.Validate()

	// Assert
	assert.ErrorContains(t, err, "display_name")
}

func TestShouldReturnErrorWhenUserPrincipalNameIsEmpty(t *testing.T) {
	// Arrange
	p := &payloads.ProvisionUserPayload{
		DisplayName:  "Test User",
		MailNickname: "testuser",
		Password:     "P@ssw0rd!",
	}

	// Act
	err := p.Validate()

	// Assert
	assert.ErrorContains(t, err, "user_principal_name")
}

func TestShouldReturnErrorWhenPasswordIsEmpty(t *testing.T) {
	// Arrange
	p := &payloads.ProvisionUserPayload{
		DisplayName:       "Test User",
		UserPrincipalName: "testuser@example.com",
		MailNickname:      "testuser",
	}

	// Act
	err := p.Validate()

	// Assert
	assert.ErrorContains(t, err, "password")
}
