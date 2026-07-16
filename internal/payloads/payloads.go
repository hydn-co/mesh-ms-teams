package payloads

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/hydn-co/substrate/json/polymorphic"
)

func init() {
	polymorphic.RegisterType[SendMessagePayload]()
}

// SendMessagePayload is the action payload schema for sending a message to a Teams channel.
type SendMessagePayload struct {
	// Message is the content of the message to send.
	Message string `json:"message" binding:"required" title:"Message" description:"Message body to post to the target Microsoft Teams channel. Maximum 4000 characters."`
}

func (p *SendMessagePayload) GetDiscriminator() string {
	return "mesh://ms-teams/actions/ms_teams_send_message_action_payload"
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

	p.Message = trimmed

	return nil
}
