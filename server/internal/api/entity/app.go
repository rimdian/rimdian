package entity

import (
	"crypto/sha1"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/golang-jwt/jwt/v4"
	"github.com/rimdian/rimdian/internal/api/common"
	"github.com/rimdian/rimdian/internal/common/auth"
	"github.com/rotisserie/eris"
)

var (
	AppStatusActive    string = "active"
	AppStatusInit      string = "initializing"
	AppStatusStopped   string = "stopped"
	AppStatusUpgrading string = "upgrading"
	AppStatusDeleted   string = "deleted"

	ErrAppAlreadyExists = eris.New("app already exists")
	ErrAppNotActive     = eris.New("app is not active")
)

type AppUITokenClaim struct {
	*jwt.RegisteredClaims
	WorkspaceID       string `json:"workspace_id"`
	AppID             string `json:"app_id"`
	APIEndpoint       string `json:"api_endpoint"`
	CollectorEndpoint string `json:"collector_endpoint"`
	AccountTimezone   string `json:"account_timezone"`
	AccountID         string `json:"account_id"` // admin account id that his consuming the app
}

type App struct {
	ID                   string          `db:"id" json:"id"`
	Name                 string          `db:"name" json:"name"`
	Status               string          `db:"status" json:"status"`
	State                MapOfInterfaces `db:"state" json:"state"`
	Manifest             AppManifest     `db:"manifest" json:"manifest"`
	IsNative             bool            `db:"is_native" json:"is_native"`
	EncryptedSecretKey   string          `db:"encrypted_secret_key" json:"encrypted_secret_key"`
	SQLUsername          string          `db:"sql_username" json:"sql_username"`
	EncryptedSQLPassword string          `db:"encrypted_sql_password" json:"encrypted_sql_password"`
	CreatedAt            time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt            time.Time       `db:"updated_at" json:"updated_at"`
	DeletedAt            *time.Time      `db:"deleted_at" json:"deleted_at"`

	// enriched server-side for apps
	SecretKey         string        `db:"-" json:"-"`
	WorkspaceID       string        `db:"-" json:"workspace_id,omitempty"`
	WorkspaceCurrency string        `db:"-" json:"workspace_currency,omitempty"`
	APIEndpoint       string        `db:"-" json:"api_endpoint,omitempty"`
	CollectorEndpoint string        `db:"-" json:"collector_endpoint,omitempty"`
	CubeJSEndpoint    string        `db:"-" json:"cubejs_endpoint,omitempty"`
	CubeJSToken       string        `db:"-" json:"cubejs_token,omitempty"`
	UIToken           string        `db:"-" json:"ui_token,omitempty"` // JWT given to the app iframe to authenticate the Rimdian Console
	AccountTimezone   string        `db:"-" json:"account_timezone,omitempty"`
	SQLAccess         *AppSQLAccess `db:"-" json:"sql_access,omitempty"`
}

func (app *App) DecryptSQLPassword(secretKey string) (password string, err error) {
	return common.DecryptFromHexString(app.EncryptedSQLPassword, secretKey)
}

func (app *App) GenerateSQLCredentials(workspaceID string, secretKey string) (password string, err error) {

	// generate SQL username, hash of workspaceID and appID
	username := fmt.Sprintf("%x", sha1.Sum([]byte(fmt.Sprintf("%s_%s", workspaceID, app.ID))))

	// keep first 16 characters
	app.SQLUsername = "app_" + username[:16]

	// generate SQL password
	password = common.RandomString(32)

	encrypted, err := common.EncryptString(password, secretKey)

	if err != nil {
		return password, eris.Wrap(err, "GenerateSQLCredentials")
	}

	app.EncryptedSQLPassword = encrypted

	return password, nil
}

