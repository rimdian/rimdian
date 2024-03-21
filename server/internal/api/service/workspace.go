package service

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/georgysavva/scany/v2/sqlscan"
	"github.com/rimdian/rimdian/internal/api/common"
	"github.com/rimdian/rimdian/internal/api/dto"
	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rotisserie/eris"
)

func (svc *ServiceImpl) WorkspaceCreateOrResetDemo(ctx context.Context, accountID string, workspaceDemoDTO *dto.WorkspaceCreateOrResetDemo) (workspace *entity.Workspace, code int, err error) {

	// verify that token is owner of its organization
	isOwner, code, err := svc.IsOwnerOfOrganization(ctx, accountID, workspaceDemoDTO.OrganizationID)

	if err != nil {
		return nil, code, eris.Wrap(err, "WorkspaceCreateOrResetDemo")
	}

	if !isOwner {
		return nil, 401, eris.Wrap(entity.ErrAccountIsNotOwner, "WorkspaceCreateOrResetDemo")
	}

	if workspaceDemoDTO == nil {
		return nil, 400, eris.New("workspace create demo payload is missing")
	}

	if workspaceDemoDTO.Kind != entity.WorkspaceDemoOrder {
		return nil, 400, eris.Errorf("workspace demo %v is not implemented", workspaceDemoDTO.Kind)
	}

	// append organization ID to the workspace ID
	workspaceID := fmt.Sprintf("%v_%v", workspaceDemoDTO.OrganizationID, "demoecommerce")

	// fetch eventual existing demo
	workspace, err = svc.Repo.GetWorkspace(ctx, workspaceID)

	if err != nil && !sqlscan.NotFound(err) {
		return nil, 500, eris.Wrap(err, "WorkspaceCreateOrResetDemo")
	}

	// delete previous demo
	if workspace != nil {
		if err := svc.Repo.DeleteWorkspace(ctx, workspace.ID); err != nil {
			return nil, 500, eris.Wrap(err, "WorkspaceCreateOrResetDemo")
		}
	}

	// create/reset demo workspace
	workspace, err = entity.GenerateDemoWorkspace(workspaceID, workspaceDemoDTO.Kind, workspaceDemoDTO.OrganizationID, svc.Config.SECRET_KEY)

	if err != nil {
		return nil, 500, eris.Wrap(err, "WorkspaceCreateOrResetDemo")
	}

	code, err = svc.Repo.RunInTransactionForSystem(ctx, func(ctx context.Context, tx *sql.Tx) (code int, err error) {

		if code, err = svc.doCreateWorkspaceInDB(ctx, workspace, tx); err != nil {
			return code, eris.Wrap(err, "WorkspaceCreateOrResetDemo")
		}

		// launch demo fixtures task
		if workspace.IsDemo {
			state := entity.NewTaskState()
			state.Workers[0]["status"] = entity.DemoTaskStatusInit
			state.Workers[0]["processed_data_logs"] = 0

			taskExec := &entity.TaskExec{
				TaskID:         entity.TaskKindGenerateDemo,
				Name:           entity.TaskNameGenerateDemo,
				OnMultipleExec: entity.OnMultipleExecAllow,
				State:          state,
			}

			if code, err := svc.doTaskCreate(ctx, workspace.ID, taskExec); err != nil {
				return code, err
			}
		}

		return 200, nil
	})

	if err != nil {
		return nil, code, eris.Wrap(err, "WorkspaceCreateOrResetDemo")
	}

	workspace.AttachMetadatas(ctx, svc.Config)

	return workspace, 201, nil
}

