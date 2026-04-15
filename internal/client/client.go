package client

import "encoding/json"

// Credentials holds the OAuth 2.0 client credentials for the Microsoft Teams API.
type Credentials struct {
	TenantID     string `json:"tenant_id"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

// ParseCredentials deserialises raw JSON credentials into a Credentials struct.
func ParseCredentials(raw json.RawMessage) (*Credentials, error) {
	var c Credentials
	if err := json.Unmarshal(raw, &c); err != nil {
		return nil, err
	}
	return &c, nil
}

// Client is a stub MS Teams Graph API client.
type Client struct {
	creds *Credentials
}

// New creates a new Client from the provided credentials.
func New(creds *Credentials) *Client {
	return &Client{creds: creds}
}
