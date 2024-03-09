package dto

type AccountProfile struct {
	FullName string `json:"full_name"`
	Timezone string `json:"timezone"`
	Locale   string `json:"locale"`
}

type AccountResetPassword struct {
	Email string `json:"email"`
}

type AccountConsumeResetPassword struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
	UserAgent   string `json:"user_agent"`
	ClientIP    string `json:"client_ip"`
}

// provide account token if already has account, otherwise creates an account
type OrganizationInvitationConsume struct {
	InvitationToken string  `json:"token"`
	Name            *string `json:"name,omitempty"`
	Password        *string `json:"password,omitempty"`
	UserAgent       string  `json:"user_agent"`
	ClientIP        string  `json:"client_ip"`
	// added server side after account token authentication
	AccountID *string `json:"account_id,omitempty"`
}
