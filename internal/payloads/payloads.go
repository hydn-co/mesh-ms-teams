package payloads

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/fgrzl/json/polymorphic"
)

func init() {
	polymorphic.RegisterType[SendMessagePayload]()
}

// SendMessagePayload is the action payload schema for sending a message to a Teams channel.
type SendMessagePayload struct {
	// Message is the content of the message to send.
	Message string `json:"message" binding:"required" description:"Message content to send"`
}

func (p *SendMessagePayload) GetDiscriminator() string {
	return "mesh://ms-teams/payloads/send-message"
}

// Validate ensures the message payload is valid.
func (p *SendMessagePayload) Validate() error {
	if p.Message == "" {
		return fmt.Errorf("message cannot be empty")
	}

	trimmed := strings.TrimSpace(p.Message)
	if trimmed == "" {
		return fmt.Errorf("message cannot be empty")
	}

	if utf8.RuneCountInString(trimmed) > 4000 {
		return fmt.Errorf("message exceeds maximum length of 4000 characters (got %d)",
			utf8.RuneCountInString(trimmed))
	}

	if !utf8.ValidString(trimmed) {
		return fmt.Errorf("message contains invalid UTF-8")
	}

	return nil
}
