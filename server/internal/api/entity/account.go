package entity

import (
	"time"

	"github.com/rimdian/rimdian/internal/api/common"
	"github.com/rotisserie/eris"
)

var (
	ErrAccountEmailInvalid             = eris.New("invalid email")
	ErrFullNameIsRequired              = eris.New("fullName is required")
	ErrInvalidAccountID                = eris.New("invalid account id")
	ErrInvalidPassword                 = eris.New("invalid password")
	ErrAccountDeactivated              = eris.New("account deactivated")
	ErrAccountAlreadyExists            = eris.New("account already exists")
	ErrAccountEmailAlreadyUsed         = eris.New("account email already used")
	ErrAccountInvalidTimezone          = eris.New("account timezone is not valid")
	ErrAccountInvalidLocale            = eris.New("account locale is not valid")
	ErrAccountNotFound                 = eris.New("account not found")
	ErrServiceAccountNameIsRequired    = eris.New("service account name is required")
	ErrServiceAccountEmailIDIsRequired = eris.New("service account emailId is required")
	ErrServiceAccountPasswordIsInvalid = eris.New("service account password should contain at least 16 characters")
	ErrServiceAccountResetPassword     = eris.New("a service account password cannot be changed")

	ErrResetPasswordTokenInvalid = eris.New("reset password token is not valid")
	ErrResetPasswordTokenExpired = eris.New("reset password token has expired")
	ErrNewPasswordInvalid        = eris.New("the new password should contain at least 8 characters")

	AccountLocales = []string{
		"en-US",
		"fr-FR",
	}
)

type Account struct {
	ID               string    `db:"id" json:"id"`
	FullName         *string   `db:"full_name" json:"full_name,omitempty"`
	Timezone         string    `db:"timezone" json:"timezone"`
	Email            string    `db:"email" json:"email"`
	Locale           string    `db:"locale" json:"locale"` // en-US | fr-FR
	HashedPassword   string    `db:"hashed_password" json:"-"`
	IsServiceAccount bool      `db:"is_service_account" json:"is_service_account"`
	IsRoot           bool      `db:"is_root" json:"-"`
	CreatedAt        time.Time `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time `db:"updated_at" json:"updated_at"`
}

// generate root account from config
func GenerateRootAccount(cfg *Config) (account *Account, err error) {
	account = &Account{
		ID:        "root",
		Timezone:  "UTC",
		Email:     cfg.ROOT_EMAIL,
		Locale:    "en-US",
		IsRoot:    true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	pwd, err := common.HashPassword("root")
	if err != nil {
		return nil, eris.Wrap(err, "GenerateRootAccount error")
	}

	account.HashedPassword = pwd

	return account, nil
}

var AccountSchema string = `CREATE ROWSTORE TABLE IF NOT EXISTS account (
	id VARCHAR(64) NOT NULL,
	full_name VARCHAR(50),
	timezone VARCHAR(50) NOT NULL,
	email VARCHAR(64) NOT NULL,
	locale VARCHAR(10) NOT NULL DEFAULT 'en-US',
	hashed_password VARCHAR(256),
	is_service_account BOOLEAN NOT NULL DEFAULT FALSE,
	is_root BOOLEAN DEFAULT FALSE,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
	PRIMARY KEY (id),
	SHARD KEY (id)
  ) COLLATE utf8mb4_general_ci;
`

var AccountSchemaMYSQL string = `CREATE TABLE IF NOT EXISTS account (
	id VARCHAR(64) NOT NULL,
	full_name VARCHAR(50),
	timezone VARCHAR(50) NOT NULL,
	email VARCHAR(64) NOT NULL,
	locale VARCHAR(10) NOT NULL DEFAULT 'en-US',
	hashed_password VARCHAR(256),
	is_service_account BOOLEAN NOT NULL DEFAULT FALSE,
	is_root BOOLEAN DEFAULT FALSE,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
	PRIMARY KEY (id)
	-- SHARD KEY (id)
  ) COLLATE utf8mb4_general_ci;
`
