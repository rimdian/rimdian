package dto

var EMAIL_RESET_PASSWORD = "resetPassword"
var EMAIL_ORGANIZATION_INVITE = "organizationInvite"
var EMAIL_OBSERVABILITY_INCIDENT = "observabilityIncident"

type SystemEmail struct {
	To                                       []string
	Kind                                     string
	EmailResetPasswordPayload                EmailResetPasswordPayload
	EmailOrganizationInvitationCreatePayload EmailOrganizationInvitationCreatePayload
	EmailObservabilityIncidentsPayload       EmailObservabilityIncidentPayload
}

type EmailResetPasswordPayload struct {
	FullName  *string
	ActionURL string
}

type EmailOrganizationInvitationCreatePayload struct {
	OrganizationID   string
	OrganizationName string
	ActionURL        string
}

type EmailObservabilityIncidentPayload struct {
	WorkspaceID   string
	WorkspaceName string
	Incident      string
}
