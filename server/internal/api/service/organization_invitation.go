package service

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/rimdian/rimdian/internal/api/common"
	"github.com/rimdian/rimdian/internal/api/dto"
	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rotisserie/eris"
)

func (svc *ServiceImpl) OrganizationInvitationConsume(ctx context.Context, consumeInvitationDTO *dto.OrganizationInvitationConsume) (result *dto.AccountLoginResult, code int, err error) {

	// decode token
	email, organizationID, code, err := entity.DecodeAndVerifyInvitationToken(consumeInvitationDTO.InvitationToken, svc.Config.SECRET_KEY)

	if err != nil {
		return nil, code, eris.Wrap(err, "OrganizationInvitationConsume")
	}

	// get invitation

	invitation, err := svc.Repo.GetInvitation(ctx, email, organizationID)

	if err != nil {
		return nil, 500, eris.Wrap(err, "OrganizationInvitationConsume")
	}

	if invitation == nil || invitation.ConsumedAt != nil {
		return nil, 400, entity.ErrInvitationHasBeenConsumedOrCancelled
	}

	if invitation.ExpiresAt.Before(time.Now()) {
		return nil, 400, entity.ErrInvitationHasExpired
	}

	var accountID string
	var newAccount *entity.Account

	// check if account token is not provided, create account account
	if consumeInvitationDTO.AccountID == nil {

		existing, err := svc.Repo.GetAccountFromEmail(ctx, email)

		// log.Printf("is ? %v", eris.Is(err, entity.ErrAccountNotFound))
		if err != nil && !eris.Is(err, entity.ErrAccountNotFound) {
			return nil, 500, eris.Wrapf(err, "OrganizationInvitationConsume, got err: %v", err)
		}

		if existing != nil {
			return nil, 400, entity.ErrAccountEmailAlreadyUsed
		}

		now := time.Now().UTC()

		newAccount = &entity.Account{
			FullName:         consumeInvitationDTO.Name,
			Timezone:         "UTC",
			Email:            email,
			IsServiceAccount: false,
			IsRoot:           false,
			CreatedAt:        now,
			UpdatedAt:        now,
		}

		// generate an ID
		saID, err := uuid.NewRandom()
		if err != nil {
			return nil, 500, eris.Wrap(err, "OrganizationInvitationConsume")
		}
		newAccount.ID = saID.String()
		accountID = newAccount.ID

		pwd, err := common.HashPassword(*consumeInvitationDTO.Password)
		if err != nil {
			return nil, 500, eris.Wrap(err, "OrganizationInvitationConsume error")
		}

		// Hash password
		newAccount.HashedPassword = pwd
	} else {

		// an account is already authenticated, check if account has same email

		account, err := svc.Repo.GetAccountFromID(ctx, *consumeInvitationDTO.AccountID)

		if err != nil {
			if eris.Is(err, entity.ErrAccountNotFound) {
				return nil, 400, eris.Wrapf(err, "OrganizationInvitationConsume, got account id %v", consumeInvitationDTO.AccountID)
			}
			return nil, 500, eris.Wrap(err, "OrganizationInvitationConsume")
		}

		if account.Email != invitation.Email {
			return nil, 400, entity.ErrAccountEmailDoesntMatchInvitationEmail
		}

		accountID = account.ID
	}

	if err := svc.Repo.ConsumeInvitation(ctx, accountID, newAccount, invitation); err != nil {
		if eris.Is(err, entity.ErrAccountAlreadyInOrganization) {
			return nil, 400, eris.Wrap(err, "OrganizationInvitationConsume")
		}
		if eris.Is(err, entity.ErrAccountEmailAlreadyUsed) {
			return nil, 400, eris.Wrap(err, "OrganizationInvitationConsume")
		}
		return nil, 500, eris.Wrap(err, "OrganizationInvitationConsume")
	}

	// login new account
	if newAccount != nil {
		result, code, err = svc.AccountLogin(ctx, &dto.AccountLogin{
			Email:     newAccount.Email,
			Password:  *consumeInvitationDTO.Password,
			UserAgent: consumeInvitationDTO.UserAgent,
			ClientIP:  consumeInvitationDTO.ClientIP,
		})

		if err != nil {
			return nil, code, eris.Wrap(err, "OrganizationInvitationConsume")
		}
	}

	return result, 200, nil
}

func (svc *ServiceImpl) OrganizationInvitationRead(ctx context.Context, token string) (result *dto.OrganizationInvitationReadResult, code int, err error) {

	email, orgID, code, err := entity.DecodeAndVerifyInvitationToken(token, svc.Config.SECRET_KEY)

	if err != nil {
		return nil, code, eris.Wrap(err, "OrganizationInvitationRead")
	}

	// fetch org
	organization, err := svc.Repo.GetOrganization(ctx, orgID)

	if err != nil {
		return nil, 500, eris.Wrap(err, "OrganizationInvitationRead")
	}

	result = &dto.OrganizationInvitationReadResult{
		Email:            email,
		OrganizationID:   orgID,
		OrganizationName: organization.Name,
	}

	return result, 200, nil
}

