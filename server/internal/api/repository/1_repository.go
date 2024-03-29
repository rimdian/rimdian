//go:generate moq -out repository_moq.go . Repository
package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/rimdian/rimdian/internal/api/dto"
	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rotisserie/eris"
)

var (
	ErrRowAlreadyExists      = eris.New("row already exists")
	ErrRowNotUpdated         = eris.New("row not updated")
	ErrAcquireUserLockFailed = eris.New("acquire user lock failed")
)

// records a query executed with its parameters for testing a debugging
type QueryExecuted struct {
	SQL  string
	Args []interface{}
}

type RowScanner interface {
	Scan(dest ...interface{}) error
}

type Repository interface {
	// utils
	GetSystemConnection(ctx context.Context) (*sql.Conn, error)
	GetWorkspaceConnection(ctx context.Context, workspaceID string) (*sql.Conn, error)
	RunInTransactionForSystem(ctx context.Context, f func(context.Context, *sql.Tx) (int, error)) (code int, err error)
	RunInTransactionForWorkspace(ctx context.Context, workspaceID string, f func(context.Context, *sql.Tx) (int, error)) (code int, err error)
	UseWorkspaceDBWithTx(ctx context.Context, workspaceID string, tx *sql.Tx) error
	IsDuplicateEntry(err error) bool

	// user lock
	EnsureUsersLock(ctx context.Context, workspaceID string, lock *entity.UsersLock, withRetry bool) error
	ReleaseUsersLock(workspaceID string, lock *entity.UsersLock) error

	// system
	Install(ctx context.Context, rootAccount *entity.Account, defaultOrganization *entity.Organization) error
	DevResetDB(ctx context.Context, rootAccount *entity.Account, defaultOrganization *entity.Organization) error
	GetSettings(ctx context.Context) (settings *entity.Settings, err error)

	// tasks
	ListTasks(ctx context.Context, workspaceID string) (tasks []*entity.Task, err error)
	ListTasksToWakeUp(ctx context.Context) (tasks []*entity.Task, err error)
	ActivateAppTasks(ctx context.Context, workspaceID string, appID string, tx *sql.Tx) (err error)
	GetTask(ctx context.Context, workspaceID string, taskID string, tx *sql.Tx) (task *entity.Task, err error)
	InsertTask(ctx context.Context, task *entity.Task, tx *sql.Tx) (err error)
	UpdateTask(ctx context.Context, task *entity.Task, tx *sql.Tx) (err error)
	StopAppTasks(ctx context.Context, workspaceID string, appID string, tx *sql.Tx) (err error)
	DeleteAppTasks(ctx context.Context, workspaceID string, appID string, tx *sql.Tx) (err error)

	// organizations + accounts
	CreateOrganization(ctx context.Context, organization *entity.Organization, tx *sql.Tx) error
	GetOrganization(ctx context.Context, organizationID string) (organization *entity.Organization, err error)
	InsertAccount(ctx context.Context, account *entity.Account, organizationID string, fromAccountID string, tx *sql.Tx) error
	AccountLogout(ctx context.Context, accountID string, sessionID string) error
	ResetAccountPassword(ctx context.Context, accountID string, newPassword string) error
	AddAccountToOrganization(ctx context.Context, accountID string, organizationID string, isOwner bool, fromAccountID *string, workspaceScopes entity.WorkspacesScopes, tx *sql.Tx) error
	IsAccountOfOrganization(ctx context.Context, accountId string, organizationId string, shouldBeOwner bool) (isAccount bool, err error)
	InsertAccountSession(ctx context.Context, accountSession *entity.AccountSession) error
	UpdateAccountSessionLastAccess(ctx context.Context, accountID string, accountSessionID string, now time.Time) error
	GetAccountFromEmail(ctx context.Context, email string) (account *entity.Account, err error)
	GetAccountFromID(ctx context.Context, accountID string) (account *entity.Account, err error)
	ListAccountsForOrganization(ctx context.Context, organizationID string) (accounts []*entity.AccountWithOrganizationRole, err error)
	ListOrganizationsForAccount(ctx context.Context, accountID string) (organizations []*entity.Organization, err error)
	ListInvitationsForOrganization(ctx context.Context, organizationID string) (invitations []*entity.OrganizationInvitation, err error)
	UpsertOrganizationInvitation(ctx context.Context, accountInvitation *entity.OrganizationInvitation) error
	InsertServiceAccount(ctx context.Context, account *entity.Account, organizationID string, fromAccountID string, workspaceScopes entity.WorkspacesScopes) error
	UpdateAccountProfile(ctx context.Context, account *entity.Account) error
	UpdateOrganizationProfile(ctx context.Context, organization *entity.Organization) error
	CancelOrganizationInvitation(ctx context.Context, organizationID string, email string) error
	TransferOrganizationOwnsership(ctx context.Context, accountID string, toAccountID string, organizationID string) error
	DeactivateOrganizationAccount(ctx context.Context, accountID string, deactivateAccountID string, organizationID string) error
	GetInvitation(ctx context.Context, email string, organizationID string) (invitation *entity.OrganizationInvitation, err error)
	ConsumeInvitation(ctx context.Context, accountID string, insertAccount *entity.Account, invitation *entity.OrganizationInvitation) error

	// workspaces
	GetWorkspace(ctx context.Context, workspaceID string) (workspace *entity.Workspace, err error)
	InsertWorkspace(ctx context.Context, workspace *entity.Workspace, tx *sql.Tx) (err error)
	ListWorkspaces(ctx context.Context, organizationID *string) (workspaces []*entity.Workspace, err error)
	UpdateWorkspace(ctx context.Context, workspace *entity.Workspace, tx *sql.Tx) error
	DeleteWorkspace(ctx context.Context, workspaceID string) error
	CreateWorkspaceTables(ctx context.Context, workspaceID string, tx *sql.Tx) error
	ShowTables(ctx context.Context, workspaceID string) (tables []*entity.TableInformationSchema, err error)

	// domains
	DeleteDomain(ctx context.Context, workspace *entity.Workspace, deletedDomainID string, migrateToDomainID string) error

	// channels
	CreateChannel(ctx context.Context, workspace *entity.Workspace, channel *entity.Channel) error
	UpdateChannel(ctx context.Context, workspace *entity.Workspace, updatedChannel *entity.Channel) error
	DeleteChannel(ctx context.Context, workspace *entity.Workspace, deletedChannelID string) error

	// app table
	CreateTable(ctx context.Context, workspace *entity.Workspace, table *entity.AppTableManifest) error
	DeleteTable(ctx context.Context, workspaceID string, tableName string) (err error)
	IsExistingTableTheSame(ctx context.Context, workspaceID string, table *entity.AppTableManifest) (err error)

	// extra column
	AddColumn(ctx context.Context, workspace *entity.Workspace, tableName string, column *entity.TableColumn) error
	DeleteColumn(ctx context.Context, workspace *entity.Workspace, tableName string, column *entity.TableColumn) error
	IsExistingColumnTheSame(ctx context.Context, workspaceID string, tableName string, column *entity.TableColumn) error

	// app item
	FindAppItemByID(ctx context.Context, workspace *entity.Workspace, kind string, id string, tx *sql.Tx) (item *entity.AppItem, err error)
	FindAppItemByExternalID(ctx context.Context, workspace *entity.Workspace, kind string, externalID string, tx *sql.Tx) (item *entity.AppItem, err error)
	InsertAppItem(ctx context.Context, kind string, upsertedAppItem *entity.AppItem, tx *sql.Tx) error
	UpdateAppItem(ctx context.Context, kind string, upsertedAppItem *entity.AppItem, tx *sql.Tx) error
	DeleteAppItemByID(ctx context.Context, workspace *entity.Workspace, kind string, ID string, tx *sql.Tx) error
	DeleteAppItemByExternalID(ctx context.Context, workspace *entity.Workspace, kind string, externalID string, tx *sql.Tx) error
	FetchAppItems(ctx context.Context, workspace *entity.Workspace, kind string, query sq.SelectBuilder, tx *sql.Tx) (items []*entity.AppItem, err error)

	// task_exec
	InsertTaskExec(ctx context.Context, workspaceID string, task *entity.TaskExec, job *entity.TaskExecJob, tx *sql.Tx) (err error)
	GetTaskExec(ctx context.Context, workspaceID string, taskID string) (task *entity.TaskExec, err error)
	GetRunningTaskExecByTaskID(ctx context.Context, taskID string, multipleExecKey *string, tx *sql.Tx) (task *entity.TaskExec, err error)
	SetTaskExecError(ctx context.Context, workspaceID string, taskExecID string, workerID int, status int, message string) error
	UpdateTaskExecFromResult(ctx context.Context, taskExecRequestPayload *dto.TaskExecRequestPayload, taskExecResult *entity.TaskExecResult, tx *sql.Tx) error
	AddTaskExecWorker(ctx context.Context, taskID string, newJobID string, workerID int, initialWorkerState entity.TaskWorkerState, tx *sql.Tx) error
	ListTaskExecs(ctx context.Context, workspaceID string, params *dto.TaskExecListParams) (tasks []*entity.TaskExec, nextToken string, previousToken string, code int, err error)
	StopTaskExecsForApp(ctx context.Context, appID string, tx *sql.Tx) error
	AddJobToTaskExec(ctxWithTimeout context.Context, taskExecID string, newJobID string, tx *sql.Tx) error
	AbortTaskExec(ctx context.Context, taskID string, message string, tx *sql.Tx) error
	GetTaskExecJob(ctx context.Context, workspaceID string, jobID string) (job *entity.TaskExecJob, err error)
	GetTaskExecJobs(ctx context.Context, workspaceID string, taskExecID string, offset int, limit int) (jobs []*entity.TaskExecJob, total int, err error)

	// data log
	GetDataLog(ctx context.Context, workspaceID string, dataLogID string) (dataLog *entity.DataLog, err error)
	InsertDataLog(ctx context.Context, workspaceID string, dataLog *entity.DataLog, tx *sql.Tx) (err error)
	UpdateDataLog(ctx context.Context, workspaceID string, dataLog *entity.DataLog) (err error)
	GetDataLogChildren(ctx context.Context, workspaceID string, dataLogID string) (dataLogs []*entity.DataLog, err error)
	ListDataLogs(ctx context.Context, workspaceID string, params *dto.DataLogListParams) (dataLogs []*entity.DataLog, nextToken string, code int, err error)
	ListDataLogsToReprocess(ctx context.Context, workspaceID string, lastID string, lastIDEventAt time.Time, limit int) ([]*entity.DataLog, error)
	HasDataLogsToReprocess(ctx context.Context, workspaceID string, untilDate time.Time) (foundOne bool, err error)
	CountSuccessfulDataLogsForDemo(ctx context.Context, workspaceID string) (count int64, err error)
	InsertSegmentDataLogs(ctx context.Context, workspaceID string, segmentID string, segmentVersion int, taskID string, isEnter bool, createdAt time.Time, checkpoint int) (err error)
	ListDataLogsToRespawn(ctx context.Context, workspaceID string, origin int, originID string, checkpoint int, limit int, withNextToken *string) (dataLogs []*dto.DataLogToRespawn, err error)

	// session
	FindSessionByID(ctx context.Context, workspace *entity.Workspace, sessionID string, userID string, tx *sql.Tx) (sessionFound *entity.Session, err error)
	InsertSession(ctx context.Context, session *entity.Session, tx *sql.Tx) (err error)
	UpdateSession(ctx context.Context, session *entity.Session, tx *sql.Tx) (err error)
	MergeUserSessions(ctx context.Context, workspace *entity.Workspace, fromUserID string, toUserID string, tx *sql.Tx) (err error)
	ListSessionsForUser(ctx context.Context, workspace *entity.Workspace, userID string, orderBy string, tx *sql.Tx) (sessions []*entity.Session, err error)
	ResetSessionsAttributedForConversion(ctx context.Context, userID string, conversionID string, tx *sql.Tx) (err error)
	FetchSessions(ctx context.Context, workspace *entity.Workspace, query sq.SelectBuilder, tx *sql.Tx) (sessions []*entity.Session, err error)

	// postviews
	FindPostviewByID(ctx context.Context, workspaceID *entity.Workspace, postviewID string, userID string, tx *sql.Tx) (postviewFound *entity.Postview, err error)
	InsertPostview(ctx context.Context, postview *entity.Postview, tx *sql.Tx) (err error)
	UpdatePostview(ctx context.Context, postview *entity.Postview, tx *sql.Tx) (err error)
	MergeUserPostviews(ctx context.Context, workspace *entity.Workspace, fromUserID string, toUserID string, tx *sql.Tx) (err error)
	ListPostviewsForUser(ctx context.Context, workspace *entity.Workspace, userID string, orderBy string, tx *sql.Tx) (postviews []*entity.Postview, err error)
	ResetPostviewsAttributedForConversion(ctx context.Context, userID string, conversionID string, tx *sql.Tx) (err error)

	// user aliases
	FindUserAlias(ctx context.Context, fromUserExternalID string, tx *sql.Tx) (aliasFound *entity.UserAlias, err error)
	CreateUserAlias(ctx context.Context, fromUserExternalID string, toUserExternalID string, toUserIsAuthenticated bool, tx *sql.Tx) (err error)
	FindUsersAliased(ctx context.Context, workspaceID string, toUserExternalID string) (aliases []*entity.UserAlias, err error)
	CleanAfterUserAlias(workspaceID string, fromUserExternalID string) (err error)

	// users
	FindUserByID(ctx context.Context, workspace *entity.Workspace, userID string, tx *sql.Tx) (userFound *entity.User, err error)
	FindEventualUsersToMergeWith(ctx context.Context, workspace *entity.Workspace, withUser *entity.User, withReconciliationKeys entity.MapOfInterfaces, tx *sql.Tx) (usersFound []*entity.User, err error)
	InsertUser(ctx context.Context, user *entity.User, tx *sql.Tx) (err error)
	UpdateUser(ctx context.Context, user *entity.User, tx *sql.Tx) (err error)
	ListUsers(ctx context.Context, workspace *entity.Workspace, params *dto.UserListParams) (users []*entity.User, nextToken string, previousToken string, err error)

	// devices
	FindDeviceByID(ctx context.Context, workspace *entity.Workspace, deviceID string, userID string, tx *sql.Tx) (deviceFound *entity.Device, err error)
	InsertDevice(ctx context.Context, device *entity.Device, tx *sql.Tx) (err error)
	UpdateDevice(ctx context.Context, device *entity.Device, tx *sql.Tx) (err error)
	MergeUserDevices(ctx context.Context, workspace *entity.Workspace, fromUserID string, toUserID string, tx *sql.Tx) (err error)
	ListDevicesForUser(ctx context.Context, workspace *entity.Workspace, userID string, orderBy string, tx *sql.Tx) (devices []*entity.Device, err error)
	FetchDevices(ctx context.Context, workspace *entity.Workspace, query sq.SelectBuilder, tx *sql.Tx) (devices []*entity.Device, err error)

	// segments
	InsertSegment(ctx context.Context, segment *entity.Segment, tx *sql.Tx) (err error)
	UpdateSegment(ctx context.Context, segment *entity.Segment, tx *sql.Tx) (err error)
	ListSegments(ctx context.Context, workspaceID string, withUsersCount bool) (segments []*entity.Segment, err error)
	GetSegment(ctx context.Context, workspaceID string, segmentID string) (segment *entity.Segment, err error)
	PreviewSegment(ctx context.Context, workspaceID string, parentSegmentID *string, filter *entity.SegmentTreeNode, timezone string) (count int64, sql string, args []interface{}, err error)
	DeleteSegment(ctx context.Context, workspaceID string, segmentID string) (err error)
	ActivateSegment(ctx context.Context, workspaceID string, segmentID string, segmentVersion int) (didActivate bool, err error)
	MergeUserDataLogs(ctx context.Context, workspace *entity.Workspace, fromUserID string, fromUserExternalID, toUserID string, tx *sql.Tx) (err error)

	// subscription lists
	ListSubscriptionLists(ctx context.Context, workspaceID string, withUsersCount bool) (lists []*entity.SubscriptionList, err error)
	CreateSubscriptionList(ctx context.Context, workspaceID string, list *entity.SubscriptionList) (err error)

	// message templates
	ListMessageTemplates(ctx context.Context, workspaceID string, params *dto.MessageTemplateListParams) (templates []*entity.MessageTemplate, err error)
	InsertMessageTemplate(ctx context.Context, workspaceID string, template *entity.MessageTemplate, tx *sql.Tx) (err error)
	GetMessageTemplate(ctx context.Context, workspaceID string, id string, version *int, tx *sql.Tx) (template *entity.MessageTemplate, err error)

	// user segments
	InsertUserSegment(ctx context.Context, userSegment *entity.UserSegment, tx *sql.Tx) (err error)
	DeleteUserSegment(ctx context.Context, userID string, segmentID string, tx *sql.Tx) (err error)
	ListUserSegments(ctx context.Context, workspaceID string, userIDs []string, tx *sql.Tx) (userSegments []*entity.UserSegment, err error)
	EnterUserSegmentFromQueue(ctx context.Context, workspaceID string, segmentID string, segmentVersion int) (err error)
	ExitUserSegmentFromQueue(ctx context.Context, workspaceID string, segmentID string, segmentVersion int) (err error)
	MatchSegmentUsers(ctx context.Context, workspaceID string, segment *entity.Segment, userIDs []string) (matchingUserIDs []*string, err error)

	// user segment queue
	EnqueueMatchingSegmentUsers(ctx context.Context, workspaceID string, segment *entity.Segment) (entersCount int, exitCount int, err error)
	ClearUserSegmentQueue(ctx context.Context, workspaceID string, segmentID string, segmentVersion int) (err error)
	// GetUserSegmentQueueRowsForWorker(ctx context.Context, workspaceID string, segmentID string, segmentVersion int, workerID int, limit int) (rows []*entity.UserSegmentQueue, err error)
	// GetUserSegmentQueueRow(ctx context.Context, workspaceID string, segmentID string, segmentVersion int, userID string) (row *string, err error)
	DeleteUserSegmentQueueRow(ctx context.Context, workspaceID string, segmentID string, segmentVersion int, userID string, tx *sql.Tx) (err error)

	// pageviews
	FindPageviewByID(ctx context.Context, workspace *entity.Workspace, pageviewID string, userID string, tx *sql.Tx) (pageviewFound *entity.Pageview, err error)
	InsertPageview(ctx context.Context, pageview *entity.Pageview, tx *sql.Tx) (err error)
	UpdatePageview(ctx context.Context, pageview *entity.Pageview, tx *sql.Tx) (err error)
	MergeUserPageviews(ctx context.Context, workspace *entity.Workspace, fromUserID string, toUserID string, tx *sql.Tx) (err error)

	// custom events
	FindCustomEventByID(ctx context.Context, workspace *entity.Workspace, eventID string, userID string, tx *sql.Tx) (eventFound *entity.CustomEvent, err error)
	InsertCustomEvent(ctx context.Context, event *entity.CustomEvent, tx *sql.Tx) (err error)
	UpdateCustomEvent(ctx context.Context, event *entity.CustomEvent, tx *sql.Tx) (err error)
	MergeUserCustomEvents(ctx context.Context, workspace *entity.Workspace, fromUserID string, toUserID string, tx *sql.Tx) (err error)
	ListCustomEventsForUser(ctx context.Context, workspace *entity.Workspace, userID string, orderBy string, tx *sql.Tx) (events []*entity.CustomEvent, err error)
	FetchCustomEvents(ctx context.Context, workspace *entity.Workspace, query sq.SelectBuilder, tx *sql.Tx) (events []*entity.CustomEvent, err error)
	// carts
	FindCartByID(ctx context.Context, workspaceID string, cartID string, userID string, tx *sql.Tx) (cartFound *entity.Cart, err error)
	InsertCart(ctx context.Context, cart *entity.Cart, tx *sql.Tx) (err error)
	UpdateCart(ctx context.Context, cart *entity.Cart, tx *sql.Tx) (err error)
	MergeUserCarts(ctx context.Context, workspace *entity.Workspace, fromUserID string, toUserID string, tx *sql.Tx) (err error)

	// cart items
	InsertCartItem(ctx context.Context, cartItem *entity.CartItem, tx *sql.Tx) (err error)
	UpdateCartItem(ctx context.Context, cartItem *entity.CartItem, tx *sql.Tx) (err error)
	MergeUserCartItems(ctx context.Context, workspace *entity.Workspace, fromUserID string, toUserID string, tx *sql.Tx) (err error)
	FindCartItemsByCartID(ctx context.Context, workspaceID string, cartID string, userID string, tx *sql.Tx) (cartItems []*entity.CartItem, err error)
	DeleteCartItem(ctx context.Context, cartItemID string, userID string, tx *sql.Tx) (err error)

	// orders
	FindOrderByID(ctx context.Context, workspace *entity.Workspace, orderID string, userID string, tx *sql.Tx) (orderFound *entity.Order, err error)
	InsertOrder(ctx context.Context, order *entity.Order, tx *sql.Tx) (err error)
	UpdateOrder(ctx context.Context, order *entity.Order, tx *sql.Tx) (err error)
	MergeUserOrders(ctx context.Context, workspace *entity.Workspace, fromUserID string, toUserID string, tx *sql.Tx) (err error)
	ListOrdersForUser(ctx context.Context, workspace *entity.Workspace, userID string, orderBy string, tx *sql.Tx) (orders []*entity.Order, err error)
	UpdateOrderAttribution(ctx context.Context, order *entity.Order, tx *sql.Tx) (err error)
	FindUserIDsWithOrdersToReattribute(ctx context.Context, workspaceID string, limit int) (userIDs []string, err error)

	// order items
	InsertOrderItem(ctx context.Context, orderItem *entity.OrderItem, tx *sql.Tx) (err error)
	UpdateOrderItem(ctx context.Context, orderItem *entity.OrderItem, tx *sql.Tx) (err error)
	MergeUserOrderItems(ctx context.Context, workspace *entity.Workspace, fromUserID string, toUserID string, tx *sql.Tx) (err error)
	FindOrderItemsByOrderID(ctx context.Context, workspaceID string, orderID string, userID string, tx *sql.Tx) (orderItems []*entity.OrderItem, err error)
	DeleteOrderItem(ctx context.Context, orderItemID string, userID string, tx *sql.Tx) (err error)

	// apps
	InsertApp(ctx context.Context, app *entity.App, tx *sql.Tx) (err error)
	DeleteApp(ctx context.Context, appID string, tx *sql.Tx) (err error)
	ListApps(ctx context.Context, workspaceID string) (apps []*entity.App, err error)
	GetApp(ctx context.Context, workspaceID string, appID string) (app *entity.App, err error)
	UpdateApp(ctx context.Context, app *entity.App, tx *sql.Tx) (err error)
}

