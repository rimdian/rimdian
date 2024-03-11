package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/google/uuid"
	"github.com/rimdian/rimdian/internal/api/common"
	"github.com/rimdian/rimdian/internal/api/dto"
	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rimdian/rimdian/internal/common/auth"
	"github.com/rotisserie/eris"
)

var AccessTokenDuration int = 60 * 8 // 8 hours
var ActionResetPassword = "reset-password"
var ResetPasswordTokenDuration int = 60 // in mins

// Login an account, creates a session and returns a refresh+access tokens
func (svc *ServiceImpl) AccountLogin(ctx context.Context, loginDTO *dto.AccountLogin) (loginResult *dto.AccountLoginResult, code int, err error) {

	if loginDTO == nil {
		return nil, 500, eris.Wrap(ErrServicePayloadRequired, "AccountConsumeResetPassword")
	}

	loginDTO.Email = strings.TrimSpace(loginDTO.Email)
	loginDTO.Password = strings.TrimSpace(loginDTO.Password)

	if !govalidator.IsEmail(loginDTO.Email) {
		return nil, 400, eris.Wrap(entity.ErrAccountEmailInvalid, "AccountLogin")
	}

	// fetch account
	account, err := svc.Repo.GetAccountFromEmail(ctx, loginDTO.Email)

	if err != nil {
		if eris.Is(err, entity.ErrAccountNotFound) {
			return nil, 400, err
		}
		return nil, 500, eris.Wrap(err, "AccountLogin")
	}

	// compare password hash
	if isValid := common.CheckPasswordHash(loginDTO.Password, account.HashedPassword); !isValid {
		return nil, 400, eris.Wrap(entity.ErrInvalidPassword, "AccountLogin")
	}

	now := time.Now().UTC()

	// create refresh+access tokens
	refreshTokenSessionID, err := uuid.NewRandom()
	if err != nil {
		return nil, 500, eris.Wrap(err, "Login generate uuid")
	}

	refreshTokenExpiration := now.Add(60 * 24 * 7 * time.Minute)                       // 7 days
	accessTokenExpiration := now.Add(time.Duration(AccessTokenDuration) * time.Minute) // 15 minutes

	refreshToken, err := auth.CreateAccountToken(svc.Config.SECRET_KEY, svc.Config.API_ENDPOINT, now, refreshTokenExpiration, auth.TypeRefreshToken, account.ID, refreshTokenSessionID.String())

	if err != nil {
		return nil, 500, eris.Wrap(err, "Login create refresh token")
	}

	accessToken, err := auth.CreateAccountToken(svc.Config.SECRET_KEY, svc.Config.API_ENDPOINT, now, accessTokenExpiration, auth.TypeAccessToken, account.ID, "")

	if err != nil {
		return nil, 500, eris.Wrap(err, "Login create access token")
	}

	// encrypt refresh token for DB persistence
	encryptedRefreshToken, err := common.EncryptString(refreshToken, svc.Config.SECRET_KEY)

	if err != nil {
		return nil, 500, eris.Wrap(err, "Login encrypt refresh token")
	}

	// create account_session
	accountSession := &entity.AccountSession{
		ID:                    refreshTokenSessionID.String(),
		AccountID:             account.ID,
		EncryptedRefreshToken: encryptedRefreshToken,
		ExpiresAt:             refreshTokenExpiration,
		UserAgent:             loginDTO.UserAgent,
		ClientIP:              loginDTO.ClientIP,
		LastAccessTokenAt:     now,
	}

	if err := svc.Repo.InsertAccountSession(ctx, accountSession); err != nil {
		return nil, 500, eris.Wrap(err, "Login insert account session")
	}

	loginResult = &dto.AccountLoginResult{
		Account:               account,
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: refreshTokenExpiration, // 7 days
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  accessTokenExpiration, // 15 mins
	}

	return loginResult, 200, nil
}

