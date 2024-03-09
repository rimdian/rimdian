package entity

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/rotisserie/eris"
)

var (
	ErrInvalidInvitationToken                 = eris.New("invitation token is invalid")
	ErrInvalidInvitationTokenSignature        = eris.New("invitation token signature is invalid")
	ErrAccountEmailDoesntMatchInvitationEmail = eris.New("authenticated account email doesn't match invitation email")
	ErrInvitationConsumedOrDeleted            = eris.New("This invitation has already been consumed or deleted")
	ErrAccountAlreadyInOrganization           = eris.New("account already belongs to this organization")
	ErrInvitationHasBeenConsumedOrCancelled   = eris.New("invitation has been consumed or cancelled")
	ErrInvitationHasExpired                   = eris.New("invitation has expired")
)

type OrganizationInvitation struct {
	Email            string           `db:"email" json:"email"`
	OrganizationID   string           `db:"organization_id" json:"organization_id"`
	FromAccountID    string           `db:"from_account_id"  json:"from_account_id"`
	ExpiresAt        time.Time        `db:"expires_at" json:"expires_at"`
	ConsumedAt       *time.Time       `db:"consumed_at" json:"consumed_at,omitempty"`
	WorkspacesScopes WorkspacesScopes `db:"workspaces_scopes" json:"workspaces_scopes"`
	CreatedAt        time.Time        `db:"created_at" json:"created_at"`
	UpdateddAt       time.Time        `db:"updated_at" json:"updated_at"`
}

var OrganizationInvitationSchema string = `CREATE ROWSTORE TABLE IF NOT EXISTS organization_invitation (
	email VARCHAR(64) NOT NULL,
	organization_id VARCHAR(64) NOT NULL,
	from_account_id VARCHAR(128) NOT NULL,
	expires_at DATETIME NOT NULL,
	workspaces_scopes JSON NOT NULL,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
	consumed_at DATETIME,
	
	PRIMARY KEY (email, organization_id),
    SHARD KEY (email, organization_id)
  ) COLLATE utf8mb4_general_ci;
`

var OrganizationInvitationSchemaMYSQL string = `CREATE TABLE IF NOT EXISTS organization_invitation (
	email VARCHAR(64) NOT NULL,
	organization_id VARCHAR(64) NOT NULL,
	from_account_id VARCHAR(128) NOT NULL,
	expires_at DATETIME NOT NULL,
	workspaces_scopes JSON NOT NULL,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
	consumed_at DATETIME,
	
	PRIMARY KEY (email, organization_id)
    -- SHARD KEY (email, organization_id)
  ) COLLATE utf8mb4_general_ci;
`

func CreateInvitationToken(email string, organizationID string, secretKey string) string {

	payload := fmt.Sprintf("%v~%v", email, organizationID)
	payloadb64 := base64.URLEncoding.EncodeToString([]byte(payload))

	// sign the payload
	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write([]byte(payload))

	return fmt.Sprintf("%v.%x", payloadb64, h.Sum(nil))
}

func DecodeAndVerifyInvitationToken(token string, secretKey string) (email string, organizationID string, code int, err error) {

	// log.Printf("token %v, key %v", token, secretKey)

	// decide & verify token
	parts := strings.Split(token, ".")

	if len(parts) != 2 {
		return "", "", 400, ErrInvalidInvitationToken
	}

	// invitation is base64 URL encoded email + org ID
	decoded, err := base64.URLEncoding.DecodeString(parts[0])

	if err != nil {
		return "", "", 500, eris.Wrap(err, "OrganizationConsumeInvitation")
	}

	// the invitation token contains the invitation ID + its hmac256 signature
	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write(decoded)

	signature := fmt.Sprintf("%x", h.Sum(nil))

	// log.Println(signature)

	if signature != parts[1] {
		return "", "", 400, ErrInvalidInvitationTokenSignature
	}

	decodedParts := strings.Split(string(decoded), "~")

	if len(decodedParts) != 2 {
		return "", "", 400, ErrInvalidInvitationToken
	}

	email = decodedParts[0]
	organizationID = decodedParts[1]

	return email, organizationID, 200, nil
}
