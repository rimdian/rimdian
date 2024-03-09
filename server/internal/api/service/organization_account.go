package service

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rimdian/rimdian/internal/api/common"
	"github.com/rimdian/rimdian/internal/api/dto"
	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rotisserie/eris"
)

func (svc *ServiceImpl) IsAccountOfOrganization(ctx context.Context, accountID string, organizationID string) (isAccount bool, code int, err error) {

	accountID = strings.TrimSpace(accountID)
	organizationID = strings.TrimSpace(organizationID)

	if accountID == "" {
		return false, 400, eris.Wrap(entity.ErrInvalidAccountID, "IsAccountOfOrganization")
	}

	if organizationID == "" {
		return false, 400, entity.ErrInvalidOrganizationID
	}

	isAccount, err = svc.Repo.IsAccountOfOrganization(ctx, accountID, organizationID, false)

	if err != nil {
		return false, 500, eris.Wrap(err, "IsAccountOfOrganization")
	}

	return isAccount, 200, nil
}

func (svc *ServiceImpl) IsOwnerOfOrganization(ctx context.Context, accountID string, organizationID string) (isOwner bool, code int, err error) {

	accountID = strings.TrimSpace(accountID)
	organizationID = strings.TrimSpace(organizationID)

	if accountID == "" {
		return false, 400, eris.Wrap(entity.ErrInvalidAccountID, "IsOwnerOfOrganization")
	}

	if organizationID == "" {
		return false, 400, entity.ErrInvalidOrganizationID
	}

	isOwner, err = svc.Repo.IsAccountOfOrganization(ctx, accountID, organizationID, true)

	if err != nil {
		return false, 500, eris.Wrap(err, "IsOwnerOfOrganization")
	}

	return isOwner, 200, nil
}