func (svc *ServiceImpl) WorkspaceCreate(ctx context.Context, accountID string, workspaceDTO *dto.WorkspaceCreate) (workspace *entity.Workspace, code int, err error) {

	// verify that token is owner of its organization
	isOwner, code, err := svc.IsOwnerOfOrganization(ctx, accountID, workspaceDTO.OrganizationID)

	if err != nil {
		return nil, code, eris.Wrap(err, "WorkspaceCreate")
	}

	if !isOwner {
		return nil, 401, eris.Wrap(entity.ErrAccountIsNotOwner, "WorkspaceCreate")
	}

	// convert DTO to entity
	workspace = &entity.Workspace{
		ID:                     workspaceDTO.ID,
		Name:                   workspaceDTO.Name,
		IsDemo:                 false,
		WebsiteURL:             workspaceDTO.WebsiteURL,
		PrivacyPolicyURL:       workspaceDTO.PrivacyPolicyURL,
		Industry:               workspaceDTO.Industry,
		Currency:               workspaceDTO.Currency,
		OrganizationID:         workspaceDTO.OrganizationID,
		DefaultUserTimezone:    workspaceDTO.DefaultUserTimezone,
		DefaultUserCountry:     workspaceDTO.DefaultUserCountry,
		DefaultUserLanguage:    workspaceDTO.DefaultUserLanguage,
		UserReconciliationKeys: entity.DefaultUserReconciliationKeys,
		UserIDSigning:          entity.UserIDSigningNone,
		SessionTimeout:         60 * 30, // in secs = 30 mins
		ChannelGroups:          entity.ChannelGroups{},
		Channels:               entity.Channels{},
		Domains:                entity.Domains{},
		HasOrders:              true, // hardcode order for now
		// HasLeads:               workspaceDTO.HasLeads,
		// LeadStages:             workspaceDTO.LeadStages,
		InstalledApps: entity.InstalledApps{},
		DataHooks:     entity.DataHooks{},

		FilesSettings: entity.FilesSettings{},
	}

	// append organization ID to the workspace ID
	workspace.ID = fmt.Sprintf("%v_%v", workspaceDTO.OrganizationID, workspace.ID)

	// add default channels & groups
	now := time.Now().UTC()

	for _, gr := range entity.DefaultChannelGroups {
		gr.CreatedAt = now
		gr.UpdatedAt = now
		workspace.ChannelGroups = append(workspace.ChannelGroups, gr)
	}

	for _, ch := range entity.DefaultChannels {
		ch.CreatedAt = now
		ch.UpdatedAt = now
		if err := ch.Validate(workspace.Channels, workspace.ChannelGroups); err != nil {
			return nil, 500, err
		}
		workspace.Channels = append(workspace.Channels, ch)
	}

	// Generate a secret key
	key := common.RandomString(32)

	// hardcode testing secret key
	if workspace.ID == "testing" {
		key = entity.WorkspaceTestingSecretKey
	}

	encryptedKey, err := common.EncryptString(key, svc.Config.SECRET_KEY)

	if err != nil {
		return nil, 500, eris.Wrap(err, "WorkspaceCreate")
	}

	secretKey := &entity.SecretKey{
		Key:          key,
		EncryptedKey: encryptedKey,
		CreatedAt:    time.Now(),
	}

	workspace.SecretKeys = []*entity.SecretKey{secretKey}

	if err := workspace.Validate(); err != nil {
		return nil, 400, eris.Wrap(err, "WorkspaceCreate")
	}

	code, err = svc.Repo.RunInTransactionForSystem(ctx, func(ctx context.Context, tx *sql.Tx) (code int, err error) {

		if code, err = svc.doCreateWorkspaceInDB(ctx, workspace, tx); err != nil {
			return code, eris.Wrap(err, "WorkspaceCreate")
		}

		return 200, nil
	})

	if err != nil {
		return nil, code, err
	}

	workspace.AttachMetadatas(ctx, svc.Config)

	return workspace, 201, nil
}

// list workspace for an organization
func (svc *ServiceImpl) WorkspaceList(ctx context.Context, accountID string, organizationID string) (result *dto.WorkspaceListResult, code int, err error) {

	// verify that token is account of its organization
	isAccount, code, err := svc.IsAccountOfOrganization(ctx, accountID, organizationID)

	if err != nil {
		return nil, code, eris.Wrap(err, "WorkspaceList")
	}

	if !isAccount {
		return nil, 400, eris.New("account is not part of the organization")
	}

	workspaces, err := svc.Repo.ListWorkspaces(ctx, &organizationID)

	if err != nil {
		return nil, 500, eris.Wrap(err, "WorkspaceList")
	}

	for _, w := range workspaces {
		w.AttachMetadatas(ctx, svc.Config)
	}

	result = &dto.WorkspaceListResult{Workspaces: workspaces}

	return result, 200, nil
}

