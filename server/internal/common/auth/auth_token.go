package auth

import (
	"context"
	"time"

	"aidanwoods.dev/go-paseto"
	"github.com/rotisserie/eris"
	"github.com/teris-io/shortid"
)

var TypeRefreshToken string = "refreshToken"
var TypeAccessToken string = "accessToken"
var TokenAccountAudience string = "account"

type ContextKey string

var AccountTokenContextKey ContextKey = "accountToken"
var AccountRawTokenContextKey ContextKey = "accountRawToken"

type AccountTokenClaims struct {
	// *jwt.Claims
	Type      string `json:"type"` // refreshToken | accessToken
	AccountID string `json:"account_id"`
	SessionID string `json:"session_id,omitempty"` // account_session ID for refresh tokens
}

func (tc *AccountTokenClaims) Validate() error {
	if tc.Type != TypeAccessToken && tc.Type != TypeRefreshToken {
		return ErrInvalidTokenType
	}
	if tc.AccountID == "" {
		return ErrAccountIDIsMissing
	}
	if tc.Type == TypeRefreshToken && tc.SessionID == "" {
		return ErrRefreshTokenSessionIDIsMissing
	}
	return nil
}

var (
	ErrRefreshTokenRequired           = eris.New("a refresh token is required")
	ErrSecretKeyIsMissing             = eris.New("secret key is missing to create account token")
	ErrSecretKeyShouldBe32bytes       = eris.New("secret key should be 32 characters long")
	ErrAccountIDIsMissing             = eris.New("accountID is missing to create account token")
	ErrInvalidTokenType               = eris.New("invalid token type to create account token")
	ErrRefreshTokenSessionIDIsMissing = eris.New("refreshTokenSessionID is missing to create account token")
)

// create a PASETO token for account
func CreateAccountToken(secretKey string, issuer string, issuedAt time.Time, expiresAt time.Time, tokenType string, accountID string, refreshTokenSessionID string) (token string, err error) {

	if secretKey == "" {
		return "", ErrSecretKeyIsMissing
	}
	if len(secretKey) != 32 {
		return "", ErrSecretKeyShouldBe32bytes
	}
	if accountID == "" {
		return "", ErrAccountIDIsMissing
	}
	if tokenType != TypeRefreshToken && tokenType != TypeAccessToken {
		return "", ErrInvalidTokenType
	}
	if tokenType == TypeRefreshToken && refreshTokenSessionID == "" {
		return "", ErrRefreshTokenSessionIDIsMissing
	}

	symmetricKey := []byte(secretKey) // Must be 32 bytes

	sid, err := shortid.New(1, shortid.DefaultABC, 777)
	if err != nil {
		return "", eris.Wrap(err, "CreateAccountToken shortid")
	}

	jti, err := sid.Generate()
	if err != nil {
		return "", eris.Wrap(err, "CreateAccountToken shortid")
	}

	pasetoToken := paseto.NewToken()

	pasetoToken.SetAudience(TokenAccountAudience)
	pasetoToken.SetIssuer(issuer)
	pasetoToken.SetJti(jti)
	pasetoToken.SetSubject(accountID)
	pasetoToken.SetIssuedAt(time.Now())
	pasetoToken.SetExpiration(expiresAt)

	// Add custom claim    to the token
	pasetoToken.SetString("type", tokenType)

	if tokenType == TypeRefreshToken {
		pasetoToken.SetString("session_id", refreshTokenSessionID)
	}

	key, err := paseto.V4SymmetricKeyFromBytes(symmetricKey)
	if err != nil {
		return "", eris.Wrap(err, "CreateAccountToken V4SymmetricKeyFromBytes")
	}

	return pasetoToken.V4Encrypt(key, nil), nil
}

func GetAccountTokenClaimsFromContext(ctx context.Context) (claims *AccountTokenClaims, err error) {

	// check that ctx value is not nil
	eventualToken := ctx.Value(AccountTokenContextKey)
	if eventualToken == nil {
		return nil, ErrRefreshTokenRequired
	}

	token := eventualToken.(*paseto.Token)

	// log.Printf("claims: %+v\n", token.Claims())

	tokenType, err := token.GetString("type")
	if err != nil {
		return nil, eris.Wrap(err, "GetAccountTokenClaimsFromContext")
	}

	subject, err := token.GetSubject()
	if err != nil {
		return nil, eris.Wrap(err, "GetAccountTokenClaimsFromContext")
	}

	claims = &AccountTokenClaims{
		Type:      tokenType,
		AccountID: subject,
	}

	if tokenType == TypeRefreshToken {
		claims.SessionID, err = token.GetString("session_id")
		if err != nil {
			return nil, eris.Wrap(err, "GetAccountTokenClaimsFromContext")
		}
	}

	if err := claims.Validate(); err != nil {
		return nil, eris.Wrap(err, "GetAccountTokenClaimsFromContext")
	}
	return claims, nil
}

func GetAccountRawTokenFromContext(ctx context.Context) (token string) {
	eventualToken := ctx.Value(AccountRawTokenContextKey)
	if eventualToken == nil {
		// in tests, the token is not set
		return ""
	}
	return eventualToken.(string)
}
