package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"
	"testing"
	"time"

	"aidanwoods.dev/go-paseto"
	"github.com/rimdian/rimdian/internal/api/common"
	"github.com/rimdian/rimdian/internal/api/dto"
	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rimdian/rimdian/internal/api/repository"
	"github.com/rimdian/rimdian/internal/common/auth"
	"github.com/rimdian/rimdian/internal/common/mailer"
	"github.com/rotisserie/eris"
)

var userAgent string = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:97.0) Gecko/20100101 Firefox/97.0"
var secretKey string = "12345678901234567890123456789012"

func TestServiceImpl_AccountLogin(t *testing.T) {

	t.Run("should reject invalid email", func(t *testing.T) {
		svc := &ServiceImpl{
			Config: &entity.Config{},
			Repo:   &repository.RepositoryMock{},
			Mailer: nil,
		}

		_, _, err := svc.AccountLogin(context.Background(), &dto.AccountLogin{
			Email: "wrong@email",
		})

		if !eris.Is(err, entity.ErrAccountEmailInvalid) {
			t.Errorf("should fail on invalid account email, got: %v", err)
		}
	})

	t.Run("should reject unknown account and call repo once", func(t *testing.T) {
		repoMock := &repository.RepositoryMock{
			GetAccountFromEmailFunc: func(ctx context.Context, email string) (account *entity.Account, err error) {
				return nil, entity.ErrAccountNotFound
			},
		}

		svc := &ServiceImpl{
			Config: &entity.Config{},
			Repo:   repoMock,
			Mailer: nil,
		}

		_, _, err := svc.AccountLogin(context.Background(), &dto.AccountLogin{
			Email: "unknown@account.com",
		})

		if len(repoMock.GetAccountFromEmailCalls()) != 1 {
			t.Error("should call Repo.GetAccountFromEmail one time")
		}

		if !eris.Is(err, entity.ErrAccountNotFound) {
			t.Errorf("should fail on unknown account, got: %v", err)
		}
	})

	t.Run("should fail on wrong password", func(t *testing.T) {
		password := "123"
		hashed, err := common.HashPassword(password)

		if err != nil {
			t.Error(err)
		}

		svc := &ServiceImpl{
			Config: &entity.Config{},
			Repo: &repository.RepositoryMock{
				GetAccountFromEmailFunc: func(ctx context.Context, email string) (account *entity.Account, err error) {
					return &entity.Account{
						HashedPassword: hashed,
					}, nil
				},
			},
			Mailer: nil,
		}

		_, _, err = svc.AccountLogin(context.Background(), &dto.AccountLogin{
			Email:    "unknown@account.com",
			Password: "wrong",
		})

		if !eris.Is(err, entity.ErrInvalidPassword) {
			t.Errorf("should fail on wrong password, got: %v", err)
		}
	})

	t.Run("should return valid login results and being able to refresh then token", func(t *testing.T) {
		password := "123"
		hashed, err := common.HashPassword(password)

		if err != nil {
			t.Error(err)
		}

		accountID := "test"
		clientIP := "123"

		repoMock := &repository.RepositoryMock{
			GetAccountFromEmailFunc: func(ctx context.Context, email string) (account *entity.Account, err error) {
				return &entity.Account{
					ID:             accountID,
					Email:          "good@account.com",
					HashedPassword: hashed,
				}, nil
			},
			InsertAccountSessionFunc: func(ctx context.Context, accountSession *entity.AccountSession) error {
				return nil
			},
		}

		svc := &ServiceImpl{
			Config: &entity.Config{
				SECRET_KEY: secretKey,
			},
			Repo:   repoMock,
			Mailer: nil,
		}

		results, code, err := svc.AccountLogin(context.Background(), &dto.AccountLogin{
			Email:     "good@account.com",
			Password:  password,
			UserAgent: userAgent,
			ClientIP:  clientIP,
		})

		if err != nil {
			t.Errorf("should not fail, got: %v", err)
		}

		if results == nil {
			t.Error("login results should not fail be nil")
			return
		}

		if code != 200 {
			t.Errorf("login code should be 200, got: %v", code)
		}

		if len(repoMock.InsertAccountSessionCalls()) != 1 {
			t.Error("should call Repo.InsertAccountSession one time")
		}

		// verify account Session
		session := repoMock.InsertAccountSessionCalls()[0].AccountSession

		if session.AccountID != accountID {
			t.Errorf("account id for session should be %v, got: %v", accountID, session.AccountID)
		}

		if session.ClientIP != clientIP {
			t.Errorf("account client IP for session should be %v, got: %v", clientIP, session.ClientIP)
		}

		if session.EncryptedRefreshToken == "" {
			t.Error("account encrypted refresh token for session should not be blank")
		}

		if session.UserAgent != userAgent {
			t.Errorf("account user agent for session should be %v, got: %v", userAgent, session.UserAgent)
		}

		if !session.ExpiresAt.After(time.Now()) {
			t.Errorf("account session expiration should be in the future, got: %v", session.ExpiresAt)
		}

		if session.LastAccessTokenAt.IsZero() {
			t.Error("account session last token at should not be blank")
		}

		// verify login results

		parser := paseto.NewParser()
		parser.AddRule(paseto.ForAudience(auth.TokenAccountAudience))
		parser.AddRule(paseto.NotExpired())

		key, err := paseto.V4SymmetricKeyFromBytes([]byte(svc.Config.SECRET_KEY))
		if err != nil {
			t.Errorf("error creating PASETO symmetric v4 key: %s", err.Error())
		}

		_, err = parser.ParseV4Local(key, results.AccessToken, nil)
		if err != nil {
			t.Errorf("error decrypt access token: %s", err.Error())
		}

		_, err = parser.ParseV4Local(key, results.RefreshToken, nil)
		if err != nil {
			t.Errorf("error decrypt refresh token: %s", err.Error())
		}

		// err = v2.Decrypt(results.RefreshToken, []byte(svc.Config.SECRET_KEY), &jsonToken, &footer)

		// if err != nil {
		// 	t.Errorf("decrypt refresh token err: %v", err)
		// }

		// verify account

		if results.Account == nil {
			t.Error("account login account should not be nil")
		}

		if results.Account.ID != accountID {
			t.Errorf("account id for results account should be %v, got: %v", accountID, results.Account.ID)
		}
	})
}

