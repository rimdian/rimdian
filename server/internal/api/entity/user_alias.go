package entity

import (
	"strings"
	"time"

	"github.com/rotisserie/eris"
	"github.com/tidwall/gjson"
)

var (
	ErrUserAliasAlreadyExists = eris.New("user alias already exists")
)

type UserAlias struct {
	DBCreatedAt           time.Time `db:"db_created_at" json:"db_created_at"`
	FromUserExternalID    string    `db:"from_user_external_id" json:"from_user_external_id"`
	ToUserExternalID      string    `db:"to_user_external_id" json:"to_user_external_id"`
	ToUserIsAuthenticated bool      `db:"to_user_is_authenticated" json:"to_user_is_authenticated"`
	// server-side fields:
	ToUserCreatedAt *time.Time `db:"-" json:"-"`
}

func NewUserAliasFromDataLog(dataLog *DataLog, clockDifference time.Duration, dataSentAt *time.Time, dataReceivedAt time.Time) (userAlias *UserAlias, err error) {

	result := gjson.Get(dataLog.Item, "user_alias")
	if !result.Exists() {
		return nil, eris.New("item has no user_alias object")
	}

	if result.Type != gjson.JSON {
		return nil, eris.New("item user_alias is not an object")
	}

	// init
	userAlias = &UserAlias{}

	// loop over fields
	result.ForEach(func(key, value gjson.Result) bool {

		keyString := key.String()
		switch keyString {

		case "from_user_external_id":
			userAlias.FromUserExternalID = strings.TrimSpace(value.String())

		case "to_user_external_id":
			userAlias.ToUserExternalID = strings.TrimSpace(value.String())

		case "to_user_is_authenticated":
			if value.Type != gjson.Null {
				userAlias.ToUserIsAuthenticated = value.Bool()
			}

		case "to_user_created_at":

			if value.Type != gjson.Null {
				toUserCreatedAt, err := time.Parse(time.RFC3339, value.String())

				if err != nil {
					err = eris.Wrap(err, "user_alias.to_user_created_at")
					return false
				}

				// apply clock difference
				if toUserCreatedAt.After(time.Now()) {

					toUserCreatedAt = toUserCreatedAt.Add(clockDifference)
					if toUserCreatedAt.After(time.Now()) {
						err = eris.New("session.to_user_created_at cannot be in the future")
						return false
					}
				}

				userAlias.ToUserCreatedAt = &toUserCreatedAt
			}

		default:
		}

		return true
	})

	if err != nil {
		return nil, err
	}

	// set default values
	if userAlias.ToUserCreatedAt == nil {
		userAlias.ToUserCreatedAt = dataSentAt
	}

	if userAlias.ToUserCreatedAt == nil {
		userAlias.ToUserCreatedAt = &dataReceivedAt
	}

	// required fields
	if userAlias.FromUserExternalID == "" {
		return nil, eris.New("user_alias.from_user_external_id is required")
	}

	if userAlias.ToUserExternalID == "" {
		return nil, eris.New("user_alias.to_user_external_id is required")
	}

	if userAlias.ToUserCreatedAt == nil {
		return nil, eris.New("user_alias.to_user_created_at is required")
	}

	return userAlias, nil
}

var UserAliasSchema string = `CREATE ROWSTORE TABLE IF NOT EXISTS user_alias (
  from_user_external_id VARCHAR(64) NOT NULL,
  to_user_external_id VARCHAR(64) NOT NULL,
  to_user_is_authenticated BOOLEAN NOT NULL,
  db_created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

  PRIMARY KEY (from_user_external_id, to_user_external_id),
  KEY (to_user_external_id),
  SHARD KEY (from_user_external_id)
) COLLATE utf8mb4_general_ci;`

var UserAliasSchemaMYSQL string = `CREATE TABLE IF NOT EXISTS user_alias (
  from_user_external_id VARCHAR(64) NOT NULL,
  to_user_external_id VARCHAR(64) NOT NULL,
  to_user_is_authenticated BOOLEAN NOT NULL,
  db_created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

  PRIMARY KEY (from_user_external_id, to_user_external_id),
  KEY (to_user_external_id)
  -- SHARD KEY (from_user_external_id)
) COLLATE utf8mb4_general_ci;`
