package dto

import (
	"time"

	"github.com/rimdian/rimdian/internal/api/entity"
)

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
	FromAdrress     string               `json:"from_address"`
	FromName        string               `json:"from_name"`
	ReplyTo         *string              `json:"reply_to,omitempty"`
	ToAdrress       string               `json:"to_address"`
	Subject         string               `json:"subject"`
	HTML            string               `json:"html"`
	Text            string               `json:"text,omitempty"`
	IsTransactional bool                 `json:"is_transactional"`
	EmailProvider   entity.EmailProvider `json:"email_provider"`
}

// https://developers.sparkpost.com/api/transmissions/#transmissions-post-send-inline-content
type SparkPostAddress struct {
	Email string  `json:"email"`
	Name  *string `json:"name,omitempty"`
}

type SparkPostRecipient struct {
	Address          SparkPostAddress       `json:"address"`
	Tags             []string               `json:"tags,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
	SubstitutionData map[string]interface{} `json:"substitution_data,omitempty"`
}

type SparkPostcontent struct {
	From    SparkPostAddress `json:"from"`
	Subject string           `json:"subject"`
	ReplyTo *string          `json:"reply_to,omitempty"`
	Headers struct {
		XCustomerCampaignID string `json:"X-Customer-Campaign-ID"`
	} `json:"headers,omitempty"`
	HTML *string `json:"html,omitempty"`
	Text *string `json:"text,omitempty"`
}

type SparkPostOptions struct {
	ClickTracking bool   `json:"click_tracking"`
	OpenTracking  bool   `json:"open_tracking"`
	Transactional bool   `json:"transactional"`
	IPPool        string `json:"ip_pool"`
	InlineCSS     bool   `json:"inline_css"`
}

type SparkPostMessage struct {
	Options SparkPostOptions `json:"options"`
	// Description string           `json:"description"`
	// CampaignID  string           `json:"campaign_id"`
	// Metadata    struct {
	// 	UserType       string `json:"user_type"`
	// 	EducationLevel string `json:"education_level"`
	// } `json:"metadata"`
	// SubstitutionData struct {
	// 	Sender      string `json:"sender"`
	// 	HolidayName string `json:"holiday_name"`
	// } `json:"substitution_data"`
	Recipients []SparkPostRecipient `json:"recipients"`
	Content    SparkPostcontent     `json:"content"`
}
