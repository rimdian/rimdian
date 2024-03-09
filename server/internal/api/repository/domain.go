package repository

import (
	"context"

	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rotisserie/eris"
)

// replace domain_id in all concerned tables for provided workspace ID
func (repo *RepositoryImpl) DeleteDomain(ctx context.Context, workspace *entity.Workspace, deletedDomainID string, migrateToDomainID string) error {

	conn, err := repo.GetWorkspaceConnection(ctx, workspace.ID)

	if err != nil {
		return err
	}

	defer conn.Close()

	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return eris.Wrap(err, "DeleteDomain")
	}

	// update workspace
	if err := repo.UpdateWorkspace(ctx, workspace, tx); err != nil {
		tx.Rollback()
		return eris.Wrap(err, "DeleteDomain")
	}

	// TODO: update timelines that had this domain_id

	if err := tx.Commit(); err != nil {
		return eris.Wrap(err, "DeleteDomain")
	}

	return nil
}
