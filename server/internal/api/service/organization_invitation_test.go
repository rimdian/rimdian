package service

import (
	"context"
	"testing"
	"time"

	"github.com/rimdian/rimdian/internal/api/common"
	"github.com/rimdian/rimdian/internal/api/dto"
	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rimdian/rimdian/internal/api/repository"
	"github.com/rotisserie/eris"
)

func TestServiceImpl_OrganizationInvitationConsume(t *testing.T) {

	t.Run("should reject consumed or cancelled invitation", func(t *testing.T) {

		email := "invited@account.com"
		orgID := "acme"

		token := entity.CreateInvitationToken(email, orgID, secretKey)

		repoMock := &repository.RepositoryMock{
			GetInvitationFunc: func(ctx context.Context, email, organizationID string) (*entity.OrganizationInvitation, error) {
				return nil, entity.ErrInvitationHasBeenConsumedOrCancelled
			},
		}

		svc := &ServiceImpl{
			Config: &entity.Config{SECRET_KEY: secretKey},
			Repo:   repoMock,
			Mailer: nil,
		}

		_, _, err := svc.OrganizationInvitationConsume(context.Background(), &dto.OrganizationInvitationConsume{
			InvitationToken: token,
		})

		if !eris.Is(err, entity.ErrInvitationHasBeenConsumedOrCancelled) {
			t.Errorf("should fail on consumed or cancelled invitation, got: %v", err)
		}
	})

	t.Run("should reject expired invitation", func(t *testing.T) {

		email := "invited@account.com"
		orgID := "acme"

		token := entity.CreateInvitationToken(email, orgID, secretKey)

		repoMock := &repository.RepositoryMock{
			GetInvitationFunc: func(ctx context.Context, email, organizationID string) (*entity.OrganizationInvitation, error) {
				return &entity.OrganizationInvitation{
					Email:          email,
					OrganizationID: orgID,
					FromAccountID:  "root",
					ExpiresAt:      time.Now().AddDate(0, 0, -1),
				}, nil
			},
		}

		svc := &ServiceImpl{
			Config: &entity.Config{SECRET_KEY: secretKey},
			Repo:   repoMock,
			Mailer: nil,
		}

		_, _, err := svc.OrganizationInvitationConsume(context.Background(), &dto.OrganizationInvitationConsume{
			InvitationToken: token,
		})

		if !eris.Is(err, entity.ErrInvitationHasExpired) {
			t.Errorf("should fail on expired invitation, got: %v", err)
		}
	})

	t.Run("should create account if has no account yet", func(t *testing.T) {

		email := "invited@account.com"
		name := "John"
		pass := "123"
		orgID := "acme"

		token := entity.CreateInvitationToken(email, orgID, secretKey)
		hashedPassword, _ := common.HashPassword(pass)

		firstCall := true

		repoMock := &repository.RepositoryMock{
			GetInvitationFunc: func(ctx context.Context, email, organizationID string) (*entity.OrganizationInvitation, error) {
				return &entity.OrganizationInvitation{
					Email:          email,
					OrganizationID: orgID,
					FromAccountID:  "root",
					ExpiresAt:      time.Now().AddDate(0, 0, 1),
				}, nil
			},
			GetAccountFromEmailFunc: func(ctx context.Context, email string) (*entity.Account, error) {

				// no account found on first call, to create account
				if firstCall {
					firstCall = false
					return nil, entity.ErrAccountNotFound
				}

				// then we return the account
				return &entity.Account{
					ID:             "uuid",
					Email:          email,
					HashedPassword: hashedPassword,
				}, nil
			},
			ConsumeInvitationFunc: func(ctx context.Context, accountID string, insertAccount *entity.Account, invitation *entity.OrganizationInvitation) error {
				return nil
			},
			InsertAccountSessionFunc: func(ctx context.Context, accountSession *entity.AccountSession) error {
				return nil
			},
		}

		svc := &ServiceImpl{
			Config: &entity.Config{SECRET_KEY: secretKey},
			Repo:   repoMock,
			Mailer: nil,
		}

		_, code, err := svc.OrganizationInvitationConsume(context.Background(), &dto.OrganizationInvitationConsume{
			InvitationToken: token,
			Name:            &name,
			Password:        &pass,
		})

		if err != nil {
			t.Errorf("got err %v: %v", code, err)
		}

		// called consume invitation repo once
		if len(repoMock.ConsumeInvitationCalls()) != 1 {
			t.Error("should call Repo.ConsumeInvitation one time")
		}

		// check inserted account
		if repoMock.ConsumeInvitationCalls()[0].InsertAccount == nil {
			t.Error("insertAccount should not be nil")
		}

		// TODO: check that returns a login refresh token
	})

	t.Run("should use account if already authenticated", func(t *testing.T) {

		accountID := "existing"
		email := "invited@account.com"
		orgID := "acme"

		token := entity.CreateInvitationToken(email, orgID, secretKey)

		repoMock := &repository.RepositoryMock{
			GetInvitationFunc: func(ctx context.Context, email, organizationID string) (*entity.OrganizationInvitation, error) {
				return &entity.OrganizationInvitation{
					Email:          email,
					OrganizationID: orgID,
					FromAccountID:  "root",
					ExpiresAt:      time.Now().AddDate(0, 0, 1),
				}, nil
			},
			GetAccountFromIDFunc: func(ctx context.Context, id string) (*entity.Account, error) {
				return &entity.Account{
					ID:    id,
					Email: email,
				}, nil
			},
			ConsumeInvitationFunc: func(ctx context.Context, accountID string, insertAccount *entity.Account, invitation *entity.OrganizationInvitation) error {
				return nil
			},
		}

		svc := &ServiceImpl{
			Config: &entity.Config{SECRET_KEY: secretKey},
			Repo:   repoMock,
			Mailer: nil,
		}

		_, _, err := svc.OrganizationInvitationConsume(context.Background(), &dto.OrganizationInvitationConsume{
			InvitationToken: token,
			AccountID:       &accountID, // already authenticated account
		})

		if err != nil {
			t.Errorf("got err: %v", err)
		}

		// called consume invitation repo once
		if len(repoMock.ConsumeInvitationCalls()) != 1 {
			t.Error("should call Repo.ConsumeInvitation one time")
		}

		if repoMock.ConsumeInvitationCalls()[0].AccountID != accountID {
			t.Errorf("account id should be %v, got %v", accountID, repoMock.ConsumeInvitationCalls()[0].AccountID)
		}

		if repoMock.ConsumeInvitationCalls()[0].InsertAccount != nil {
			t.Error("insertAccount should be nil")
		}
	})
}

func TestServiceImpl_OrganizationInvitationRead(t *testing.T) {
	// TODO
}

func TestServiceImpl_OrganizationInvitationCreate(t *testing.T) {
	// TODO
}

func TestServiceImpl_OrganizationInvitationList(t *testing.T) {
	// TODO
}

func TestServiceImpl_OrganizationInvitationCancel(t *testing.T) {
	// TODO
}
