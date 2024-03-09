package entity

import "time"

type AccountSession struct {
	ID                    string     `db:"id" json:"id"`
	AccountID             string     `db:"account_id" json:"account_id"`
	EncryptedRefreshToken string     `db:"encrypted_refresh_token" json:"encrypted_refresh_token"`
	ExpiresAt             time.Time  `db:"expires_at" json:"expires_at"`
	UserAgent             string     `db:"user_agent" json:"user_agent"`
	ClientIP              string     `db:"client_ip" json:"client_ip"`
	LastAccessTokenAt     time.Time  `db:"last_access_token_at" json:"last_access_token_at"`
	BlockedAt             *time.Time `db:"blocked_at" json:"blocked_at"`
	CreatedAt             time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt             time.Time  `db:"updated_at" json:"updated_at"`
}

var AccountSessionsSchema string = `CREATE ROWSTORE TABLE IF NOT EXISTS account_session (
	id VARCHAR(64) NOT NULL,
	account_id VARCHAR(64) NOT NULL,
	encrypted_refresh_token VARCHAR(1024) NOT NULL,
	expires_at DATETIME NOT NULL,
	user_agent VARCHAR(128) NOT NULL,
	client_ip VARCHAR(64) NOT NULL,
	last_access_token_at DATETIME,
	blocked_at DATETIME,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
	PRIMARY KEY (id),
	KEY (account_id),
	SHARD KEY (id)
  ) COLLATE utf8mb4_general_ci;
`

var AccountSessionsSchemaMYSQL string = `CREATE TABLE IF NOT EXISTS account_session (
	id VARCHAR(64) NOT NULL,
	account_id VARCHAR(64) NOT NULL,
	encrypted_refresh_token VARCHAR(1024) NOT NULL,
	expires_at DATETIME NOT NULL,
	user_agent VARCHAR(128) NOT NULL,
	client_ip VARCHAR(64) NOT NULL,
	last_access_token_at DATETIME,
	blocked_at DATETIME,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
	PRIMARY KEY (id),
	KEY (account_id)
	-- SHARD KEY (id)
  ) COLLATE utf8mb4_general_ci;
`
