package entity

import "github.com/rotisserie/eris"

var ErrSettingsTableNotFound = eris.New("setting table not found")

type Settings struct {
	InstalledVersion float64 `json:"installed_version"` // installed version in DB (ex: 12.3)
}

var SettingsSchema string = `CREATE REFERENCE TABLE IF NOT EXISTS setting (
	installed_version FLOAT NOT NULL,
	db_created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	db_updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
	PRIMARY KEY (db_updated_at)
  );
`

var SettingsSchemaMYSQL string = `CREATE TABLE IF NOT EXISTS setting (
	installed_version FLOAT NOT NULL,
	db_created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	db_updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
	PRIMARY KEY (db_updated_at)
  );
`
