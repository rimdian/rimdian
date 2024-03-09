package repository

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rotisserie/eris"
)

func (repo *RepositoryImpl) AddColumn(ctx context.Context, workspace *entity.Workspace, tableName string, column *entity.TableColumn) error {

	conn, err := repo.GetWorkspaceConnection(ctx, workspace.ID)

	if err != nil {
		return err
	}

	defer conn.Close()

	query := fmt.Sprintf("ALTER TABLE `%v` ADD ", tableName)

	columnDDL, err := tableColumnDDL(column)

	if err != nil {
		return err
	}
	query = query + columnDDL

	// naive SQL injection prevention
	query = strings.ReplaceAll(query, ";", "")
	query = strings.ReplaceAll(query, "DELETE", "")
	query = strings.ReplaceAll(query, "UPDATE", "")
	query = strings.ReplaceAll(query, "SELECT", "")

	query = query + ";"

	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return eris.Wrap(err, "AddColumn")
	}

	if _, err := tx.ExecContext(ctx, query); err != nil {
		shouldRollback := true

		// check if error Error Duplicate column name
		if strings.Contains(err.Error(), "Duplicate column name") {
			// check that the table is the same
			err = repo.IsExistingColumnTheSame(ctx, workspace.ID, tableName, column)

			if err == nil {
				shouldRollback = false
			} else {
				err = eris.Wrapf(err, "existing column in table %v is different: %v", tableName, column.Name)
			}
		}

		if shouldRollback {
			if rollErr := tx.Rollback(); rollErr != nil {
				log.Printf("rollback error %v", rollErr)
			}
			return eris.Wrapf(err, "AddColumn, query: %v", query)
		}
	}

	// update workspace
	if err := repo.UpdateWorkspace(ctx, workspace, tx); err != nil {
		tx.Rollback()
		return eris.Wrap(err, "AddColumn")
	}

	if err := tx.Commit(); err != nil {
		return eris.Wrap(err, "AddColumn")
	}

	return nil
}

func (repo *RepositoryImpl) DeleteColumn(ctx context.Context, workspace *entity.Workspace, tableName string, column *entity.TableColumn) error {

	conn, err := repo.GetWorkspaceConnection(ctx, workspace.ID)

	if err != nil {
		return err
	}

	defer conn.Close()

	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return eris.Wrap(err, "DeleteColumn")
	}

	query := fmt.Sprintf("ALTER TABLE `%v` DROP COLUMN %v", tableName, column.Name)

	_, err = tx.ExecContext(ctx, query)

	// ignore if table doesnt exist
	if err != nil && !(strings.Contains(err.Error(), "Error 1146: Table") && strings.Contains(err.Error(), "doesn't exist")) {
		tx.Rollback()
		return eris.Wrapf(err, "DeleteColumn, query: %v", query)
	}

	// update workspace
	if err := repo.UpdateWorkspace(ctx, workspace, tx); err != nil {
		tx.Rollback()
		return eris.Wrap(err, "DeleteColumn")
	}

	if err := tx.Commit(); err != nil {
		return eris.Wrap(err, "DeleteColumn")
	}

	return nil
}

func (repo *RepositoryImpl) IsExistingColumnTheSame(ctx context.Context, workspaceID string, tableName string, columnManifest *entity.TableColumn) error {

	conn, err := repo.GetWorkspaceConnection(ctx, workspaceID)

	if err != nil {
		return err
	}

	defer conn.Close()

	// TODO: refactor columns comparison with IsExistingTableTheSame
	// only fetch schema_info for the column we are interested in

	return nil
}
