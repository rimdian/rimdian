package service

import (
	"context"
	"math"

	"github.com/georgysavva/scany/v2/sqlscan"
	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rimdian/rimdian/internal/api/migration"
	"github.com/rotisserie/eris"
)

var (
	migrations = map[float64]entity.MajorMigrationInterface{
		// add new migrations here
		38: migration.NewMigration39(),
		37: migration.NewMigration38(),
		36: migration.NewMigration37(),
		35: migration.NewMigration36(),
	}
)

func (svc *ServiceImpl) ExecuteMigration(ctx context.Context, installedVersion float64, codeVersion float64) error {

	svc.Logger.Printf("Migrating from %v to %v", installedVersion, codeVersion)

	sysConn, err := svc.Repo.GetSystemConnection(ctx)
	if err != nil {
		return eris.Wrap(err, "failed to get system connection")
	}
	defer sysConn.Close()

	installedMajorVersion := math.Trunc(installedVersion)

	// bump version on minor version change
	if math.Trunc(installedVersion) == math.Trunc(codeVersion) {
		if _, err = sysConn.ExecContext(ctx, "UPDATE setting SET installed_version = ?", codeVersion); err != nil {
			svc.Logger.Printf("error updating setting table: %v", err)
			return err
		}
		svc.Logger.Printf("bumped version to %v", codeVersion)
		return nil
	}

	// get migrations to run
	migration, found := migrations[installedMajorVersion]

	if !found {
		return eris.Errorf("no migration found for version %v", codeVersion)
	}

	if migration.HasSystemUpdate() {
		svc.Logger.Println("running system migration")
		if err = migration.UpdateSystem(ctx, svc.Config, sysConn); err != nil {
			svc.Logger.Printf("error updating system: %v", err)
			return eris.Wrap(err, "ExecuteMigration")
		}
	}

	if migration.HasWorkspaceUpdate() {

		workspaces := []*entity.Workspace{}

		// list all workspaces
		if err = sqlscan.Select(ctx, sysConn, &workspaces, `SELECT id FROM workspace`); err != nil {
			svc.Logger.Printf("error listing workspaces: %v", err)
			return eris.Wrap(err, "ExecuteMigration")
		}

		// update each workspace
		for _, workspace := range workspaces {

			svc.Logger.Printf("updating workspace %s", workspace.ID)

			conn, err := svc.Repo.GetWorkspaceConnection(ctx, workspace.ID)

			if err != nil {
				svc.Logger.Printf("error getting workspace connection: %v", err)
				continue
			}
			defer conn.Close()

			if err = migration.UpdateWorkspace(ctx, svc.Config, workspace, conn); err != nil {
				svc.Logger.Printf("error updating workspace %s: %v", workspace.ID, err)
				return eris.Wrap(err, "ExecuteMigration")
			}

			conn.Close()
		}
	}

	toVersion := migration.GetMajorVersion()

	// bump version
	if _, err = sysConn.ExecContext(ctx, "UPDATE setting SET installed_version = ?", toVersion); err != nil {
		svc.Logger.Printf("error updating setting table: %v", err)
		return err
	}
	svc.Logger.Printf("bumped version to %v", toVersion)

	return nil
}
