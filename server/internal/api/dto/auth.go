package dto

import (
	"time"

	"github.com/rimdian/rimdian/internal/api/entity"
)

type AccountLogin struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	UserAgent string `json:"user_agent"`
	ClientIP  string `json:"client_ip"`
}

type AccountLoginResult struct {
	Account               *entity.Account `json:"account"`
	RefreshToken          string          `json:"refresh_token"`
	RefreshTokenExpiresAt time.Time       `json:"refresh_token_expires_at"`
	AccessToken           string          `json:"access_token"`
	AccessTokenExpiresAt  time.Time       `json:"access_token_expires_at"`
}

type AccountRefreshAccessTokenResult struct {
	AccessToken          string    `json:"access_token"`
	AccessTokenExpiresAt time.Time `json:"access_token_expires_at"`
}
