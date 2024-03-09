package dto

import "github.com/rimdian/rimdian/internal/api/entity"

// create / update workspace "Data Transfer Object"
type WorkspaceCreate struct {
	ID                  string `json:"id"`
	Name                string `json:"name"`
	WebsiteURL          string `json:"website_url"`
	PrivacyPolicyURL    string `json:"privacy_policy_url"`
	Industry            string `json:"industry"`
	Currency            string `json:"currency"`
	OrganizationID      string `json:"organization_id"`
	DefaultUserTimezone string `json:"default_user_timezone"`
	DefaultUserCountry  string `json:"default_user_country"`
	DefaultUserLanguage string `json:"default_user_language"`
}

type WorkspaceUpdate struct {
	*WorkspaceCreate
	DataProtectionOfficerID string       `json:"dpo_id"`
	UserReconciliationKeys  []string     `json:"user_reconciliation_keys"`
	UserIDSigning           string       `json:"user_id_signing"`
	SessionTimeout          int          `json:"session_timeout"`
	HasOrders               bool         `json:"has_orders"`
	HasLeads                bool         `json:"has_leads"`
	LeadStages              []*LeadStage `json:"lead_stages"`
	LicenseKey              *string      `json:"license_key,omitempty"`
}

type WorkspaceListResult struct {
	Workspaces []*entity.Workspace `json:"workspaces"`
}

type WorkspaceShowResult struct {
	Workspace *entity.Workspace `json:"workspace"`
}

type WorkspaceShowTablesResult struct {
	Tables []*entity.TableInformationSchema `json:"tables"`
}

type WorkspaceCreateOrResetDemo struct {
	OrganizationID string `json:"organization_id"`
	Kind           string `json:"kind"` // order / lead
}

type WorkspaceGetSecretKey struct {
	ID string `json:"id"`
}

type WorkspaceSecretKeyResult struct {
	SecretKey string `json:"secret_key"`
}
