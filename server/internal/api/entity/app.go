package entity

import (
	"time"

	"github.com/rimdian/rimdian/internal/api/common"
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

	// joined server-side
	UIToken           string `db:"-" json:"ui_token,omitempty"` // JWT given to the app iframe to authenticate the Rimdian Console
	CollectorEndpoint string `db:"-" json:"collector_endpoint,omitempty"`
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
