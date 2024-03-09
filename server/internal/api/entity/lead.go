package entity

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"strings"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/rotisserie/eris"
)

var (
	LeadStageStatusOpen      string = "open"
	LeadStageStatusConverted string = "converted"
	LeadStageStatusLost      string = "lost"

	ErrLeadStageIDRequired        = eris.New("lead stage id is required")
	ErrLeadStageLabelRequired     = eris.New("lead stage label is required")
	ErrLeadStageColorRequired     = eris.New("lead stage color is required")
	ErrLeadStageStatusInvalid     = eris.New("lead stage status is invalid")
	ErrLeadStageMigrateIDRequired = eris.New("lead stage migrateToId is required")
	ErrLeadStageMigrateIDInvalid  = eris.New("lead stage migrateToId is not valid")
)

type LeadStages []*LeadStage

func (x *LeadStages) Scan(val interface{}) error {

	var data []byte

	if b, ok := val.([]byte); ok {
		// VERY IMPORTANT: we need to clone the bytes here
		// The sql driver will reuse the same bytes RAM slots for future queries
		// Thank you St Antoine De Padoue for helping me find this bug
		data = bytes.Clone(b)
	} else if s, ok := val.(string); ok {
		data = []byte(s)
	} else if val == nil {
		return nil
	}

	return json.Unmarshal(data, x)
}

func (x LeadStages) Value() (driver.Value, error) {
	return json.Marshal(x)
}

type LeadStage struct {
	ID          string     `json:"id"`
	Label       string     `json:"label"`
	Status      string     `json:"status"` // open | converted | lost
	Color       string     `json:"color"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
	MigrateToID *string    `json:"migrate_to_id,omitempty"` // used when delete a status, migrate existing leads to new stage ID
}

func (l *LeadStage) Validate() error {

	// sanitize
	l.ID = strings.TrimSpace(l.ID)
	l.Label = strings.TrimSpace(l.Label)
	l.Color = strings.TrimSpace(l.Color)

	if l.ID == "" {
		return ErrLeadStageIDRequired
	}
	if l.Label == "" {
		return ErrLeadStageLabelRequired
	}
	if l.Color == "" {
		return ErrLeadStageColorRequired
	}
	if !govalidator.IsIn(l.Status, LeadStageStatusOpen, LeadStageStatusConverted, LeadStageStatusLost) {
		return ErrLeadStageStatusInvalid
	}
	if l.CreatedAt.IsZero() {
		l.CreatedAt = time.Now().UTC()
	}
	if l.UpdatedAt.IsZero() {
		l.UpdatedAt = time.Now().UTC()
	}

	return nil
}

// TODO:
// keep one row per lead with most recent status
// and product a timeline_lead_update for status changes

// type Lead struct {
// 	*TimelineCommonFields
// 	*TimelineCommonInteractionFields

// 	StageId                 string                `json:"stageId"`
// 	Status                  string                `json:"status"`
// 	PublicURL               *string               `json:"publicUrl,omitempty"`
// 	LeadStageID        string                `json:"conversionRuleId"`
// 	IsFirstConversion       bool                  `json:"isFirstConversion"`
// 	TimeToConversion        int64                 `json:"timeToConversion"`
// 	Revenue                 int64                 `json:"revenue"`
// 	RevenueSource           *int64                `json:"revenueSource,omitempty"`
// 	Currency                string                `json:"currency"`
// 	CurrencyConversionError *string               `json:"currencyConversionError,omitempty"`
// 	DevicesFunnel           string                `json:"devicesFunnel"`
// 	DevicesTypeCount        int64                 `json:"devicesTypeCount"`
// 	DomainsFunnel           string                `json:"domainsFunnel"`
// 	DomainsTypeFunnel       string                `json:"domainsTypeFunnel"`
// 	DomainsCount            int64                 `json:"domainsCount"`
// 	Funnel                  ConversionFunnelSteps `json:"funnel"`
// 	FunnelHash              string                `json:"funnelHash"`
// 	AttributionUpdatedAt    *time.Time            `json:"attributionUpdatedAt,omitempty"`
// }

// var TableTimelineLeads string = `CREATE TABLE IF NOT EXISTS timeline_leads (
// 	id VARCHAR(64) NOT NULL,
// 	external_id VARCHAR(128) NOT NULL,
// 	kind VARCHAR(20) NOT NULL,
// 	created_at DATETIME NOT NULL,
// 	deleted_at DATETIME,
// 	trunc_created_at AS DATE_TRUNC('hour', created_at) PERSISTED DATETIME,
// 	user_id VARCHAR(64) NOT NULL,
// 	parent_id VARCHAR(64),
// 	merged_from_user_external_id VARCHAR(64),
// 	db_created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
// 	db_updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
// 	year TINYINT UNSIGNED NOT NULL,
// 	month TINYINT UNSIGNED NOT NULL,
// 	month_day TINYINT UNSIGNED NOT NULL,
// 	week_day TINYINT UNSIGNED NOT NULL,
// 	hour TINYINT UNSIGNED NOT NULL,

// 	domain_id VARCHAR(64),
// 	device_id VARCHAR(64),
// 	country VARCHAR(3),
// 	latitude DECIMAL(9,6),
// 	longitude DECIMAL(9,6),
// 	ip VARCHAR(45),

// 	stage_id VARCHAR(30) NOT NULL,
// 	status VARCHAR(30) NOT NULL,
// 	public_url VARCHAR(2083),
// 	is_first_conversion BOOLEAN DEFAULT FALSE,
// 	time_to_conversion INT DEFAULT 0,
// 	revenue INT DEFAULT 0,
// 	revenue_source INT DEFAULT 0,
// 	currency VARCHAR(3),
// 	currency_error VARCHAR(255),
// 	devices_funnel VARCHAR(512),
// 	devices_type_count INT DEFAULT 1,
// 	domains_funnel VARCHAR(512),
// 	domains_type_funnel VARCHAR(512),
// 	domains_count INT DEFAULT 1,
// 	funnel JSON,
// 	funnel_hash NVARCHAR(512),
// 	attribution_updated_at DATETIME,

// 	KEY (trunc_created_at) USING CLUSTERED COLUMNSTORE,
// 	PRIMARY KEY (id, user_id),
// 	KEY (stage_id),
// 	KEY (status),
// 	SHARD KEY (user_id)
//   );`

// // Leads View keeps the most recent lead stage
// var LeadsView string = `CREATE VIEW leads_view AS SELECT * FROM (
// 	SELECT *,
// 	ROW_NUMBER() OVER (PARTITION BY user_id, conversion_external_id ORDER BY created_at DESC) _rank
// 	FROM timeline_leads
// ) WHERE _rank = 1;`
