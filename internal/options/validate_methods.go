package options

import "github.com/hydn-co/mesh-sdk/pkg/connectorutil"

func validateTeamsOptionsCore(o *TeamsOptionsCore) error {
	if o == nil {
		_, err := connectorutil.RequireTrimmedString("tenant_id", "")
		return err
	}

	tenantID, err := connectorutil.RequireTrimmedString("tenant_id", o.GetTenantID())
	if err != nil {
		return err
	}

	o.TenantID = tenantID
	return nil
}

func (o *TeamsCollectorOptions) Validate() error {
	if o == nil {
		return validateTeamsOptionsCore(nil)
	}

	return validateTeamsOptionsCore(&o.TeamsOptionsCore)
}

func (o *ChannelsCollectorOptions) Validate() error {
	if o == nil {
		return validateTeamsOptionsCore(nil)
	}

	return validateTeamsOptionsCore(&o.TeamsOptionsCore)
}

func (o *SendMessageActionOptions) Validate() error {
	if o == nil {
		return validateTeamsOptionsCore(nil)
	}

	if err := validateTeamsOptionsCore(&o.TeamsOptionsCore); err != nil {
		return err
	}

	teamID, err := connectorutil.RequireTrimmedString("team_id", o.TeamID)
	if err != nil {
		return err
	}

	channelID, err := connectorutil.RequireTrimmedString("channel_id", o.ChannelID)
	if err != nil {
		return err
	}

	o.TeamID = teamID
	o.ChannelID = channelID
	return nil
}
