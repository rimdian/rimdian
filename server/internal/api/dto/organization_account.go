package dto

import (
	"fmt"
	"strings"

	"github.com/rimdian/rimdian/internal/api/entity"
)

type OrganizationAccountListResult struct {
	Accounts []*entity.AccountWithOrganizationRole `json:"accounts"`
}

type OrganizationAccountCreateServiceAccount struct {
	OrganizationID   string                   `json:"organization_id"`
	Name             string                   `json:"name"`
	EmailID          string                   `json:"email_id"`
	Password         string                   `json:"password"`
	WorkspacesScopes []*entity.WorkspaceScope `json:"workspaces_scopes"`
}

func (data *OrganizationAccountCreateServiceAccount) Validate() error {
	data.OrganizationID = strings.TrimSpace(data.OrganizationID)
	data.Name = strings.TrimSpace(data.Name)
	data.EmailID = strings.TrimSpace(data.EmailID)
	data.Password = strings.TrimSpace(data.Password)

	if data.OrganizationID == "" {
		return fmt.Errorf("organization_id is required")
	}
	if data.Name == "" {
		return fmt.Errorf("name is required")
	}
	if data.EmailID == "" {
		return fmt.Errorf("email_id is required")
	}
	if data.Password == "" {
		return fmt.Errorf("password is required")
	}
	if len(data.WorkspacesScopes) == 0 {
		return fmt.Errorf("workspaces_scopes is required")
	}
	for _, scope := range data.WorkspacesScopes {
		if err := scope.Validate(); err != nil {
			return err
		}
	}
	return nil
}

type OrganizationAccountTransferOwnership struct {
	OrganizationID string `json:"organization_id"`
	ToAccountID    string `json:"to_account_id"`
}

type OrganizationAccountDeactivate struct {
	OrganizationID      string `json:"organization_id"`
	DeactivateAccountID string `json:"deactivate_account_id"`
}