// list workspace tables
func (svc *ServiceImpl) WorkspaceShowTables(ctx context.Context, accountID string, workspaceID string) (result *dto.WorkspaceShowTablesResult, code int, err error) {

	organizationID := entity.ExtractOrganizationIDFromWorkspaceID(workspaceID)

	// verify that token is account of its organization
	isAccount, code, err := svc.IsAccountOfOrganization(ctx, accountID, organizationID)

	if err != nil {
		return nil, code, eris.Wrap(err, "WorkspaceShowTables")
	}

	if !isAccount {
		return nil, 400, eris.New("account is not part of the organization")
	}

	workspace, err := svc.Repo.GetWorkspace(ctx, workspaceID)

	if err != nil {
		if sqlscan.NotFound(err) {
			return nil, 400, err
		}
		return nil, 500, eris.Wrap(err, "WorkspaceShow")
	}

	if workspace.OrganizationID != organizationID {
		return nil, 400, eris.New("workspace does not belong to organization")
	}

	tables, err := svc.Repo.ShowTables(ctx, workspaceID)

	if err != nil {
		return nil, 500, eris.Wrap(err, "WorkspaceShowTables")
	}

	// svc.Logger.Printf("formated %+v\n", tables)

	result = &dto.WorkspaceShowTablesResult{
		Tables: tables,
	}

	return result, 200, nil
}

// get workspace
func (svc *ServiceImpl) WorkspaceShow(ctx context.Context, accountID string, workspaceID string) (result *dto.WorkspaceShowResult, code int, err error) {

	organizationID := entity.ExtractOrganizationIDFromWorkspaceID(workspaceID)

	// verify that token is account of its organization
	isAccount, code, err := svc.IsAccountOfOrganization(ctx, accountID, organizationID)

	if err != nil {
		return nil, code, eris.Wrap(err, "WorkspaceShow")
	}

	if !isAccount {
		return nil, 400, eris.New("account is not part of the organization")
	}

	workspace, err := svc.Repo.GetWorkspace(ctx, workspaceID)

	if err != nil {
		if sqlscan.NotFound(err) {
			return nil, 400, err
		}
		return nil, 500, eris.Wrap(err, "WorkspaceShow")
	}

	if workspace.OrganizationID != organizationID {
		return nil, 400, eris.New("workspace does not belong to organization")
	}

	result = &dto.WorkspaceShowResult{Workspace: workspace}

	result.Workspace.AttachMetadatas(ctx, svc.Config)

	return result, 200, nil
}

