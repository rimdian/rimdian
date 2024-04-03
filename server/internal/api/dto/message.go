package dto

import "time"

// the SendMessage object should contain everything to send a message
// without having to do any additional lookups
type SendMessage struct {
	WorkspaceID         string            `json:"workspace_id"`
	MessageID           string            `json:"message_id"`
	MessageExternalID   string            `json:"message_external_id"`
	UserID              string            `json:"user_id"`
	UserExternalID      string            `json:"user_external_id"`
	UserIsAuthenticated bool              `json:"user_is_authenticated"`
	Channel             string            `json:"channel"` // email | sms | push..
	ScheduledAt         *time.Time        `json:"scheduled_at,omitempty"`
	Email               *SendMessageEmail `json:"email"`
}

type SendMessageEmail struct {
	Subject               string                 `json:"subject"`
	HTML                  string                 `json:"html"`
	Text                  string                 `json:"text,omitempty"`
	IsTransactional       bool                   `json:"is_transactional"`
	Provider              string                 `json:"provider"` // sparkpost ...
	SparkPostCrendentials *SparkPostCrendentials `json:"sparkpost,omitempty"`
	SMTPCredentials       *SMTPCredentials       `json:"smtp,omitempty"`
}

type SparkPostCrendentials struct {
	EncryptedApiKey string `json:"encrypted_api_key"`
}

type SMTPCredentials struct {
	Host              string `json:"host"`
	Port              int    `json:"port"`
	EncryptedUsername string `json:"encrypted_username"`
	EncryptedPassword string `json:"encrypted_password"`
	Encryption        string `json:"encryption"` // tls | ssl
}
