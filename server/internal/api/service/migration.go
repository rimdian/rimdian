package service

import (
	"context"
	"log"
	"math"

	"github.com/georgysavva/scany/v2/sqlscan"
	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rimdian/rimdian/internal/api/migration"
	"github.com/rotisserie/eris"
)

var (
	migrations = map[float64]entity.MajorMigrationInterface{
		// add new migrations here
		36.0: migration.NewMigration37(),
		35.0: migration.NewMigration36(),
	}
)

func (svc *ServiceImpl) ExecuteMigration(ctx context.Context, installedVersion float64, codeVersion float64) error {

	log.Printf("Migrating from %v to %v", installedVersion, codeVersion)

	sysConn, err := svc.Repo.GetSystemConnection(ctx)
	if err != nil {
		return eris.Wrap(err, "failed to get system connection")
	}
	defer sysConn.Close()

	// bump version on minor version change
	if math.Trunc(installedVersion) == math.Trunc(codeVersion) {
		if _, err = sysConn.ExecContext(ctx, "UPDATE setting SET installed_version = ?", codeVersion); err != nil {
			log.Printf("error updating setting table: %v", err)
			return err
		}
		log.Printf("bumped version to %v", codeVersion)
		return nil
	}

	// get migrations to run
	migration, found := migrations[installedVersion]

	if !found {
		return eris.Errorf("no migration found for version %v", codeVersion)
	}

	if migration.HasSystemUpdate() {
		log.Println("running system migration")
		if err = migration.UpdateSystem(ctx, svc.Config, sysConn); err != nil {
			log.Printf("error updating system: %v", err)
			return eris.Wrap(err, "ExecuteMigration")
		}
	}

	if migration.HasWorkspaceUpdate() {

		workspaces := []*entity.Workspace{}

		// list all workspaces
		if err = sqlscan.Select(ctx, sysConn, &workspaces, `SELECT id FROM workspace`); err != nil {
			log.Printf("error listing workspaces: %v", err)
			return eris.Wrap(err, "ExecuteMigration")
		}

		// update each workspace
		for _, workspace := range workspaces {

			log.Printf("updating workspace %s", workspace.ID)

			conn, err := svc.Repo.GetWorkspaceConnection(ctx, workspace.ID)

			if err != nil {
				log.Printf("error getting workspace connection: %v", err)
				continue
			}
			defer conn.Close()

			if err = migration.UpdateWorkspace(ctx, svc.Config, workspace, conn); err != nil {
				log.Printf("error updating workspace %s: %v", workspace.ID, err)
				return eris.Wrap(err, "ExecuteMigration")
			}

			conn.Close()
		}

		toVersion := migration.GetMajorVersion()

		// bump version
		if _, err = sysConn.ExecContext(ctx, "UPDATE setting SET installed_version = ?", toVersion); err != nil {
			log.Printf("error updating setting table: %v", err)
			return err
		}
		log.Printf("bumped version to %v", toVersion)

		return nil
	}

	return nil
}