func (app *App) EnrichMetadatas(cfg *Config, workspaceCurrency string, workspaceID string, accountID string, timezone string, withUIToken bool) (err error) {

	app.WorkspaceID = workspaceID
	app.WorkspaceCurrency = workspaceCurrency
	app.CollectorEndpoint = cfg.COLLECTOR_ENDPOINT
	app.APIEndpoint = cfg.API_ENDPOINT
	app.CubeJSEndpoint = cfg.CUBEJS_ENDPOINT
	app.AccountTimezone = timezone

	// generate an accountToken
	now := time.Now().UTC()
	accessTokenExpiration := now.Add(time.Duration(AccessTokenDuration) * time.Minute)

	accessToken, err := auth.CreateAccountToken(cfg.SECRET_KEY, cfg.API_ENDPOINT, now, accessTokenExpiration, auth.TypeAccessToken, accountID, "")

	if err != nil {
		return err
	}

	app.CubeJSToken, err = GenerateCubeJSToken(cfg, workspaceID, accessToken)

	if err != nil {
		return err
	}

	// decrypt secret key
	app.SecretKey, err = common.DecryptFromHexString(app.EncryptedSecretKey, cfg.SECRET_KEY)
	if err != nil {
		return err
	}

	tokenToSign := jwt.NewWithClaims(jwt.SigningMethodHS256, &AppUITokenClaim{
		&jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(6 * time.Hour)), // 12 hours
		},
		workspaceID,
		app.ID,
		cfg.API_ENDPOINT,
		cfg.COLLECTOR_ENDPOINT,
		timezone,
		accountID,
	})

	// if request comes from a UI token we don't give another one to avoid infinite expiration loop
	if withUIToken {
		app.UIToken, err = tokenToSign.SignedString([]byte(app.SecretKey))
		if err != nil {
			return err
		}
	}

	// enrich SQLAccess
	if app.SQLUsername != "" {
		app.SQLAccess = &AppSQLAccess{
			User:     app.SQLUsername,
			Database: cfg.DB_PREFIX + workspaceID,
		}

		// parse config DSN to extract host and port
		sqlConfig, err := mysql.ParseDSN(cfg.DB_DSN)

		if err != nil {
			return err
		}

		app.SQLAccess.Host = sqlConfig.Addr
		app.SQLAccess.Port = "3306"

		if strings.Contains(sqlConfig.Addr, ":") {
			host, port, err := net.SplitHostPort(sqlConfig.Addr)
			if err != nil {
				return err
			}
			app.SQLAccess.Host = host
			app.SQLAccess.Port = port
		}

		// decrypt SQL password
		app.SQLAccess.Password, err = common.DecryptFromHexString(app.EncryptedSQLPassword, app.SecretKey)
		if err != nil {
			return err
		}
	}

	return nil
}

func (app *App) GetAppSecretKey(serverSecretKey string) (string, error) {
	return common.DecryptFromHexString(app.EncryptedSecretKey, serverSecretKey)
}

type AppSQLAccess struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Database string `json:"database"`
}

type AppStateMutation struct {
	Operation string      `json:"operation"` // set | delete
	Key       string      `json:"key"`
	Value     interface{} `json:"value"`
}

func (app *App) ApplyMutations(mutations []*AppStateMutation) {
	if app.State == nil {
		app.State = MapOfInterfaces{}
	}

	for _, mutation := range mutations {
		if mutation.Operation == "set" {
			app.State[mutation.Key] = mutation.Value
		}
		if mutation.Operation == "delete" {
			delete(app.State, mutation.Key)
		}
	}
}

var AppSchema string = `CREATE ROWSTORE TABLE IF NOT EXISTS app (
	id VARCHAR(64) NOT NULL,
	-- namespace VARCHAR(25) NOT NULL,
	name VARCHAR(60) NOT NULL,
	status VARCHAR(20) NOT NULL,
	state JSON NOT NULL,
	manifest JSON NOT NULL,
	is_native BOOLEAN NOT NULL,
	encrypted_secret_key VARCHAR(255) NOT NULL,
	sql_username VARCHAR(32) NOT NULL,
	encrypted_sql_password VARCHAR(512),
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
	deleted_at TIMESTAMP,
	PRIMARY KEY (id),
	SHARD KEY (id)
  ) COLLATE utf8mb4_general_ci;
`

var AppSchemaMYSQL string = `CREATE TABLE IF NOT EXISTS app (
	id VARCHAR(64) NOT NULL,
	-- namespace VARCHAR(25) NOT NULL,
	name VARCHAR(60) NOT NULL,
	status VARCHAR(20) NOT NULL,
	state JSON NOT NULL,
	manifest JSON NOT NULL,
	is_native BOOLEAN NOT NULL,
	encrypted_secret_key VARCHAR(255) NOT NULL,
	sql_username VARCHAR(32) NOT NULL,
	encrypted_sql_password VARCHAR(512),
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
	deleted_at TIMESTAMP,
	PRIMARY KEY (id)
	-- SHARD KEY (id)
  ) COLLATE utf8mb4_general_ci;
`