func (svc *ServiceImpl) AccountRefreshAccessToken(ctx context.Context, accountID string, accountSessionID string) (refreshResult *dto.AccountRefreshAccessTokenResult, code int, err error) {

	if accountSessionID == "" {
		return nil, 400, auth.ErrRefreshTokenRequired
	}

	// fetch account
	account, err := svc.Repo.GetAccountFromID(ctx, accountID)

	if err != nil {
		if eris.Is(err, entity.ErrAccountNotFound) {
			return nil, 400, err
		}
		return nil, 500, eris.Wrap(err, "AccountRefreshAccessToken")
	}

	now := time.Now().UTC()
	accessTokenExpiration := now.Add(time.Duration(AccessTokenDuration) * time.Minute) // 15 minutes

	accessToken, err := auth.CreateAccountToken(svc.Config.SECRET_KEY, svc.Config.API_ENDPOINT, now, accessTokenExpiration, auth.TypeAccessToken, account.ID, "")

	if err != nil {
		return nil, 500, eris.Wrap(err, "Login create access token")
	}

	refreshResult = &dto.AccountRefreshAccessTokenResult{
		AccessToken:          accessToken,
		AccessTokenExpiresAt: accessTokenExpiration, // 15 mins
	}

	// update LastSeenAt in account_session
	if err := svc.Repo.UpdateAccountSessionLastAccess(ctx, accountID, accountSessionID, now); err != nil {
		return nil, 500, eris.Wrap(err, "AccountRefreshAccessToken")
	}

	return refreshResult, 200, nil
}

func (svc *ServiceImpl) AccountSetProfile(ctx context.Context, accountID string, accountProfileDTO *dto.AccountProfile) (updatedAccount *entity.Account, code int, err error) {

	if accountProfileDTO == nil {
		return nil, 500, eris.Wrap(ErrServicePayloadRequired, "AccountConsumeResetPassword")
	}

	// fetch account
	account, err := svc.Repo.GetAccountFromID(ctx, accountID)

	if err != nil {
		if eris.Is(err, entity.ErrAccountNotFound) {
			return nil, 400, err
		}
		return nil, 500, eris.Wrap(err, "AccountSetProfile")
	}

	// validate profile
	accountProfileDTO.FullName = strings.TrimSpace(accountProfileDTO.FullName)

	if accountProfileDTO.FullName == "" {
		return nil, 400, entity.ErrFullNameIsRequired
	}

	if !govalidator.IsIn(accountProfileDTO.Timezone, common.Timezones...) {
		return nil, 400, entity.ErrAccountInvalidTimezone
	}

	if !govalidator.IsIn(accountProfileDTO.Locale, entity.AccountLocales...) {
		return nil, 400, entity.ErrAccountInvalidLocale
	}

	account.FullName = &accountProfileDTO.FullName
	account.Timezone = accountProfileDTO.Timezone
	account.Locale = accountProfileDTO.Locale
	account.UpdatedAt = time.Now()

	if err := svc.Repo.UpdateAccountProfile(ctx, account); err != nil {
		return nil, 500, eris.Wrap(err, "AccountSetProfile")
	}

	return account, 200, nil
}

// Generates a stateless reset-password token, sent by email to the account
// the token contains a timestamp to restrict its usage over time
func (svc *ServiceImpl) AccountResetPassword(ctx context.Context, resetDTO *dto.AccountResetPassword) (code int, err error) {

	if resetDTO == nil {
		return 500, eris.Wrap(ErrServicePayloadRequired, "AccountConsumeResetPassword")
	}

	// get account from email
	account, err := svc.Repo.GetAccountFromEmail(ctx, resetDTO.Email)

	if err != nil {
		if eris.Is(err, entity.ErrAccountNotFound) {
			return 400, err
		}
		return 500, eris.Wrap(err, "AccountResetPassword")
	}

	if account.IsServiceAccount {
		return 400, entity.ErrServiceAccountResetPassword
	}

	// reset password token is composed of b64(action_timestamp_accountID).signature
	payload := fmt.Sprintf("%v_%v_%v", ActionResetPassword, time.Now().Unix(), account.ID)
	payloadb64 := base64.URLEncoding.EncodeToString([]byte(payload))

	// sign the payload
	h := hmac.New(sha256.New, []byte(svc.Config.SECRET_KEY))
	h.Write([]byte(payload))

	resetToken := fmt.Sprintf("%v.%x", payloadb64, h.Sum(nil))

	// log token in dev to simplify testing
	if svc.Config.ENV == entity.ENV_DEV {
		svc.Logger.Printf("resetToken %v", resetToken)
	}

	if err := svc.SendSystemEmail(ctx, &dto.SystemEmail{
		To:   []string{resetDTO.Email},
		Kind: dto.EMAIL_RESET_PASSWORD,
		EmailResetPasswordPayload: dto.EmailResetPasswordPayload{
			FullName:  account.FullName,
			ActionURL: svc.Config.API_ENDPOINT + "/consume-reset-password?token=" + resetToken,
		},
	}); err != nil {
		return 500, eris.Wrap(err, "AccountResetPassword")
	}

	return 200, nil
}