func TestServiceImpl_AccountRefreshAccessToken(t *testing.T) {

	t.Run("should reject empty session ID", func(t *testing.T) {
		svc := &ServiceImpl{
			Config: &entity.Config{},
			Repo:   &repository.RepositoryMock{},
			Mailer: nil,
		}

		_, _, err := svc.AccountRefreshAccessToken(context.Background(), "root", "")

		if !eris.Is(err, auth.ErrRefreshTokenRequired) {
			t.Error("should fail on empty session ID")
		}
	})

	t.Run("should reject unknown account", func(t *testing.T) {
		svc := &ServiceImpl{
			Config: &entity.Config{},
			Repo: &repository.RepositoryMock{
				GetAccountFromIDFunc: func(ctx context.Context, accountID string) (*entity.Account, error) {
					return nil, entity.ErrAccountNotFound
				},
			},
			Mailer: nil,
		}

		_, _, err := svc.AccountRefreshAccessToken(context.Background(), "unknown", "sessionid")

		if !eris.Is(err, entity.ErrAccountNotFound) {
			t.Error("should fail on unknown account")
		}
	})

	t.Run("should return valid results and updates session", func(t *testing.T) {
		accountID := "root"
		sessionID := "sessionid"

		repoMock := &repository.RepositoryMock{
			GetAccountFromIDFunc: func(ctx context.Context, accountID string) (*entity.Account, error) {
				return &entity.Account{
					ID: accountID,
				}, nil
			},
			UpdateAccountSessionLastAccessFunc: func(ctx context.Context, accountID, accountSessionID string, now time.Time) error {
				return nil
			},
		}

		svc := &ServiceImpl{
			Config: &entity.Config{
				SECRET_KEY: secretKey,
			},
			Repo:   repoMock,
			Mailer: nil,
		}

		results, code, err := svc.AccountRefreshAccessToken(context.Background(), accountID, sessionID)

		if err != nil {
			t.Errorf("should not fail, got: %v", err)
		}

		if results == nil {
			t.Error("refresh token results should not be nil")
			return
		}

		if code != 200 {
			t.Errorf("refresh token code should be 200, got: %v", code)
		}

		if len(repoMock.UpdateAccountSessionLastAccessCalls()) != 1 {
			t.Error("should call Repo.UpdateAccountSessionLastAccess one time")
		}

		// verify account Session
		if repoMock.UpdateAccountSessionLastAccessCalls()[0].AccountID != accountID {
			t.Errorf("account id for session should be %v, got: %v", accountID, repoMock.UpdateAccountSessionLastAccessCalls()[0].AccountID)
		}

		if repoMock.UpdateAccountSessionLastAccessCalls()[0].AccountSessionID != sessionID {
			t.Errorf("account session ID should be %v, got: %v", sessionID, repoMock.UpdateAccountSessionLastAccessCalls()[0].AccountSessionID)
		}

		// verify results

		parser := paseto.NewParser()
		parser.AddRule(paseto.ForAudience(auth.TokenAccountAudience))
		parser.AddRule(paseto.NotExpired())

		key, err := paseto.V4SymmetricKeyFromBytes([]byte(svc.Config.SECRET_KEY))
		if err != nil {
			t.Errorf("error creating PASETO symmetric v4 key: %s", err.Error())
		}

		_, err = parser.ParseV4Local(key, results.AccessToken, nil)
		if err != nil {
			t.Errorf("error decrypt access token: %s", err.Error())
		}

	})
}