// create a service account Account and adds it to the organization
func (svc *ServiceImpl) OrganizationAccountCreateServiceAccount(ctx context.Context, accountID string, createServiceAccount *dto.OrganizationAccountCreateServiceAccount) (code int, err error) {

	if err := createServiceAccount.Validate(); err != nil {
		return 400, eris.Wrap(err, "OrganizationAccountCreateServiceAccount")
	}

	// verify that token is owner of its organization
	isOwner, code, err := svc.IsOwnerOfOrganization(ctx, accountID, createServiceAccount.OrganizationID)

	if err != nil {
		return code, eris.Wrap(err, "OrganizationAccountCreateServiceAccount")
	}

	if !isOwner {
		return 401, eris.Wrap(entity.ErrAccountIsNotOwner, "OrganizationAccountCreateServiceAccount")
	}

	// list workspaces for organization
	workspaces, err := svc.Repo.ListWorkspaces(ctx, &createServiceAccount.OrganizationID)

	if err != nil {
		return 500, eris.Wrap(err, "OrganizationAccountCreateServiceAccount")
	}

	createServiceAccount.Name = strings.TrimSpace(createServiceAccount.Name)
	createServiceAccount.EmailID = strings.TrimSpace(createServiceAccount.EmailID)
	createServiceAccount.Password = strings.TrimSpace(createServiceAccount.Password)

	if createServiceAccount.Name == "" {
		return 400, entity.ErrServiceAccountNameIsRequired
	}
	if createServiceAccount.EmailID == "" {
		return 400, entity.ErrServiceAccountEmailIDIsRequired
	}
	if len(createServiceAccount.Password) < 16 {
		return 400, entity.ErrServiceAccountPasswordIsInvalid
	}

	// extract hostname from API endpoint
	u, err := url.Parse(svc.Config.API_ENDPOINT)
	if err != nil {
		return 500, eris.Wrap(err, "OrganizationAccountCreateServiceAccount")
	}

	email := fmt.Sprintf("%v.%v@%v", createServiceAccount.EmailID, createServiceAccount.OrganizationID, u.Hostname())
	now := time.Now().UTC()

	serviceAccount := &entity.Account{
		FullName:         &createServiceAccount.Name,
		Timezone:         "UTC",
		Email:            email,
		IsServiceAccount: true,
		IsRoot:           false,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	// generate an ID
	saID, err := uuid.NewRandom()
	if err != nil {
		return 500, eris.Wrap(err, "OrganizationAccountCreateServiceAccount")
	}
	serviceAccount.ID = saID.String()

	pwd, err := common.HashPassword(createServiceAccount.Password)
	if err != nil {
		return 500, eris.Wrap(err, "OrganizationAccountCreateServiceAccount error")
	}

	// Hash password
	serviceAccount.HashedPassword = pwd

	// log.Printf("serviceAccount %+v\n", serviceAccount)
	workspaceScopes := entity.WorkspacesScopes{}

	for _, ws := range createServiceAccount.WorkspacesScopes {
		// check workspace exists
		for _, w := range workspaces {
			if w.ID == ws.WorkspaceID {
				workspaceScopes = append(workspaceScopes, &entity.WorkspaceScope{
					WorkspaceID: ws.WorkspaceID,
					Scopes:      ws.Scopes,
				})
			}
		}
	}

	if err := svc.Repo.InsertServiceAccount(ctx, serviceAccount, createServiceAccount.OrganizationID, accountID, workspaceScopes); err != nil {
		return 500, eris.Wrap(err, "OrganizationAccountCreateServiceAccount")
	}

	return 200, nil
}

// deactivates account access to organization
func (svc *ServiceImpl) OrganizationAccountDeactivate(ctx context.Context, accountID string, deactivateAccountDTO *dto.OrganizationAccountDeactivate) (code int, err error) {

	// verify that token is owner of its organization
	isOwner, code, err := svc.IsOwnerOfOrganization(ctx, accountID, deactivateAccountDTO.OrganizationID)

	if err != nil {
		return code, eris.Wrap(err, "OrganizationAccountDeactivate")
	}

	if !isOwner {
		return 401, eris.Wrap(entity.ErrAccountIsNotOwner, "OrganizationAccountDeactivate")
	}

	// find account to deactivate

	accounts, err := svc.Repo.ListAccountsForOrganization(ctx, deactivateAccountDTO.OrganizationID)

	if err != nil {
		return 500, eris.Wrap(err, "OrganizationAccountDeactivate")
	}

	var accountFound *entity.AccountWithOrganizationRole

	for _, account := range accounts {
		if account.ID == deactivateAccountDTO.DeactivateAccountID && account.DeactivatedAt == nil {
			accountFound = account
		}
	}

	if accountFound == nil {
		return 400, entity.ErrDeactivateInvalidAccount
	}

	// deactivate account
	if err := svc.Repo.DeactivateOrganizationAccount(ctx, accountID, deactivateAccountDTO.DeactivateAccountID, deactivateAccountDTO.OrganizationID); err != nil {
		return 500, eris.Wrap(err, "OrganizationAccountDeactivate")
	}

	return 200, nil
}

// transfers organization ownership to another account
func (svc *ServiceImpl) OrganizationAccountTransferOwnership(ctx context.Context, accountID string, transferOwnership *dto.OrganizationAccountTransferOwnership) (code int, err error) {

	// verify that token is owner of its organization
	isOwner, code, err := svc.IsOwnerOfOrganization(ctx, accountID, transferOwnership.OrganizationID)

	if err != nil {
		return code, eris.Wrap(err, "OrganizationAccountTransferOwnership")
	}

	if !isOwner {
		return 401, eris.Wrap(entity.ErrAccountIsNotOwner, "OrganizationAccountTransferOwnership")
	}

	// find destination account

	accounts, err := svc.Repo.ListAccountsForOrganization(ctx, transferOwnership.OrganizationID)

	if err != nil {
		return 500, eris.Wrap(err, "OrganizationAccountTransferOwnership")
	}

	var accountFound *entity.AccountWithOrganizationRole

	for _, account := range accounts {
		if account.ID == transferOwnership.ToAccountID && !account.IsServiceAccount && account.DeactivatedAt == nil {
			accountFound = account
		}
	}

	if accountFound == nil {
		return 400, entity.ErrTransferOwnershipToInvalidAccount
	}

	// update accounts
	if err := svc.Repo.TransferOrganizationOwnsership(ctx, accountID, transferOwnership.ToAccountID, transferOwnership.OrganizationID); err != nil {
		return 500, eris.Wrap(err, "OrganizationAccountTransferOwnership")
	}

	return 200, nil
}

// list accounts for organization
func (svc *ServiceImpl) OrganizationAccountList(ctx context.Context, accountID string, organizationID string) (result *dto.OrganizationAccountListResult, code int, err error) {

	// verify that token is account of its organization
	isAccount, code, err := svc.IsAccountOfOrganization(ctx, accountID, organizationID)

	if err != nil {
		return nil, code, eris.Wrap(err, "OrganizationAccountList")
	}

	if !isAccount {
		return nil, 400, eris.New("account is not part of the organization")
	}

	accounts, err := svc.Repo.ListAccountsForOrganization(ctx, organizationID)

	if err != nil {
		return nil, 500, eris.Wrap(err, "OrganizationAccountList")
	}

	result = &dto.OrganizationAccountListResult{
		Accounts: []*entity.AccountWithOrganizationRole{},
	}

	result.Accounts = append(result.Accounts, accounts...)

	return result, 200, nil
}
