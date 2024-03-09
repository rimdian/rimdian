package repository

import (
	"context"
	"strings"

	"github.com/georgysavva/scany/v2/sqlscan"
	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rotisserie/eris"
)

// Reset DB in dev env only
func (repo *RepositoryImpl) DevResetDB(ctx context.Context, rootAccount *entity.Account, defaultOrganization *entity.Organization) (err error) {

	conn, err := repo.GetSystemConnection(ctx)

	if err != nil {
		return err
	}

	defer conn.Close()

	// list workspaces to delete
	var workspaceIDs []string

	err = sqlscan.Select(ctx, conn, &workspaceIDs, "SELECT id FROM workspace")

	if err != nil {
		// ignore error if table doesnt exist
		if !strings.Contains(err.Error(), "Error 1146: Table") {
			return eris.Wrap(err, "DevResetDB select workspaces")
		}
	}

	for _, id := range workspaceIDs {
		if _, err := conn.ExecContext(ctx, "DROP DATABASE IF EXISTS "+repo.Config.DB_PREFIX+id); err != nil {
			return eris.Wrapf(err, "DevResetDB drop workspace db %v", repo.Config.DB_PREFIX+id)
		}

	}

	// delete system DB
	if _, err := conn.ExecContext(ctx, "DROP DATABASE IF EXISTS "+repo.Config.DB_PREFIX+repo.GetSystemDB()); err != nil {
		return eris.Wrapf(err, "DevResetDB drop system db %v", repo.Config.DB_PREFIX+repo.GetSystemDB())
	}

	return repo.Install(ctx, rootAccount, defaultOrganization)
}