// invites an account into the organization
func (svc *ServiceImpl) OrganizationInvitationCreate(ctx context.Context, accountInvitationDTO *dto.OrganizationInvitation) (code int, err error) {

	if err := accountInvitationDTO.Validate(); err != nil {
		return 400, eris.Wrap(err, "OrganizationInvitationCreate")
	}

	// verify that token is owner of its organization
	isOwner, code, err := svc.IsOwnerOfOrganization(ctx, accountInvitationDTO.FromAccountID, accountInvitationDTO.OrganizationID)

	if err != nil {
		return code, eris.Wrap(err, "OrganizationInvitationCreate")
	}

	if !isOwner {
		return 401, eris.Wrap(entity.ErrAccountIsNotOwner, "OrganizationInvitationCreate")
	}

	// fetch organization

	organization, err := svc.Repo.GetOrganization(ctx, accountInvitationDTO.OrganizationID)

	if err != nil {
		return 500, eris.Wrap(err, "OrganizationInvitationCreate")
	}

	// invitation is base64 URL encoded email + organization ID

	token := entity.CreateInvitationToken(accountInvitationDTO.Email, accountInvitationDTO.OrganizationID, svc.Config.SECRET_KEY)

	// in dev we print the token in the console for testing purpose
	if svc.Config.ENV == entity.ENV_DEV {
		log.Printf("invitation token is: %v", token)
	}

	invitation := &entity.OrganizationInvitation{
		Email:            accountInvitationDTO.Email,
		OrganizationID:   accountInvitationDTO.OrganizationID,
		FromAccountID:    accountInvitationDTO.FromAccountID,
		ExpiresAt:        time.Now().UTC().AddDate(0, 0, 15), // invitation is valid for 15 days
		WorkspacesScopes: accountInvitationDTO.WorkspacesScopes,
	}

	// inserts invitation or updates expiration date if already exists
	if err := svc.Repo.UpsertOrganizationInvitation(ctx, invitation); err != nil {
		if eris.Is(err, entity.ErrInvitationConsumedOrDeleted) {
			return 400, err
		} else {
			return 500, eris.Wrap(err, "OrganizationInvitationCreate")
		}
	}

	// end token by email

	if err := svc.SendSystemEmail(ctx, &dto.SystemEmail{
		To:   []string{accountInvitationDTO.Email},
		Kind: dto.EMAIL_ORGANIZATION_INVITE,
		EmailOrganizationInvitationCreatePayload: dto.EmailOrganizationInvitationCreatePayload{
			OrganizationID:   organization.ID,
			OrganizationName: organization.Name,
			ActionURL:        svc.Config.API_ENDPOINT + "/accept-invitation?token=" + token,
		},
	}); err != nil {
		return 500, eris.Wrap(err, "OrganizationInvitationCreate")
	}

	return 201, nil
}

// list invitations for organization
func (svc *ServiceImpl) OrganizationInvitationList(ctx context.Context, accountID string, organizationID string) (result *dto.OrganizationInvitationListResult, code int, err error) {

	// verify that token is account of its organization
	isAccount, code, err := svc.IsAccountOfOrganization(ctx, accountID, organizationID)

	if err != nil {
		return nil, code, eris.Wrap(err, "OrganizationInvitationList")
	}

	if !isAccount {
		return nil, 400, eris.New("account is not part of the organization")
	}

	invitations, err := svc.Repo.ListInvitationsForOrganization(ctx, organizationID)

	if err != nil {
		return nil, 500, eris.Wrap(err, "OrganizationInvitationList")
	}

	result = &dto.OrganizationInvitationListResult{
		Invitations: []*entity.OrganizationInvitation{},
	}

	result.Invitations = append(result.Invitations, invitations...)

	return result, 200, nil
}

func (svc *ServiceImpl) OrganizationInvitationCancel(ctx context.Context, accountID string, deleteInvitation *dto.OrganizationInvitationCancel) (code int, err error) {

	// verify that token is owner of its organization
	isOwner, code, err := svc.IsOwnerOfOrganization(ctx, accountID, deleteInvitation.OrganizationID)

	if err != nil {
		return code, eris.Wrap(err, "OrganizationInvitationCancel")
	}

	if !isOwner {
		return 401, eris.Wrap(entity.ErrAccountIsNotOwner, "OrganizationInvitationCancel")
	}

	if err := svc.Repo.CancelOrganizationInvitation(ctx, deleteInvitation.OrganizationID, deleteInvitation.Email); err != nil {
		if eris.Is(err, entity.ErrInvitationConsumedOrDeleted) {
			return 400, eris.Wrap(err, "OrganizationInvitationCancel")
		} else {
			return 500, eris.Wrap(err, "OrganizationInvitationCancel")
		}
	}

	return 200, nil
}
