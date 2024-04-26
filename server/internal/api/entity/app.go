package entity

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/rimdian/rimdian/internal/api/common"
	"github.com/rimdian/rimdian/internal/common/auth"
	"github.com/rotisserie/eris"
)

var (
	AppStatusActive  string = "active"
	AppStatusInit    string = "initializing"
	AppStatusStopped string = "stopped"
	AppStatusDeleted string = "deleted"

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
	ID                 string          `db:"id" json:"id"`
	Name               string          `db:"name" json:"name"`
	Status             string          `db:"status" json:"status"`
	State              MapOfInterfaces `db:"state" json:"state"`
	Manifest           AppManifest     `db:"manifest" json:"manifest"`
	IsNative           bool            `db:"is_native" json:"is_native"`
	EncryptedSecretKey string          `db:"encrypted_secret_key" json:"encrypted_secret_key"`
	CreatedAt          time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt          time.Time       `db:"updated_at" json:"updated_at"`
	DeletedAt          *time.Time      `db:"deleted_at" json:"deleted_at"`

	// enriched server-side for apps
	SecretKey         string `db:"-" json:"-"`
	WorkspaceID       string `db:"-" json:"workspace_id,omitempty"`
	APIEndpoint       string `db:"-" json:"api_endpoint,omitempty"`
	CollectorEndpoint string `db:"-" json:"collector_endpoint,omitempty"`
	CubeJSEndpoint    string `db:"-" json:"cubejs_endpoint,omitempty"`
	CubeJSToken       string `db:"-" json:"cubejs_token,omitempty"`
	UIToken           string `db:"-" json:"ui_token,omitempty"` // JWT given to the app iframe to authenticate the Rimdian Console
	AccountTimezone   string `db:"-" json:"account_timezone,omitempty"`
}

func (app *App) EnrichMetadatas(cfg *Config, workspaceID string, accountID string, timezone string, withUIToken bool) (err error) {

	app.WorkspaceID = workspaceID
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

	return nil
}

func (app *App) GetAppSecretKey(serverSecretKey string) (string, error) {
	return common.DecryptFromHexString(app.EncryptedSecretKey, serverSecretKey)
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
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
	deleted_at TIMESTAMP,
	PRIMARY KEY (id)
	-- SHARD KEY (id)
  ) COLLATE utf8mb4_general_ci;
`