func (svc *ServiceImpl) WorkspaceUpdate(ctx context.Context, accountID string, payload *dto.WorkspaceUpdate) (workspace *entity.Workspace, code int, err error) {

	if payload == nil || payload.WorkspaceCreate == nil {
		return nil, 400, eris.New("workspace update payload is missing")
	}

	// fetch workspace
	workspace, err = svc.Repo.GetWorkspace(ctx, payload.ID)

	if err != nil {
		if sqlscan.NotFound(err) {
			return nil, 400, err
		}
		return nil, 500, eris.Wrap(err, "WorkspaceUpdate")
	}

	// verify that token is owner of its organization
	isAccount, code, err := svc.IsAccountOfOrganization(ctx, accountID, workspace.OrganizationID)

	if err != nil {
		return nil, code, eris.Wrap(err, "WorkspaceUpdate")
	}

	if !isAccount {
		return nil, 400, eris.New("account is not part of the organization")
	}

	workspace.Name = payload.Name
	workspace.WebsiteURL = payload.WebsiteURL
	workspace.PrivacyPolicyURL = payload.PrivacyPolicyURL
	workspace.Industry = payload.Industry
	workspace.Currency = payload.Currency
	workspace.DefaultUserTimezone = payload.DefaultUserTimezone
	workspace.DefaultUserCountry = payload.DefaultUserCountry
	workspace.DefaultUserLanguage = payload.DefaultUserLanguage
	workspace.UserReconciliationKeys = payload.UserReconciliationKeys
	workspace.UserIDSigning = payload.UserIDSigning
	workspace.SessionTimeout = payload.SessionTimeout
	// workspace.HasOrders = payload.HasOrders
	// workspace.HasLeads = payload.HasLeads
	workspace.LeadStages = []*entity.LeadStage{}

	deletedLeadStages := []*entity.LeadStage{}

	if payload.HasLeads {

		workspace.LeadStages = []*entity.LeadStage{}

		if payload.LeadStages != nil {
			for _, s := range payload.LeadStages {
				stage := &entity.LeadStage{
					ID:     s.ID,
					Label:  s.Label,
					Status: s.Status,
					Color:  s.Color,
				}
				if s.DeletedAt != nil && s.MigrateToID != nil {
					stage.DeletedAt = s.DeletedAt
					stage.MigrateToID = s.MigrateToID
					deletedLeadStages = append(deletedLeadStages, stage)
				}
				workspace.LeadStages = append(workspace.LeadStages, stage)
			}
		}
	}

	// check S3 settings if they are set
	if payload.FilesSettings.Endpoint != "" {

		if err := payload.FilesSettings.Validate(); err != nil {
			return nil, 400, eris.Wrap(err, "WorkspaceUpdate")
		}

		// update workspace settings
		workspace.FilesSettings.Endpoint = payload.FilesSettings.Endpoint
		workspace.FilesSettings.AccessKey = payload.FilesSettings.AccessKey
		workspace.FilesSettings.Bucket = payload.FilesSettings.Bucket
		workspace.FilesSettings.Region = payload.FilesSettings.Region
		workspace.FilesSettings.CDNEndpoint = payload.FilesSettings.CDNEndpoint

		// encrypt secret key
		if payload.FilesSettings.SecretKey != "" {
			if workspace.FilesSettings.EncryptedSecretKey, err = common.EncryptString(payload.FilesSettings.SecretKey, svc.Config.SECRET_KEY); err != nil {
				return nil, 500, eris.Wrap(err, "WorkspaceUpdate")
			}
		}
	}

	if err := workspace.Validate(); err != nil {
		return nil, 400, eris.Wrap(err, "WorkspaceUpdate")
	}

	// verify license
	if payload.LicenseKey != nil && *payload.LicenseKey != "" {
		body := fmt.Sprintf(`{
			"license": "%v",
			"api_endpoint": "%v",
			"workspace_id": "%v"
		}`, *payload.LicenseKey, svc.Config.API_ENDPOINT, workspace.ID)

		req, _ := http.NewRequest("POST", "https://store.rimdian.com/verifyLicense", bytes.NewBuffer([]byte(body)))
		req.Header.Set("Content-Type", "application/json")

		// inject OpenCensus span context into the request
		req = req.WithContext(ctx)

		resp, err := svc.NetClient.Do(req)

		if err != nil {
			return nil, 500, eris.Wrap(err, "WorkspaceUpdate")
		}

		defer resp.Body.Close()

		data, _ := io.ReadAll(resp.Body)

		if resp.StatusCode != 200 {
			return nil, 400, eris.Errorf("license verification failed: %v", string(data))
		}

		workspace.LicenseKey = payload.LicenseKey
	} else {
		workspace.LicenseKey = nil
	}

	// insert workspace in repo
	if err := svc.Repo.UpdateWorkspace(ctx, workspace, nil); err != nil {
		return nil, 500, eris.Wrap(err, "WorkspaceUpdate")
	}

	// TODO: migrate deleted lead stages
	if len(deletedLeadStages) > 0 {
		svc.Logger.Printf("migrating deleted lead stages not implented %+v\n", deletedLeadStages)
	}

	// for _, stage := range leadStages {
	// 	if stage.MigrateToID == nil {
	// 		continue
	// 	}
	// 	query, args, err := sq.Update("leads").
	// 		Set("stage_id", *stage.MigrateToID).
	// 		Where(sq.Eq{"stage_id": stage.ID}).
	// 		ToSql()

	// 	if err != nil {
	// 		return eris.Wrapf(err, "MigrateDeletedLeadStages build query for stage %+v\n", stage)
	// 	}

	// 	_, err = tx.ExecContext(ctx, query, args...)

	// 	if err != nil {
	// 		return eris.Wrapf(err, "MigrateDeletedLeadStages exec query %v", query)
	// 	}
	// }

	workspace.AttachMetadatas(ctx, svc.Config)

	return workspace, 200, nil
}