type RepositoryImpl struct {
	Config *entity.Config
	DB     *sql.DB
	// second connection required to avoid memory leaks when switching between DBs in the same connection
	SystemDB *sql.DB
	// Logger     *zerolog.Logger
}

func NewRepository(cfg *entity.Config, DB *sql.DB, systemDB *sql.DB) Repository {
	return &RepositoryImpl{
		Config:   cfg,
		DB:       DB,
		SystemDB: systemDB,
		// Logger:     logger,
	}
}

func (repo *RepositoryImpl) GetSystemDB() string {
	return "system"
}

// Opens a new connection to the system DB
func (repo *RepositoryImpl) GetSystemConnection(ctx context.Context) (conn *sql.Conn, err error) {

	conn, err = repo.SystemDB.Conn(ctx)

	if err != nil {
		return nil, eris.Wrap(err, "GetSystemConnection error")
	}

	realName := repo.Config.DB_PREFIX + "system"

	// use the DB
	if _, err := conn.ExecContext(ctx, fmt.Sprintf("USE %v", realName)); err != nil {

		// create DB if doesnt exist
		if _, err := conn.ExecContext(ctx, fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %v;", realName)); err != nil {
			return nil, eris.Wrapf(err, "GetSystemConnection create DB %v error", realName)
		}

		// try again
		if _, err := conn.ExecContext(ctx, fmt.Sprintf("USE %v", realName)); err != nil {
			return nil, eris.Wrapf(err, "GetSystemConnection use DB %v error", realName)
		}
	}

	return conn, nil
}