func TestServiceImpl_AccountSetProfile(t *testing.T) {

	t.Run("should reject unknown account", func(t *testing.T) {
		svc := &ServiceImpl{
			Config: &entity.Config{},
			Repo: &repository.RepositoryMock{
				GetAccountFromIDFunc: func(ctx context.Context, accountID string) (*entity.Account, error) {
					return nil, entity.ErrAccountNotFound
				},
			},
			Mailer: nil,
		}

		_, _, err := svc.AccountSetProfile(context.Background(), "unknown", &dto.AccountProfile{})

		if !eris.Is(err, entity.ErrAccountNotFound) {
			t.Error("should fail on unknown account")
		}
	})

	t.Run("should reject empty name", func(t *testing.T) {
		svc := &ServiceImpl{
			Config: &entity.Config{},
			Repo: &repository.RepositoryMock{
				GetAccountFromIDFunc: func(ctx context.Context, accountID string) (*entity.Account, error) {
					return &entity.Account{}, nil
				},
			},
			Mailer: nil,
		}

		_, _, err := svc.AccountSetProfile(context.Background(), "root", &dto.AccountProfile{})

		if !eris.Is(err, entity.ErrFullNameIsRequired) {
			t.Error("should fail on empty name")
		}
	})

	t.Run("should reject invalid timezone", func(t *testing.T) {
		svc := &ServiceImpl{
			Config: &entity.Config{},
			Repo: &repository.RepositoryMock{
				GetAccountFromIDFunc: func(ctx context.Context, accountID string) (*entity.Account, error) {
					return &entity.Account{}, nil
				},
			},
			Mailer: nil,
		}

		_, _, err := svc.AccountSetProfile(context.Background(), "root", &dto.AccountProfile{FullName: "John", Timezone: "wrong"})

		if !eris.Is(err, entity.ErrAccountInvalidTimezone) {
			t.Error("should fail on invalid timezone")
		}
	})

	t.Run("should update account profile", func(t *testing.T) {
		accountID := "root"

		repoMock := &repository.RepositoryMock{
			GetAccountFromIDFunc: func(ctx context.Context, accountID string) (*entity.Account, error) {
				return &entity.Account{
					ID: accountID,
				}, nil
			},
			UpdateAccountProfileFunc: func(ctx context.Context, account *entity.Account) error {
				return nil
			},
		}

		svc := &ServiceImpl{
			Config: &entity.Config{},
			Repo:   repoMock,
			Mailer: nil,
		}

		account, code, err := svc.AccountSetProfile(context.Background(), "root", &dto.AccountProfile{
			FullName: "John",
			Timezone: "UTC",
			Locale:   "fr-FR",
		})

		if err != nil {
			t.Errorf("should not fail, got: %v", err)
		}

		if account == nil {
			t.Error("should return nil account")
			return
		}

		if code != 200 {
			t.Errorf("code should be 200, got: %v", code)
		}

		if len(repoMock.UpdateAccountProfileCalls()) != 1 {
			t.Error("should call Repo.UpdateAccountProfile one time")
		}

		// verify account returned
		if repoMock.UpdateAccountProfileCalls()[0].Account.ID != accountID {
			t.Errorf("account id should be %v, got: %v", accountID, repoMock.UpdateAccountProfileCalls()[0].Account.ID)
		}

	})
}

