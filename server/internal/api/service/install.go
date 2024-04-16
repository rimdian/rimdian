package service

import (
	"context"

	"github.com/rimdian/rimdian/internal/api/common"
	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rotisserie/eris"
)

// checks if the DB schemas have been installed, or compatible with this server version
// (DB cm_server : accounts, system)
func (svc *ServiceImpl) InstallOrVerifyServer(ctx context.Context) (success bool, err error) {

	// skip if db is under maintenance
	if svc.Config.DB_MAINTENANCE {
		return false, nil
	}

	// get settings
	settings, err := svc.Repo.GetSettings(ctx)

	// Install schemas if nothing found
	if eris.Is(err, entity.ErrSettingsTableNotFound) {

		rootAccount, err := entity.GenerateRootAccount(svc.Config)

		if err != nil {
			return false, eris.Wrap(err, "InstallOrVerifyServer")
		}

		// svc.Logger.Printf("svc.Config.ORGANIZATION_ID %v", svc.Config.ORGANIZATION_ID)
		defaultOrganization, err := entity.GenerateDefaultOrganization(svc.Config.ORGANIZATION_ID, svc.Config.ORGANIZATION_NAME)

		if err != nil {
			return false, eris.Wrap(err, "InstallOrVerifyServer")
		}

		if err := svc.Repo.Install(ctx, rootAccount, defaultOrganization); err != nil {
			return false, eris.Wrap(err, "InstallOrVerifyServer")
		}

		// install global queues
		if err = svc.TaskOrchestrator.EnsureQueue(ctx, svc.Config.TASK_QUEUE_LOCATION, entity.TasksQueueName, 1000); err != nil {
			return false, eris.Wrap(err, "InstallOrVerifyServer")
		}

		if err = svc.TaskOrchestrator.EnsureQueue(ctx, svc.Config.TASK_QUEUE_LOCATION, entity.ScheduledTasksQueueName, 1000); err != nil {
			return false, eris.Wrap(err, "InstallOrVerifyServer")
		}

		// refresh settings
		settings, err = svc.Repo.GetSettings(ctx)

		if err != nil {
			return false, eris.Wrap(err, "InstallOrVerifyServer")
		}
	} else if err != nil {
		return false, eris.Wrap(err, "InstallOrVerifyServer")
	}

	// installed version should match major version
	if settings.InstalledVersion == common.Version {
		return true, nil
	}

	if settings.InstalledVersion > common.Version {
		return false, eris.Errorf("Installed version %v is greater than server version %v, downgrading is not possible", settings.InstalledVersion, common.Version)
	}

	if err := svc.ExecuteMigration(ctx, settings.InstalledVersion, common.Version); err != nil {
		return false, eris.Wrap(err, "InstallOrVerifyServer")
	}

	return svc.InstallOrVerifyServer(ctx)
}