// consumes a reset password token, and updates account password
func (svc *ServiceImpl) AccountConsumeResetPassword(ctx context.Context, payload *dto.AccountConsumeResetPassword) (loginResult *dto.AccountLoginResult, code int, err error) {

	if payload == nil {
		return nil, 500, eris.Wrap(ErrServicePayloadRequired, "AccountConsumeResetPassword")
	}

	// check password length
	if len(payload.NewPassword) < 8 {
		return nil, 400, entity.ErrNewPasswordInvalid
	}

	// extract the payload from the signature
	tokenParts := strings.Split(payload.Token, ".")
	if len(tokenParts) != 2 {
		return nil, 400, entity.ErrResetPasswordTokenInvalid
	}

	decoded, err := base64.URLEncoding.DecodeString(tokenParts[0])

	if err != nil {
		return nil, 500, eris.Wrap(err, "AccountConsumeResetPassword")
	}

	// sign & compare the signatures
	h := hmac.New(sha256.New, []byte(svc.Config.SECRET_KEY))
	h.Write(decoded)
	signature := fmt.Sprintf("%x", h.Sum(nil))

	if signature != tokenParts[1] {
		return nil, 400, entity.ErrResetPasswordTokenInvalid
	}

	// extract payload data parts
	dataParts := strings.Split(string(decoded), "_")
	if len(dataParts) != 3 {
		return nil, 400, entity.ErrResetPasswordTokenInvalid
	}

	// check action is a reset password
	if dataParts[0] != ActionResetPassword {
		return nil, 400, entity.ErrResetPasswordTokenInvalid
	}

	// check that the token hasnt expired yet
	timestamp, err := strconv.ParseInt(dataParts[1], 10, 64)
	if err != nil {
		return nil, 500, eris.Wrap(err, "AccountConsumeResetPassword")
	}

	tokenCreatedAt := time.Unix(timestamp, 0)
	tokenExpiresAt := tokenCreatedAt.Add(time.Duration(ResetPasswordTokenDuration) * time.Minute) // 60 minutes

	if time.Now().After(tokenExpiresAt) {
		return nil, 400, entity.ErrResetPasswordTokenExpired
	}

	accountID := dataParts[2]

	// get the account
	account, err := svc.Repo.GetAccountFromID(ctx, accountID)

	if err != nil {
		if eris.Is(err, entity.ErrAccountNotFound) {
			return nil, 400, entity.ErrAccountNotFound
		}
		return nil, 500, eris.Wrap(err, "AccountConsumeResetPassword")
	}

	// hash new password
	hashed, err := common.HashPassword(payload.NewPassword)
	if err != nil {
		return nil, 500, eris.Wrap(err, "AccountConsumeResetPassword")
	}

	if err := svc.Repo.ResetAccountPassword(ctx, accountID, hashed); err != nil {
		return nil, 500, eris.Wrap(err, "AccountConsumeResetPassword")
	}

	// return login
	return svc.AccountLogin(ctx, &dto.AccountLogin{
		Email:     account.Email,
		Password:  payload.NewPassword,
		UserAgent: payload.UserAgent,
		ClientIP:  payload.ClientIP,
	})
}

func (svc *ServiceImpl) AccountLogout(ctx context.Context, accountID string, sessionID string) (code int, err error) {

	if err := svc.Repo.AccountLogout(ctx, accountID, sessionID); err != nil {
		return 500, err
	}

	return 200, nil
}