func TestServiceImpl_AccountResetPassword(t *testing.T) {

	t.Run("should reject unknown account", func(t *testing.T) {
		svc := &ServiceImpl{
			Config: &entity.Config{},
			Repo: &repository.RepositoryMock{
				GetAccountFromEmailFunc: func(ctx context.Context, email string) (*entity.Account, error) {
					return nil, entity.ErrAccountNotFound
				},
			},
			Mailer: nil,
		}

		_, err := svc.AccountResetPassword(context.Background(), &dto.AccountResetPassword{Email: "unknown"})

		if !eris.Is(err, entity.ErrAccountNotFound) {
			t.Error("should fail on unknown account")
		}
	})

	t.Run("should send a token to the mailer", func(t *testing.T) {
		accountID := "test"
		email := "email@account.com"

		mailerMock := &mailer.MailerMock{
			SendHTMLMailFunc: func(to []string, subject, body string) error {
				return nil
			},
		}

		svc := &ServiceImpl{
			Config: &entity.Config{},
			Repo: &repository.RepositoryMock{
				GetAccountFromEmailFunc: func(ctx context.Context, email string) (*entity.Account, error) {
					return &entity.Account{ID: accountID}, nil
				},
			},
			Mailer: mailerMock,
		}

		code, err := svc.AccountResetPassword(context.Background(), &dto.AccountResetPassword{Email: email})

		if err != nil {
			t.Errorf("should not fail, got: %v", err)
		}

		if code != 200 {
			t.Errorf("code should be 200, got: %v", code)
		}

		if len(mailerMock.SendHTMLMailCalls()) != 1 {
			t.Error("should call Repo.SendHTMLMail one time")
		}

		// verify email payload sent
		if mailerMock.SendHTMLMailCalls()[0].To[0] != email {
			t.Errorf("email recipient should be %v, got: %v", email, mailerMock.SendHTMLMailCalls()[0].To[0])
		}
	})
}

