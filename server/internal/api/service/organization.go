package service

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/rimdian/rimdian/internal/api/common"
	"github.com/rimdian/rimdian/internal/api/dto"
	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rotisserie/eris"
)

func (svc *ServiceImpl) OrganizationCreate(ctx context.Context, orgCreateDTO *dto.OrganizationCreate) (result *dto.OrganizationResult, code int, err error) {

	// only for Cloud Managed CM
	if !svc.Config.MANAGED_RMD {
		return nil, 403, eris.New("Forbidden")
	}

	org := &entity.Organization{
		ID:                      orgCreateDTO.ID,
		Name:                    orgCreateDTO.Name,
		Currency:                orgCreateDTO.Currency,
		DataProtectionOfficerID: "root",
	}

	// validate
	if err := org.Validate(); err != nil {
		return nil, 400, eris.Wrap(err, "OrganizationCreate")
	}

	code, err = svc.Repo.RunInTransactionForSystem(ctx, func(ctx context.Context, tx *sql.Tx) (code int, err error) {

		if err := svc.Repo.CreateOrganization(ctx, org, tx); err != nil {
			return 500, eris.Wrap(err, "OrganizationCreate")
		}

		workspaceScopes := entity.WorkspacesScopes{
			{
				WorkspaceID: "*",
				Scopes:      []string{"*"},
			},
		}

		if err := svc.Repo.AddAccountToOrganization(ctx, "root", org.ID, true, nil, workspaceScopes, tx); err != nil {
			return 500, eris.Wrap(err, "OrganizationCreate")
		}

		return 201, nil
	})

	if err != nil {
		return nil, code, eris.Wrap(err, "OrganizationCreate")
	}

	org.CreatedAt = time.Now()
	org.UpdatedAt = org.CreatedAt

	return &dto.OrganizationResult{
		ID:                      org.ID,
		Name:                    org.Name,
		Currency:                org.Currency,
		DataProtectionOfficerID: org.DataProtectionOfficerID,
		CreatedAt:               org.CreatedAt,
		UpdatedAt:               org.UpdatedAt,
		ImOwner:                 true,
	}, 200, nil
}

// invites an account into the organization
func (svc *ServiceImpl) OrganizationSetProfile(ctx context.Context, accountID string, profileDTO *dto.OrganizationProfile) (updatedOrg *dto.OrganizationResult, code int, err error) {

	// verify that token is owner of its organization
	isOwner, code, err := svc.IsOwnerOfOrganization(ctx, accountID, profileDTO.ID)

	if err != nil {
		return nil, code, eris.Wrap(err, "OrganizationSetProfile")
	}

	if !isOwner {
		return nil, 401, eris.Wrap(entity.ErrAccountIsNotOwner, "OrganizationSetProfile")
	}

	// fetch organization

	organization, err := svc.Repo.GetOrganization(ctx, profileDTO.ID)

	if err != nil {
		return nil, 500, eris.Wrap(err, "OrganizationSetProfile")
	}

	profileDTO.Name = strings.TrimSpace(profileDTO.Name)

	if profileDTO.Name == "" {
		return nil, 400, eris.Wrap(entity.ErrOrganizationNameRequired, "OrganizationSetProfile")
	}

	if !govalidator.IsIn(profileDTO.Currency, common.CurrenciesCodes...) {
		return nil, 400, eris.Wrap(entity.ErrOrganizationInvalidCurrency, "OrganizationSetProfile")
	}

	organization.Name = profileDTO.Name
	organization.Currency = profileDTO.Currency
	organization.AccountIsOwner = true // joined field

	if err := svc.Repo.UpdateOrganizationProfile(ctx, organization); err != nil {
		return nil, 500, eris.Wrap(err, "OrganizationSetProfile")
	}

	return &dto.OrganizationResult{
		ID:                      organization.ID,
		Name:                    organization.Name,
		Currency:                organization.Currency,
		DataProtectionOfficerID: organization.DataProtectionOfficerID,
		CreatedAt:               organization.CreatedAt,
		UpdatedAt:               organization.UpdatedAt,
		ImOwner:                 true,
	}, 200, nil
}

// list organizations for account
func (svc *ServiceImpl) OrganizationList(ctx context.Context, accountID string) (result *dto.OrganizationListResult, code int, err error) {

	orgs, err := svc.Repo.ListOrganizationsForAccount(ctx, accountID)

	if err != nil {
		return nil, 500, eris.Wrap(err, "OrganizationList")
	}

	result = &dto.OrganizationListResult{
		Organizations: []*dto.OrganizationResult{},
	}

	for _, org := range orgs {
		result.Organizations = append(result.Organizations, &dto.OrganizationResult{
			ID:                      org.ID,
			Name:                    org.Name,
			Currency:                org.Currency,
			DataProtectionOfficerID: org.DataProtectionOfficerID,
			CreatedAt:               org.CreatedAt,
			UpdatedAt:               org.UpdatedAt,
			ImOwner:                 org.AccountIsOwner,
		})
	}

	return result, 200, nil
}
