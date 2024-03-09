package entity

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/rotisserie/eris"
)

var (
	ErrAccountIsNotOwner                 = eris.New("account is not owner of this organization")
	ErrDeactivateInvalidAccount          = eris.New("deactivate invalid account")
	ErrTransferOwnershipToInvalidAccount = eris.New("transfer ownership to invalid account")
)

type WorkspaceScope struct {
	WorkspaceID string   `json:"workspace_id"`
	Scopes      []string `json:"scopes"`
}

func (data *WorkspaceScope) Validate() error {
	data.WorkspaceID = strings.TrimSpace(data.WorkspaceID)

	if data.WorkspaceID == "" {
		return fmt.Errorf("workspace_id is required")
	}

	// only "*"" scope is implemented for now
	for _, scope := range data.Scopes {
		if scope != "*" {
			return fmt.Errorf("invalid scope %s", scope)
		}
	}

	return nil
}

type WorkspacesScopes []*WorkspaceScope

func (x *WorkspacesScopes) Scan(val interface{}) error {

	var data []byte

	if b, ok := val.([]byte); ok {
		// VERY IMPORTANT: we need to clone the bytes here
		// The sql driver will reuse the same bytes RAM slots for future queries
		// Thank you St Antoine De Padoue for helping me find this bug
		data = bytes.Clone(b)
	} else if s, ok := val.(string); ok {
		data = []byte(s)
	} else if val == nil {
		return nil
	}

	return json.Unmarshal(data, x)
}

func (x WorkspacesScopes) Value() (driver.Value, error) {
	return json.Marshal(x)
}

type OrganizationAccount struct {
	AccountID        string           `db:"account_id" json:"account_id"`
	OrganizationID   string           `db:"organization_id" json:"organization_id"`
	IsOwner          bool             `db:"is_owner" json:"is_owner"`
	FromAccountID    string           `db:"from_account_id" json:"from_account_id"`
	WorkspacesScopes WorkspacesScopes `db:"workspaces_scopes" json:"workspaces_scopes"`
	CreatedAt        time.Time        `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time        `db:"updated_at" json:"updated_at"`
	DeactivatedAt    *time.Time       `db:"deactivated_at" json:"deactivated_at"`
}

type AccountWithOrganizationRole struct {
	*Account
	// joined at query
	IsOwner          bool             `db:"is_owner" json:"is_owner"`
	WorkspacesScopes WorkspacesScopes `db:"workspaces_scopes" json:"workspaces_scopes"`
	DeactivatedAt    *time.Time       `db:"deactivated_at" json:"deactivated_at,omitempty"`
}

var OrganizationAccountSchema string = `CREATE ROWSTORE TABLE IF NOT EXISTS organization_account (
	account_id VARCHAR(128) NOT NULL,
	organization_id VARCHAR(128) NOT NULL,
	is_owner BOOLEAN DEFAULT FALSE,
	from_account_id VARCHAR(128),
	workspaces_scopes JSON NOT NULL,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
	deactivated_at DATETIME,
	
	PRIMARY KEY (account_id, organization_id),
    SHARD KEY (account_id, organization_id)
  ) COLLATE utf8mb4_general_ci;
`

var OrganizationAccountSchemaMYSQL string = `CREATE TABLE IF NOT EXISTS organization_account (
	account_id VARCHAR(128) NOT NULL,
	organization_id VARCHAR(128) NOT NULL,
	is_owner BOOLEAN DEFAULT FALSE,
	from_account_id VARCHAR(128),
	workspaces_scopes JSON NOT NULL,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
	deactivated_at DATETIME,
	
	PRIMARY KEY (account_id, organization_id)
    -- SHARD KEY (account_id, organization_id)
  ) COLLATE utf8mb4_general_ci;
`