func TestServiceImpl_AccountConsumeResetPassword(t *testing.T) {

	accountID := "john"
	goodPassword := "strongPASSWORD45678098765{¶«¡Ç"

	payload := fmt.Sprintf("%v_%v_%v", ActionResetPassword, time.Now().Unix(), accountID)
	payloadb64 := base64.URLEncoding.EncodeToString([]byte(payload))
	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write([]byte(payload))
	goodToken := fmt.Sprintf("%v.%x", payloadb64, h.Sum(nil))

	t.Run("should reject weak password", func(t *testing.T) {
		svc := &ServiceImpl{
			Config: &entity.Config{
				SECRET_KEY: secretKey,
			},
			Repo:   &repository.RepositoryMock{},
			Mailer: nil,
		}

		_, _, err := svc.AccountConsumeResetPassword(context.Background(), &dto.AccountConsumeResetPassword{NewPassword: "a"})

		if !eris.Is(err, entity.ErrNewPasswordInvalid) {
			t.Errorf("should fail on weak password, got %v", err)
		}
	})

	t.Run("should reject invalid token", func(t *testing.T) {
		svc := &ServiceImpl{
			Config: &entity.Config{},
			Repo:   &repository.RepositoryMock{},
			Mailer: nil,
		}

		_, _, err := svc.AccountConsumeResetPassword(context.Background(), &dto.AccountConsumeResetPassword{NewPassword: goodPassword, Token: "foo"})

		if !eris.Is(err, entity.ErrResetPasswordTokenInvalid) {
			t.Error("should fail on invalid token")
		}
	})

	t.Run("should reject invalid b64 token", func(t *testing.T) {
		svc := &ServiceImpl{
			Config: &entity.Config{},
			Repo:   &repository.RepositoryMock{},
			Mailer: nil,
		}

		_, _, err := svc.AccountConsumeResetPassword(context.Background(), &dto.AccountConsumeResetPassword{NewPassword: goodPassword, Token: "foo.bar"})

		if !strings.Contains(err.Error(), "illegal base64") {
			t.Errorf("should fail on invalid token, got %v", err)
		}
	})

	t.Run("should reject invalid token payload", func(t *testing.T) {
		svc := &ServiceImpl{
			Config: &entity.Config{
				SECRET_KEY: secretKey,
			},
			Repo:   &repository.RepositoryMock{},
			Mailer: nil,
		}

		payload := fmt.Sprintf("%v_%v_%v", "bad-action", time.Now().Unix(), accountID)
		payloadb64 := base64.URLEncoding.EncodeToString([]byte(payload))

		// sign the payload
		h := hmac.New(sha256.New, []byte(svc.Config.SECRET_KEY))
		h.Write([]byte(payload))

		token := fmt.Sprintf("%v.%x", payloadb64, h.Sum(nil))

		_, _, err := svc.AccountConsumeResetPassword(context.Background(), &dto.AccountConsumeResetPassword{NewPassword: goodPassword, Token: token})

		if !eris.Is(err, entity.ErrResetPasswordTokenInvalid) {
			t.Errorf("should fail on invalid token, got %v", err)
		}
	})

	t.Run("should reject expired token", func(t *testing.T) {
		svc := &ServiceImpl{
			Config: &entity.Config{
				SECRET_KEY: secretKey,
			},
			Repo:   &repository.RepositoryMock{},
			Mailer: nil,
		}

		payload := fmt.Sprintf("%v_%v_%v", ActionResetPassword, time.Now().AddDate(0, 0, -1).Unix(), accountID)
		payloadb64 := base64.URLEncoding.EncodeToString([]byte(payload))

		// sign the payload
		h := hmac.New(sha256.New, []byte(svc.Config.SECRET_KEY))
		h.Write([]byte(payload))

		token := fmt.Sprintf("%v.%x", payloadb64, h.Sum(nil))

		_, _, err := svc.AccountConsumeResetPassword(context.Background(), &dto.AccountConsumeResetPassword{NewPassword: goodPassword, Token: token})

		if !eris.Is(err, entity.ErrResetPasswordTokenExpired) {
			t.Errorf("should fail on expired token, got %v", err)
		}
	})

	t.Run("should fail on unknown account", func(t *testing.T) {
		svc := &ServiceImpl{
			Config: &entity.Config{
				SECRET_KEY: secretKey,
			},
			Repo: &repository.RepositoryMock{
				GetAccountFromIDFunc: func(ctx context.Context, accountID string) (*entity.Account, error) {
					return nil, entity.ErrAccountNotFound
				},
			},
			Mailer: nil,
		}

		_, _, err := svc.AccountConsumeResetPassword(context.Background(), &dto.AccountConsumeResetPassword{NewPassword: goodPassword, Token: goodToken})

		if !eris.Is(err, entity.ErrAccountNotFound) {
			t.Errorf("should fail on unknown account, got %v", err)
		}
	})

	t.Run("should find account and fail on not found", func(t *testing.T) {
		repoMock := &repository.RepositoryMock{
			GetAccountFromIDFunc: func(ctx context.Context, accountID string) (*entity.Account, error) {
				return nil, entity.ErrAccountNotFound
			},
		}

		svc := &ServiceImpl{
			Config: &entity.Config{
				SECRET_KEY: secretKey,
			},
			Repo:   repoMock,
			Mailer: nil,
		}

		_, _, err := svc.AccountConsumeResetPassword(context.Background(), &dto.AccountConsumeResetPassword{NewPassword: goodPassword, Token: goodToken})

		if len(repoMock.GetAccountFromIDCalls()) != 1 {
			t.Error("should call GetAccountFromID only once")
		}

		if repoMock.GetAccountFromIDCalls()[0].AccountID != accountID {
			t.Error("should call GetAccountFromID with proper account id")
		}

		if !eris.Is(err, entity.ErrAccountNotFound) {
			t.Errorf("should fail on unknown account, got %v", err)
		}
	})
}

func TestServiceImpl_AccountLogout(t *testing.T) {
	// TODO
}
