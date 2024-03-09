package service

import (
	"context"
	"strings"
	"time"

	"github.com/georgysavva/scany/v2/sqlscan"
	"github.com/rimdian/rimdian/internal/api/dto"
	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rotisserie/eris"
)

var ErrDeleteDomainIDRequired = eris.New("delete domain id required")
var ErrDeleteDomainMigrationID = eris.New("delete domain: migrate_to_domain_id should be differant than id")
var ErrDeleteDomainInvalidID = eris.New("delete domain: invalid id")
var ErrDeleteDomainInvalidMigrationID = eris.New("delete domain: invalid migrate_to_domain_id")

func (svc *ServiceImpl) DomainDelete(ctx context.Context, accountID string, domainDeleteDTO *dto.DomainDelete) (updatedWorkspace *entity.Workspace, code int, err error) {

	workspace, code, err := svc.GetWorkspaceForAccount(ctx, domainDeleteDTO.WorkspaceID, accountID)

	if err != nil {
		return nil, code, eris.Wrap(err, "DomainDelete")
	}

	// validation
	domainDeleteDTO.ID = strings.TrimSpace(domainDeleteDTO.ID)
	domainDeleteDTO.MigrateToDomainID = strings.TrimSpace(domainDeleteDTO.MigrateToDomainID)

	if domainDeleteDTO.ID == domainDeleteDTO.MigrateToDomainID {
		return nil, 400, eris.Wrap(ErrDeleteDomainMigrationID, "DomainDelete")
	}

	now := time.Now().UTC()

	var deletedDomain *entity.Domain
	var migrateToDomain *entity.Domain

	for _, d := range workspace.Domains {
		if d.ID == domainDeleteDTO.ID && d.DeletedAt == nil {
			d.DeletedAt = &now
			deletedDomain = d
		}

		if d.ID == domainDeleteDTO.MigrateToDomainID {
			migrateToDomain = d
		}
	}

	if deletedDomain == nil {
		return nil, 400, eris.Wrap(ErrDeleteDomainInvalidID, "DomainDelete")
	}

	if migrateToDomain == nil {
		return nil, 400, eris.Wrap(ErrDeleteDomainInvalidMigrationID, "DomainDelete")
	}

	if deletedDomain != nil {
		if err := svc.Repo.DeleteDomain(ctx, workspace, deletedDomain.ID, migrateToDomain.ID); err != nil {
			return nil, 500, eris.Wrap(err, "DomainDelete")
		}
	}

	return workspace, 200, nil
}

func (svc *ServiceImpl) DomainUpsert(ctx context.Context, accountID string, domainDTO *dto.Domain) (updatedWorkspace *entity.Workspace, code int, err error) {

	workspace, code, err := svc.GetWorkspaceForAccount(ctx, domainDTO.WorkspaceID, accountID)

	if err != nil {
		return nil, code, eris.Wrap(err, "DomainUpsert")
	}

	now := time.Now().UTC()

	// convert DTO to entity
	domain := &entity.Domain{
		ID:              domainDTO.ID,
		Type:            domainDTO.Type,
		Name:            domainDTO.Name,
		Hosts:           domainDTO.Hosts,
		ParamsWhitelist: domainDTO.ParamsWhitelist,
		// HomepagePaths:   domainDTO.HomepagePaths,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := domain.Validate(); err != nil {
		return nil, 400, eris.Wrap(err, "DomainUpsert")
	}

	// upsert domain in workspace

	isInsert := true

	for _, d := range workspace.Domains {
		if d.ID == domain.ID {
			// is update
			d.Name = domain.Name
			if d.Type == entity.DomainWeb {
				d.Hosts = domain.Hosts
				d.ParamsWhitelist = domain.ParamsWhitelist
				// d.HomepagePaths = domain.HomepagePaths
			}
			d.UpdatedAt = domain.UpdatedAt
			isInsert = false
		}
	}

	if isInsert {
		workspace.Domains = append(workspace.Domains, domain)
	}

	if err := workspace.Validate(); err != nil {
		return nil, 400, eris.Wrap(err, "DomainUpsert")
	}

	if err := svc.Repo.UpdateWorkspace(ctx, workspace, nil); err != nil {
		if sqlscan.NotFound(err) {
			return nil, 400, err
		}
		return nil, 500, eris.Wrap(err, "DomainUpsert")
	}

	return workspace, 200, nil
}