func (svc *ServiceImpl) WorkspaceGetSecretKey(ctx context.Context, accountID string, workspaceID string) (result *dto.WorkspaceSecretKeyResult, code int, err error) {

	// fetch workspace
	workspace, err := svc.Repo.GetWorkspace(ctx, workspaceID)

	if err != nil {
		if sqlscan.NotFound(err) {
			return nil, 400, err
		}
		return nil, 500, eris.Wrap(err, "WorkspaceGetSecretKey")
	}

	// verify that token is owner of its organization
	isAccount, code, err := svc.IsAccountOfOrganization(ctx, accountID, workspace.OrganizationID)

	if err != nil {
		return nil, code, eris.Wrap(err, "WorkspaceGetSecretKey")
	}

	if !isAccount {
		return nil, 400, eris.New("account is not part of the organization")
	}

	if err := workspace.DecryptSecretKeys(svc.Config.SECRET_KEY); err != nil {
		return nil, 500, eris.Wrap(err, "WorkspaceGetSecretKey")
	}

	activeKey := ""
	for _, x := range workspace.SecretKeys {
		if x.DeletedAt == nil {
			activeKey = x.Key
		}
	}

	result = &dto.WorkspaceSecretKeyResult{
		SecretKey: activeKey,
	}

	return result, 200, nil
}

func (svc *ServiceImpl) doCreateWorkspaceInDB(ctx context.Context, workspace *entity.Workspace, tx *sql.Tx) (code int, err error) {

	// ensure Google Tasks Queues exist
	if svc.Config.ENV != entity.ENV_DEV {
		if err := svc.EnsureGoogleTasksQueues(ctx, workspace.ID); err != nil {
			return 500, eris.Wrap(err, "doCreateWorkspaceInDB")
		}
	}

	// insert workspace
	if err := svc.Repo.InsertWorkspace(ctx, workspace, tx); err != nil {
		if eris.Is(err, entity.ErrWorkspaceAlreadyExists) {
			return 400, eris.Wrap(entity.ErrWorkspaceAlreadyExists, "doCreateWorkspaceInDB")
		}
		return 500, err
	}

	if err := svc.Repo.UseWorkspaceDBWithTx(ctx, workspace.ID, tx); err != nil {
		return 500, eris.Wrap(err, "doCreateWorkspaceInDB")
	}

	// create workspace tables
	if err := svc.Repo.CreateWorkspaceTables(ctx, workspace.ID, tx); err != nil {
		return 500, eris.Wrap(err, "doCreateWorkspaceInDB")
	}

	// insert default segments
	for _, seg := range entity.GenerateDefaultSegments() {
		if err := svc.Repo.InsertSegment(ctx, &seg, tx); err != nil {
			return 500, eris.Wrap(err, "doCreateWorkspaceInDB")
		}
	}

	return 200, nil
}

// ensure Google Tasks Queues historical & live exist
func (svc *ServiceImpl) EnsureGoogleTasksQueues(ctx context.Context, workspaceID string) (err error) {
	queueName := svc.TaskOrchestrator.GetHistoricalQueueNameForWorkspace(workspaceID)
	if err = svc.TaskOrchestrator.EnsureQueue(ctx, svc.Config.TASK_QUEUE_LOCATION, queueName); err != nil {
		return err
	}
	queueName = svc.TaskOrchestrator.GetLiveQueueNameForWorkspace(workspaceID)
	return svc.TaskOrchestrator.EnsureQueue(ctx, svc.Config.TASK_QUEUE_LOCATION, queueName)
}
