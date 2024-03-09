package dto

type ChannelGroup struct {
	ID          string `json:"id"`
	WorkspaceID string `json:"workspace_id"`
	Name        string `json:"name"`
	Color       string `json:"color"`
}

type Channel struct {
	ID           string         `json:"id"`
	WorkspaceID  string         `json:"workspace_id"`
	Name         string         `json:"name"`
	Origins      []*Origin      `json:"origins"`
	VoucherCodes []*VoucherCode `json:"voucher_codes"`
	GroupID      string         `json:"group_id"`
}

type VoucherCode struct {
	Code           string  `json:"code"`
	OriginID       string  `json:"origin_id"`
	SetUTMCampaign *string `json:"set_utm_campaign,omitempty"` // utm_campaign
	SetUTMContent  *string `json:"set_utm_content,omitempty"`  // utm_content
	Description    *string `json:"description,omitempty"`
}

type Origin struct {
	ID            string  `json:"id"`             // source / medium / campaign
	MatchOperator string  `json:"match_operator"` // only "equals" at the moment
	UTMSource     string  `json:"utm_source"`
	UTMMedium     string  `json:"utm_medium"`
	UTMCampaign   *string `json:"utm_campaign,omitempty"`
}

type DeleteChannel struct {
	ID          string `json:"id"`
	WorkspaceID string `json:"workspace_id"`
}

type DeleteChannelGroup struct {
	ID          string `json:"id"`
	WorkspaceID string `json:"workspace_id"`
}
