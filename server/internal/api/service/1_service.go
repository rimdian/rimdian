//go:generate moq -out service_moq.go . Service
package service

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/georgysavva/scany/v2/sqlscan"
	"github.com/rimdian/rimdian/internal/api/dto"
	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rimdian/rimdian/internal/api/repository"
	common "github.com/rimdian/rimdian/internal/common/dto"
	"github.com/rimdian/rimdian/internal/common/httpClient"
	"github.com/rimdian/rimdian/internal/common/mailer"
	"github.com/rimdian/rimdian/internal/common/taskorchestrator"
	"github.com/rotisserie/eris"
	"github.com/sirupsen/logrus"
)

var (
	ErrServicePayloadRequired = eris.New("a payload is required to execute this function")
)

type Service interface {
	GetConfig() *entity.Config
	// dev
	DevResetDB(ctx context.Context) (err error)
	DevExecTaskWithWorkers(ctx context.Context, workspaceID string) (code int, err error)
	DevAddDataImportToQueue(dataLogInQueue *common.DataLogInQueue)
	DevExecDataImportFromQueue(ctx context.Context, concurrency int) (code int, err error)

	// system
	InstallOrVerifyServer(ctx context.Context) (success bool, err error)
	ExecuteMigration(ctx context.Context, installedVersion float64, codeVersion float64) error
	SendSystemEmail(ctx context.Context, systemEmail *dto.SystemEmail) error
	// account
	AccountLogin(ctx context.Context, loginDTO *dto.AccountLogin) (loginResult *dto.AccountLoginResult, code int, err error)
	AccountLogout(ctx context.Context, accountID string, sessionID string) (code int, err error)
	AccountRefreshAccessToken(ctx context.Context, accountID string, accountSessionID string) (loginResult *dto.AccountRefreshAccessTokenResult, code int, err error)
	AccountResetPassword(ctx context.Context, resetPasswordDTO *dto.AccountResetPassword) (code int, err error)
	AccountConsumeResetPassword(ctx context.Context, resetPasswordDTO *dto.AccountConsumeResetPassword) (loginResult *dto.AccountLoginResult, code int, err error)
	AccountSetProfile(ctx context.Context, accountID string, accountProfileDTO *dto.AccountProfile) (updatedAccount *entity.Account, code int, err error)
	IsAccountOfOrganization(ctx context.Context, accountId string, organizationId string) (isAccount bool, code int, err error)
	IsOwnerOfOrganization(ctx context.Context, accountId string, organizationId string) (isOwner bool, code int, err error)
	// organization
	OrganizationSetProfile(ctx context.Context, accountID string, orgProfileDTO *dto.OrganizationProfile) (updatedOrg *dto.OrganizationResult, code int, err error)
	OrganizationList(ctx context.Context, accountID string) (result *dto.OrganizationListResult, code int, err error)
	OrganizationCreate(ctx context.Context, orgCreateDTO *dto.OrganizationCreate) (org *dto.OrganizationResult, code int, err error)
	// organization invitation
	OrganizationInvitationCreate(ctx context.Context, accountInvitationDTO *dto.OrganizationInvitation) (code int, err error)
	// OrganizationInvitationResend(ctx context.Context, accountID string, data *dto.OrganizationInvitationResend) (code int, err error)
	OrganizationInvitationConsume(ctx context.Context, consumeInvitationDTO *dto.OrganizationInvitationConsume) (loginResult *dto.AccountLoginResult, code int, err error)
	OrganizationInvitationCancel(ctx context.Context, accountID string, cancelInvitation *dto.OrganizationInvitationCancel) (code int, err error)
	OrganizationInvitationList(ctx context.Context, accountID string, organizationID string) (result *dto.OrganizationInvitationListResult, code int, err error)
	OrganizationInvitationRead(ctx context.Context, token string) (result *dto.OrganizationInvitationReadResult, code int, err error)
	// organization account
	OrganizationAccountList(ctx context.Context, accountID string, organizationID string) (result *dto.OrganizationAccountListResult, code int, err error)
	OrganizationAccountCreateServiceAccount(ctx context.Context, accountID string, createServiceAccount *dto.OrganizationAccountCreateServiceAccount) (code int, err error)
	OrganizationAccountTransferOwnership(ctx context.Context, accountID string, transferOwnershipDTO *dto.OrganizationAccountTransferOwnership) (code int, err error)
	OrganizationAccountDeactivate(ctx context.Context, accountID string, deactivateAccountDTO *dto.OrganizationAccountDeactivate) (code int, err error)
	// workspace
	GetWorkspaceForAccount(ctx context.Context, workspaceID string, accountID string) (workspace *entity.Workspace, code int, err error)
	WorkspaceList(ctx context.Context, accountID string, organizationID string) (result *dto.WorkspaceListResult, code int, err error)
	WorkspaceShow(ctx context.Context, accountID string, workspaceID string) (result *dto.WorkspaceShowResult, code int, err error)
	WorkspaceShowTables(ctx context.Context, accountID string, workspaceID string) (result *dto.WorkspaceShowTablesResult, code int, err error)
	WorkspaceCreate(ctx context.Context, accountID string, workspaceDTO *dto.WorkspaceCreate) (workspace *entity.Workspace, code int, err error)
	WorkspaceCreateOrResetDemo(ctx context.Context, accountID string, workspaceDemoDTO *dto.WorkspaceCreateOrResetDemo) (workspace *entity.Workspace, code int, err error)
	WorkspaceUpdate(ctx context.Context, accountID string, payload *dto.WorkspaceUpdate) (workspace *entity.Workspace, code int, err error)
	WorkspaceSettingsUpdate(ctx context.Context, accountID string, payload *dto.WorkspaceSettingsUpdate) (updatedWorkspace *entity.Workspace, code int, err error)
	WorkspaceGetSecretKey(ctx context.Context, accountID string, workspaceID string) (result *dto.WorkspaceSecretKeyResult, code int, err error)
	// domain
	DomainUpsert(ctx context.Context, accountID string, domainDTO *dto.Domain) (updatedWorkspace *entity.Workspace, code int, err error)
	DomainDelete(ctx context.Context, accountID string, domainDeleteDTO *dto.DomainDelete) (updatedWorkspace *entity.Workspace, code int, err error)
	// channel group
	ChannelGroupUpsert(ctx context.Context, accountID string, channelGroupDTO *dto.ChannelGroup) (updatedWorkspace *entity.Workspace, code int, err error)
	ChannelGroupDelete(ctx context.Context, accountID string, deleteChannelGroupDTO *dto.DeleteChannelGroup) (updatedWorkspace *entity.Workspace, code int, err error)
	// channel
	ChannelCreate(ctx context.Context, accountID string, channelDTO *dto.Channel) (updatedWorkspace *entity.Workspace, code int, err error)
	ChannelUpdate(ctx context.Context, accountID string, channelDTO *dto.Channel) (updatedWorkspace *entity.Workspace, code int, err error)
	ChannelDelete(ctx context.Context, accountID string, deleteChannelDTO *dto.DeleteChannel) (updatedWorkspace *entity.Workspace, code int, err error)
	// task
	TaskList(ctx context.Context, accountID string, params *dto.TaskListParams) (result *dto.TaskListResult, code int, err error)
	TaskRun(ctx context.Context, accountID string, params *dto.TaskRunParams) (code int, err error)
	TaskWakeUpCron(ctx context.Context) (code int, err error)
	// task_exec
	TaskExecCreate(ctx context.Context, accountID string, params *dto.TaskExecCreateParams) (code int, err error)
	TaskExecDo(ctx context.Context, workspaceID string, payload *dto.TaskExecRequestPayload) (result *common.DataLogInQueueResult)
	TaskExecAbort(ctx context.Context, accountID string, params *dto.TaskExecAbortParams) (code int, err error)
	TaskExecList(ctx context.Context, accountID string, params *dto.TaskExecListParams) (result *dto.TaskExecListResult, code int, err error)

	// task_exec_job
	TaskExecJobInfo(ctx context.Context, accountID string, params *dto.TaskExecJobInfoParams) (runningJobInfo *dto.TaskExecJobInfoInfo, code int, err error)
	TaskExecJobs(ctx context.Context, accountID string, params *dto.TaskExecJobsParams) (result *dto.TaskExecJobsResult, code int, err error)
	// data log
	DataLogImportFromQueue(ctx context.Context, dataLogInQueue *common.DataLogInQueue) (result *common.DataLogInQueueResult)
	DataLogList(ctx context.Context, accountID string, params *dto.DataLogListParams) (result *dto.DataLogListResult, code int, err error)
	DataLogReprocessOne(ctx context.Context, accountID string, params *dto.DataLogReprocessOne) (result *common.DataLogInQueueResult, code int, err error)
	DataLogReprocessUntil(ctx context.Context, untilDate time.Time) (code int, err error)

	// DB Select
	DBSelect(ctx context.Context, accountID string, params *dto.DBSelectParams) (rows []map[string]interface{}, code int, err error)
	DoDBSelect(workspaceID string, query string, args []interface{}) (jsonData []byte, err error)

	// user
	UserList(ctx context.Context, accountID string, params *dto.UserListParams) (result *dto.UserListResult, code int, err error)
	UserShow(ctx context.Context, workspaceID string, accountID string, userExternalID string) (result *dto.UserShowResult, code int, err error)

	// segment
	SegmentList(ctx context.Context, accountID string, params *dto.SegmentListParams) (result *dto.SegmentListResult, code int, err error)
	SegmentPreview(ctx context.Context, accountID string, params *dto.SegmentPreviewParams) (result *dto.SegmentPreviewResult, code int, err error)
	SegmentCreate(ctx context.Context, accountID string, segmentDTO *dto.Segment) (code int, err error)
	SegmentUpdate(ctx context.Context, accountID string, segmentDTO *dto.Segment) (code int, err error)
	SegmentDelete(ctx context.Context, accountID string, deleteSegmentDTO *dto.DeleteSegment) (code int, err error)

	// subscription list
	SubscriptionListList(ctx context.Context, accountID string, params *dto.SubscriptionListListParams) (result []*entity.SubscriptionList, code int, err error)
	SubscriptionListCreate(ctx context.Context, accountID string, data *dto.SubscriptionListCreate) (code int, err error)

	// email template
	MessageTemplateList(ctx context.Context, accountID string, params *dto.MessageTemplateListParams) (result []*entity.MessageTemplate, code int, err error)
	MessageTemplateUpsert(ctx context.Context, accountID string, data *dto.MessageTemplate) (code int, err error)

	// broadcast campaign
	BroadcastCampaignList(ctx context.Context, accountID string, params *dto.BroadcastCampaignListParams) (broadcasts []*entity.BroadcastCampaign, code int, err error)
	BroadcastCampaignUpsert(ctx context.Context, accountID string, data *dto.BroadcastCampaign) (code int, err error)

	// message
	MessageSend(ctx context.Context, data *dto.SendMessage) (result *common.DataLogInQueueResult)
	SendEmailWithSparkpost(ctx context.Context, data *dto.SendMessage) (result *common.DataLogInQueueResult)
	SendEmailWithSMTP(ctx context.Context, data *dto.SendMessage) (result *common.DataLogInQueueResult)

	// data hook
	DataHookUpdate(ctx context.Context, accountID string, dataHookDTO *dto.DataHook) (updatedWorkspace *entity.Workspace, code int, err error)

	// app
	AppList(ctx context.Context, accountID string, params *dto.AppListParams) (result *dto.AppListResult, code int, err error)
	AppGet(ctx context.Context, accountID string, params *dto.AppGetParams) (app *entity.App, code int, err error)
	AppInstall(ctx context.Context, accountID string, params *dto.AppInstall) (installedApp *entity.App, code int, err error)
	AppActivate(ctx context.Context, accountID string, params *dto.AppActivate) (code int, err error)
	AppStop(ctx context.Context, accountID string, params *dto.AppDelete) (app *entity.App, code int, err error)
	AppDelete(ctx context.Context, accountID string, params *dto.AppDelete) (code int, err error)
	AppMutateState(ctx context.Context, accountID string, params *dto.AppMutateState) (code int, err error)
	AppFromToken(ctx context.Context, params *dto.AppFromTokenParams) (result *dto.AppFromToken, code int, err error)
	AppExecQuery(ctx context.Context, accountID string, params *dto.AppExecQuery) (result *dto.AppExecQueryResult, code int, err error)

	// cubejs
	CubeJSSchemas(ctx context.Context, accountID string, workspaceID string) (schemas dto.CubeJSSchemas, code int, err error)
}

