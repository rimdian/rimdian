package dto

import (
	"fmt"
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/rimdian/rimdian/internal/api/entity"
)

type OrganizationInvitation struct {
	OrganizationID   string                  `json:"organization_id"`
	FromAccountID    string                  `json:"from_account_id"`
	Email            string                  `json:"email"`
	WorkspacesScopes entity.WorkspacesScopes `json:"workspaces_scopes"`
}

func (data *OrganizationInvitation) Validate() error {
	data.OrganizationID = strings.TrimSpace(data.OrganizationID)
	data.FromAccountID = strings.TrimSpace(data.FromAccountID)
	data.Email = strings.TrimSpace(data.Email)

	if data.OrganizationID == "" {
		return fmt.Errorf("organization_id is required")
	}
	if data.FromAccountID == "" {
		return fmt.Errorf("from_account_id is required")
	}
	if !govalidator.IsEmail(data.Email) {
		return fmt.Errorf("email is invalid")
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

type OrganizationInvitationCancel struct {
	OrganizationID string `json:"organization_id"`
	Email          string `json:"email"`
}

type OrganizationInvitationResend struct {
	OrganizationID string `json:"organization_id"`
	Email          string `json:"email"`
}

type OrganizationInvitationRead struct {
	Token string `json:"token"`
}

type OrganizationInvitationReadResult struct {
	Email            string `json:"email"`
	OrganizationID   string `json:"organization_id"`
	OrganizationName string `json:"organization_name"`
}

type OrganizationInvitationListResult struct {
	Invitations []*entity.OrganizationInvitation `json:"invitations"`
}
