package payloads

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/fgrzl/json/polymorphic"
)

func init() {
	polymorphic.RegisterType[SendMessagePayload]()
	polymorphic.RegisterType[ProvisionUserPayload]()
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

	if len(trimmed) > 4000 {
		return fmt.Errorf("message exceeds maximum length of 4000 characters (got %d)", len(trimmed))
	}

	if !utf8.ValidString(trimmed) {
		return fmt.Errorf("message contains invalid UTF-8")
	}

	return nil
}

// ProvisionUserPayload is the action payload schema for provisioning a new user in Microsoft Teams.
type ProvisionUserPayload struct {
	// DisplayName is the display name for the new user.
	DisplayName string `json:"display_name" binding:"required" description:"Display name for the new user"`

	// UserPrincipalName is the user principal name (UPN) for the new account, e.g. user@domain.com.
	UserPrincipalName string `json:"user_principal_name" binding:"required" description:"User principal name (UPN) for the new account"`

	// MailNickname is the mail alias for the new user.
	MailNickname string `json:"mail_nickname" binding:"required" description:"Mail alias for the new user"`

	// Password is the initial password for the new account.
	Password string `json:"password" binding:"required" description:"Initial password for the new account"`
}

func (p *ProvisionUserPayload) GetDiscriminator() string {
	return "mesh://ms-teams/payloads/provision-user"
}

// Validate ensures the provision-user payload is valid.
func (p *ProvisionUserPayload) Validate() error {
	if strings.TrimSpace(p.DisplayName) == "" {
		return fmt.Errorf("display_name cannot be empty")
	}
	if strings.TrimSpace(p.UserPrincipalName) == "" {
		return fmt.Errorf("user_principal_name cannot be empty")
	}
	if strings.TrimSpace(p.MailNickname) == "" {
		return fmt.Errorf("mail_nickname cannot be empty")
	}
	if strings.TrimSpace(p.Password) == "" {
		return fmt.Errorf("password cannot be empty")
	}
	return nil
}
