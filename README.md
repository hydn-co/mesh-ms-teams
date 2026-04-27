# mesh-ms-teams

A mesh connector for Microsoft Teams integration. Implements standardized mesh collectors
and actions using [mesh-sdk](https://github.com/hydn-co/mesh-sdk) to receive commands,
emit catalog entities, and perform actions on Microsoft Teams via the Microsoft Graph API.

## Collectors

### `teams_collector`

Collects all teams accessible to the service principal and emits them as catalog Group
entities in the `activity` space. Supports pagination via `@odata.nextLink`.

| Option | Type | Required | Description |
|--------|------|----------|-------------|
| `include_archived` | `bool` | No | Include archived teams |

### `channels_collector`

Collects channels from a specified team and emits them as catalog Channel entities in the
`activity` space.

| Option | Type | Required | Description |
|--------|------|----------|-------------|
| `team_id` | `string` | **Yes** | ID of the team to collect channels from |

## Actions

### `send_message_action`

Posts a message to a Microsoft Teams channel.

| Option | Type | Required | Description |
|--------|------|----------|-------------|
| `team_id` | `string` | **Yes** | ID of the team containing the target channel |
| `channel_id` | `string` | **Yes** | ID of the channel to post the message to |

| Payload | Type | Required | Description |
|---------|------|----------|-------------|
| `message` | `string` | **Yes** | Message content (max 4000 characters, UTF-8) |

## Azure AD app registration

All features authenticate via OAuth 2.0 client credentials (app-only flow). The service
principal must be granted the required Microsoft Graph application permissions.

### Setup steps

1. Go to [portal.azure.com](https://portal.azure.com) → **Azure Active Directory** → **App registrations**
2. Click **New registration**, name the app (e.g. `mesh-ms-teams`), and leave the redirect URI blank
3. Click **Register**, then note the **Application (client) ID** and **Directory (tenant) ID**
4. Under **Certificates & secrets** → **Client secrets**, click **New client secret** and copy the value
5. Under **API permissions** → **Add a permission** → **Microsoft Graph** → **Application permissions**, add:

| Permission | Required by |
|------------|-------------|
| `Team.ReadBasic.All` | `teams_collector` |
| `Channel.ReadBasic.All` | `channels_collector` |
| `ChannelMessage.Send` | `send_message_action` |

6. Click **Grant admin consent** for your organization

### Credential payload

```json
{
  "tenant_id":     "your-tenant-id",
  "client_id":     "your-client-id",
  "client_secret": "your-client-secret"
}
```

## Requirements

- Go 1.25+

## Quick start

```bash
git clone https://github.com/hydn-co/mesh-ms-teams.git
cd mesh-ms-teams
go test ./...
go build ./...
```

## Usage

Generate the feature manifest:

```bash
go run ./cmd -describe
```

List registered features:

```bash
go run ./cmd -list
```
