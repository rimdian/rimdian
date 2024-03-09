package dto

import (
	"database/sql"
	"time"

	"github.com/rimdian/rimdian/internal/api/entity"
)

type UserAliasParams struct {
	Workspace              *entity.Workspace
	DataImportID           string
	FromUserExternalID     string
	ToUserExternalID       string       // destination user
	ToUserIsAuthenticated  bool         // destination user
	FromUser               *entity.User // eventually existing from user
	ToUser                 *entity.User // eventually existing to user
	ToUserDefaultCreatedAt time.Time    // default user created_at if user doesnt exist yet
	Tx                     *sql.Tx      // current transaction
}
