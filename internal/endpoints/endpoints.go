package endpoints

// Microsoft Graph API v1.0 endpoints
const (
	// Users endpoints
	GraphUsersListEndpoint     = "https://graph.microsoft.com/v1.0/users"
	GraphProvisionUserEndpoint = "https://graph.microsoft.com/v1.0/users"

	// Teams endpoints
	GraphTeamsListEndpoint = "https://graph.microsoft.com/v1.0/groups?$filter=resourceProvisioningOptions/any(x:x%20eq%20%27Team%27)"

	// Channels endpoints
	GraphChannelsListEndpoint = "https://graph.microsoft.com/v1.0/teams/{teamId}/channels"

	// Messages endpoints
	GraphSendMessageEndpoint = "https://graph.microsoft.com/v1.0/teams/{teamId}/channels/{channelId}/messages"

	// OAuth token endpoint
	GraphTokenEndpoint = "https://login.microsoftonline.com/{tenantId}/oauth2/v2.0/token"
)
