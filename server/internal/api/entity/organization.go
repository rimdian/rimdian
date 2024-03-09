package entity

import (
	"strings"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/iancoleman/strcase"
	"github.com/rimdian/rimdian/internal/api/common"
	"github.com/rotisserie/eris"
)

var (
	OrganizationDefaultCurrency string = "USD"

	ErrInvalidOrganizationID       = eris.New("invalid organizationId")
	ErrOrganizationIDRequired      = eris.New("organization id is required")
	ErrOrganizationNameRequired    = eris.New("organization name is required")
	ErrOrganizationInvalidCurrency = eris.New("the currency is not valid")
	ErrOrganizationDPOInvalid      = eris.New("organization dataProtectionOfficerId is not valid")
	ErrOrganizationNotFound        = eris.New("organization not found")
	ErrOrganizationCurrency        = eris.New("organization currency is not valid")
)

type Organization struct {
	ID                      string     `db:"id" json:"id"`
	Name                    string     `db:"name" json:"name"`
	Currency                string     `db:"currency" json:"currency"`
	DataProtectionOfficerID string     `db:"dpo_id" json:"dpo_id"`
	CreatedAt               time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt               time.Time  `db:"updated_at" json:"updated_at"`
	DeletedAt               *time.Time `db:"deleted_at" json:"deleted_at,omitempty"`

	// joined fields
	AccountIsOwner          bool             `db:"account_is_owner" json:"account_is_owner,omitempty"`
	AccountWorkspacesScopes WorkspacesScopes `db:"account_workspaces_scopes" json:"account_workspaces_scopes,omitempty"`
}

func (org *Organization) Validate() error {
	org.ID = strings.TrimSpace(org.ID)
	org.Name = strings.TrimSpace(org.Name)

	if org.ID == "" {
		return ErrOrganizationIDRequired
	}
	if org.Name == "" {
		return ErrOrganizationNameRequired
	}
	if !govalidator.IsIn(org.Currency, common.CurrenciesCodes...) {
		return ErrOrganizationCurrency
	}

	if org.DataProtectionOfficerID == "" {
		return ErrOrganizationDPOInvalid
	}

	return nil
}

func GenerateDefaultOrganization(id string, name string) (org *Organization, err error) {

	// id is in lowercase without underscores (ex: My-Company_ID -> mycompanyid)
	id = strings.ReplaceAll(strcase.ToSnake(id), "_", "")
	name = strings.TrimSpace(name)

	if id == "" {
		return nil, eris.Wrapf(ErrOrganizationIDRequired, "GenerateDefaultOrganization id: %v", id)
	}

	if name == "" {
		return nil, eris.Wrapf(ErrOrganizationNameRequired, "GenerateDefaultOrganization name: %v", name)
	}

	now := time.Now().UTC()

	return &Organization{
		ID:                      id,
		Name:                    name,
		Currency:                OrganizationDefaultCurrency,
		DataProtectionOfficerID: "root",
		CreatedAt:               now,
		UpdatedAt:               now,
	}, nil
}

var OrganizationSchema string = `CREATE ROWSTORE TABLE IF NOT EXISTS organization (
	id VARCHAR(128) NOT NULL,
	name VARCHAR(50) NOT NULL,
	currency VARCHAR(3) NOT NULL,
	dpo_id VARCHAR(128) NOT NULL,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
	deleted_at DATETIME,
	
	PRIMARY KEY (id),
    SHARD KEY (id)
  ) COLLATE utf8mb4_general_ci;
`

var OrganizationSchemaMYSQL string = `CREATE TABLE IF NOT EXISTS organization (
	id VARCHAR(128) NOT NULL,
	name VARCHAR(50) NOT NULL,
	currency VARCHAR(3) NOT NULL,
	dpo_id VARCHAR(128) NOT NULL,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
	deleted_at DATETIME,
	
	PRIMARY KEY (id)
    -- SHARD KEY (id)
  ) COLLATE utf8mb4_general_ci;
`
