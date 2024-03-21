package dto

import (
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rotisserie/eris"
)

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

type FilesSettingsUpdate struct {
	Endpoint    string `json:"endpoint"`
	AccessKey   string `json:"access_key"`
	SecretKey   string `json:"secret_key"`
	Bucket      string `json:"bucket"`
	Region      string `json:"region"`
	CDNEndpoint string `json:"cdn_endpoint"`
}

func (x *FilesSettingsUpdate) Validate() error {
	// trim spaces
	x.Endpoint = strings.TrimSpace(x.Endpoint)
	x.AccessKey = strings.TrimSpace(x.AccessKey)
	x.Bucket = strings.TrimSpace(x.Bucket)
	x.Region = strings.TrimSpace(x.Region)
	x.CDNEndpoint = strings.TrimSpace(x.CDNEndpoint)

	if x.Endpoint != "" {
		if govalidator.IsURL(x.Endpoint) == false {
			return eris.New("files settings endpoint is not valid")
		}
		if x.AccessKey == "" {
			return eris.New("files settings access key is required")
		}
		if x.Bucket == "" {
			return eris.New("files settings bucket is required")
		}
		if govalidator.IsRequestURI(x.CDNEndpoint) == false {
			return eris.New("files settings cdn_endpoint is not valid")
		}
	}
	return nil
}

type WorkspaceUpdate struct {
	*WorkspaceCreate
	DataProtectionOfficerID string              `json:"dpo_id"`
	UserReconciliationKeys  []string            `json:"user_reconciliation_keys"`
	UserIDSigning           string              `json:"user_id_signing"`
	SessionTimeout          int                 `json:"session_timeout"`
	HasOrders               bool                `json:"has_orders"`
	HasLeads                bool                `json:"has_leads"`
	LeadStages              []*LeadStage        `json:"lead_stages"`
	LicenseKey              *string             `json:"license_key,omitempty"`
	FilesSettings           FilesSettingsUpdate `json:"files_settings"`
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