// Opens a new connection to the workspace DB
func (repo *RepositoryImpl) GetWorkspaceConnection(ctx context.Context, workspaceID string) (conn *sql.Conn, err error) {

	conn, err = repo.DB.Conn(ctx)

	if err != nil {
		return nil, eris.Wrap(err, "GetConnection error")
	}

	realName := repo.Config.DB_PREFIX + workspaceID

	// use the DB
	if _, err := conn.ExecContext(ctx, fmt.Sprintf("USE %v", realName)); err != nil {
		return nil, eris.Wrapf(err, "GetConnection use DB %v error", realName)
	}

	return conn, nil
}

func (repo *RepositoryImpl) UseWorkspaceDBWithTx(ctx context.Context, workspaceID string, tx *sql.Tx) error {

	realName := repo.Config.DB_PREFIX + workspaceID

	// use the DB
	if _, err := tx.ExecContext(ctx, fmt.Sprintf("USE %v", realName)); err != nil {
		return eris.Wrapf(err, "UseWorkspaceDBWithTx %v error", realName)
	}

	return nil
}

func (repo *RepositoryImpl) RunInTransactionForSystem(ctx context.Context, f func(context.Context, *sql.Tx) (int, error)) (code int, err error) {

	conn, err := repo.GetSystemConnection(ctx)

	if err != nil {
		return 500, err
	}

	defer conn.Close()

	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return 500, eris.Wrap(err, "RunInTransactionForSystem")
	}

	code, err = f(ctx, tx)

	if err != nil {
		tx.Rollback()
		conn.Close()
		return code, err
	}

	if err := tx.Commit(); err != nil {
		return 500, eris.Wrap(err, "RunInTransactionForSystem")
	}

	conn.Close()

	return 200, nil
}

func (repo *RepositoryImpl) RunInTransactionForWorkspace(ctx context.Context, workspaceID string, f func(context.Context, *sql.Tx) (int, error)) (code int, err error) {

	conn, err := repo.GetWorkspaceConnection(ctx, workspaceID)

	if err != nil {
		return 500, err
	}

	defer conn.Close()

	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return 500, eris.Wrap(err, "RunInTransactionForWorkspace")
	}

	code, err = f(ctx, tx)

	if err != nil {
		tx.Rollback()
		conn.Close()
		return code, err
	}

	if err := tx.Commit(); err != nil {
		return 500, eris.Wrap(err, "RunInTransactionForWorkspace")
	}

	conn.Close()

	return code, nil
}

func (repo *RepositoryImpl) IsDuplicateEntry(err error) bool {
	return strings.Contains(err.Error(), "rror 1062")
}