func (svc *ServiceImpl) GetWorkspaceForAccount(ctx context.Context, workspaceID string, accountID string) (workspace *entity.Workspace, code int, err error) {

	// fetch workspace
	workspace, err = svc.Repo.GetWorkspace(ctx, workspaceID)

	if err != nil {
		if sqlscan.NotFound(err) {
			return nil, 400, fmt.Errorf("workspace %v not found", workspaceID)
		}
		return nil, 500, eris.Wrap(err, "GetWorkspaceForAccount")
	}

	// verify that token is owner of its organization
	isAccount, code, err := svc.IsAccountOfOrganization(ctx, accountID, workspace.OrganizationID)

	if err != nil {
		return nil, code, eris.Wrap(err, "GetWorkspaceForAccount")
	}

	if !isAccount {
		return nil, 400, eris.New("account is not part of the organization")
	}

	return workspace, 200, nil
}

type ServiceImpl struct {
	Config             *entity.Config
	Logger             *logrus.Logger
	Repo               repository.Repository
	Mailer             mailer.Mailer
	TaskOrchestrator   taskorchestrator.Client
	NetClient          httpClient.HTTPClient
	StorageClient      *storage.Client
	DevDataImportQueue *entity.DevDataImportQueue
}

func (svc *ServiceImpl) GetConfig() *entity.Config {
	return svc.Config
}

func GetNodeJSDir() (dir string, err error) {
	currentPath, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return strings.Split(currentPath, "server")[0] + "nodejs/", nil
}

func NewService(cfg *entity.Config, logger *logrus.Logger, repo repository.Repository, mailer mailer.Mailer, taskOrchestrator taskorchestrator.Client, storageClient *storage.Client, netClient httpClient.HTTPClient) Service {
	return &ServiceImpl{
		Config:             cfg,
		Logger:             logger,
		Repo:               repo,
		Mailer:             mailer,
		TaskOrchestrator:   taskOrchestrator,
		StorageClient:      storageClient,
		NetClient:          netClient,
		DevDataImportQueue: entity.NewDevDataImportQueue(), // only used in dev
	}
}
