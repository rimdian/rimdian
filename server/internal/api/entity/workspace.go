package entity

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"aidanwoods.dev/go-paseto"
	"github.com/asaskevich/govalidator"
	"github.com/golang-jwt/jwt/v4"
	"github.com/rimdian/rimdian/internal/api/common"
	"github.com/rimdian/rimdian/internal/common/auth"
	"github.com/rotisserie/eris"
)

type CubeClaim struct {
	*jwt.RegisteredClaims
	WorkspaceID string `json:"workspace_id"`
	SchemaURL   string `json:"schema_url"`
}

var (
	UserIDSigningNone          string = "none"
	UserIDSigningAuthenticated string = "authenticated"
	UserIDSigningAll           string = "all"

	WorkspaceDemoOrder string = "order"
	WorkspaceDemoLead  string = "lead"

	WorkspaceTestingSecretKey string = "testing_secret_key"

	ErrWorkspaceIDRequired                     = eris.New("workspace id is required")
	ErrWorkspaceIDPrefixInvalid                = eris.New("workspace id prefix is not valid")
	ErrWorkspaceIDInvalid                      = eris.New("workspace id is not valid")
	ErrWorkspaceIDReserved                     = eris.New("workspace id is reserved")
	ErrWorkspaceNameRequired                   = eris.New("workspace name is required")
	ErrWorkspaceURLInvalid                     = eris.New("workspace websiteURL is not valid")
	ErrWorkspacePrivacyURLInvalid              = eris.New("workspace privacyPolicyURL is not valid")
	ErrWorkspaceIndustryInvalid                = eris.New("workspace industry is not valid")
	ErrWorkspace                               = eris.New("workspace setupStep is not valid")
	ErrWorkspaceCurrency                       = eris.New("workspace currency is not valid")
	ErrWorkspaceOrganizationIDInvalid          = eris.New("workspace organizationId is not valid")
	ErrWorkspaceUserReconciliationKeysRequired = eris.New("workspace user reconciliation keys is required")
	ErrWorkspaceUserReconciliationKeyInvalid   = eris.New("workspace user reconciliation key is not valid")
	ErrWorkspaceSecretKeyRequired              = eris.New("a secret key is required in the workspace")
	ErrWorkspaceUserTimezoneInvalid            = eris.New("workspace defaultUserTimezone is not valid")
	ErrWorkspaceUserCountryInvalid             = eris.New("workspace defaultUserCountry is not valid")
	ErrWorkspaceUserLanguageInvalid            = eris.New("workspace defaultUserLanguage is not valid")
	ErrWorkspaceSignUserIDInvalid              = eris.New("workspace userIdSigning is not valid")
	ErrWorkspaceSessionTimeoutInvalid          = eris.New("workspace sessionTimeout should be a positive int")
	ErrWorkspaceBounceThresholdInvalid         = eris.New("workspace bounceThreshold should be a positive int")
	ErrUserHMACInvalid                         = eris.New("userHmac/invalid")
	ErrWorkspaceAlreadyExists                  = eris.New("workspace already exists")
	ErrWorkspaceHasNoConversions               = eris.New("workspace has no conversions")
	ErrWorkspaceFxRateNotFound                 = eris.New("workspace fx rate not found")

	ErrWorkspaceLeadStagesRequired = eris.New("conversion rule lead stages required")
)

func ExtractOrganizationIDFromWorkspaceID(workspaceID string) string {
	parts := strings.Split(workspaceID, "_")

	// wrong workspace id
	if len(parts) < 2 {
		return ""
	}

	return parts[0]
}

type LicenseInfo struct {
	// use short names to reduce licnese token payload size
	UserSegmentsQuota  int64 `json:"usq"`
	DataLogsOver90Days int64 `json:"dlo90"`
	HasAdminRoles      bool  `json:"ar"`
}

type Workspace struct {
	ID               string     `db:"id" json:"id"`
	Name             string     `db:"name" json:"name"`
	CreatedAt        time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time  `db:"updated_at" json:"updated_at"`
	DeletedAt        *time.Time `db:"deleted_at" json:"deleted_at,omitempty"`
	IsDemo           bool       `db:"is_demo" json:"is_demo"`
	DemoKind         *string    `db:"demo_kind" json:"demo_kind,omitempty"`
	WebsiteURL       string     `db:"website_url" json:"website_url"`
	PrivacyPolicyURL string     `db:"privacy_policy_url" json:"privacy_policy_url"`
	Industry         string     `db:"industry" json:"industry"`
	Currency         string     `db:"currency" json:"currency"`
	OrganizationID   string     `db:"organization_id" json:"organization_id"`
	// DataProtectionOfficerID string     `db:"dpo_id" json:"dataProtectionOfficerId"`
	SecretKeys                     SecretKeys             `db:"secret_keys" json:"-"`
	DefaultUserTimezone            string                 `db:"default_user_timezone" json:"default_user_timezone"`
	DefaultUserCountry             string                 `db:"default_user_country" json:"default_user_country"`
	DefaultUserLanguage            string                 `db:"default_user_language" json:"default_user_language"`
	UserReconciliationKeys         UserReconciliationKeys `db:"user_reconciliation_keys" json:"user_reconciliation_keys"`
	UserIDSigning                  string                 `db:"user_id_signing" json:"user_id_signing"` // none / authenticated / all
	SessionTimeout                 int                    `db:"session_timeout" json:"session_timeout"`
	AbandonedCartsProcessedUntil   *time.Time             `db:"abandoned_carts_processed_until" json:"abandoned_carts_processed_until,omitempty"`
	Domains                        Domains                `db:"domains" json:"domains"`
	Channels                       Channels               `db:"channels" json:"channels"`
	ChannelGroups                  ChannelGroups          `db:"channel_groups" json:"channel_groups"`
	InstalledApps                  InstalledApps          `db:"installed_apps" json:"installed_apps"` // copy of apps table to optimize requests
	HasOrders                      bool                   `db:"has_orders" json:"has_orders"`
	HasLeads                       bool                   `db:"has_leads" json:"has_leads"`
	LeadStages                     LeadStages             `db:"lead_stages" json:"lead_stages"`
	OutdatedConversionsAttribution bool                   `db:"outdated_conversions_attribution" json:"outdated_conversions_attribution"` // flag to indicate that a task for reattributing conversions should be launched
	FxRates                        *FxRates               `db:"fx_rates" json:"fx_rates"`
	DataHooks                      DataHooks              `db:"data_hooks" json:"data_hooks"`
	LicenseKey                     *string                `db:"license_key" json:"license_key,omitempty"`

	// Attached server-side
	CubeJSToken string       `json:"cubejs_token,omitempty"`
	LicenseInfo *LicenseInfo `json:"license_info,omitempty"`
}

func (w *Workspace) AttachMetadatas(ctx context.Context, cfg *Config) (err error) {

	// CubeJS token
	accountToken := auth.GetAccountRawTokenFromContext(ctx)

	schemaURL := cfg.API_ENDPOINT + "/api/cubejs.schemas?workspace_id=" + w.ID + "&rmd_token=" + accountToken

	// generate a CubeJS JWT token
	tokenToSign := jwt.NewWithClaims(jwt.SigningMethodHS256, &CubeClaim{
		&jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(8 * time.Hour)), // 8 hours
		},
		w.ID,
		schemaURL,
	})

	w.CubeJSToken, err = tokenToSign.SignedString([]byte(cfg.SECRET_KEY))

	if err != nil {
		return err
	}

	// default license info
	w.LicenseInfo = &LicenseInfo{
		UserSegmentsQuota:  5,
		DataLogsOver90Days: 10000000, // 10 million
		HasAdminRoles:      false,
	}

	// license info
	if w.LicenseKey != nil {

		publicKey, err := paseto.NewV4AsymmetricPublicKeyFromHex(cfg.LICENSE_PUBLIC_KEY) // this wil fail if given key in an invalid format

		// silent error
		if err != nil {
			log.Printf("error creating license public key: %s", err.Error())
			return nil
		}

		// parse license token
		parser := paseto.NewParser()
		// each Rimdian deployment has a different API endpoint
		parser.AddRule(paseto.ForAudience(cfg.API_ENDPOINT))
		parser.AddRule(paseto.Subject(w.ID))
		parser.AddRule(paseto.NotExpired())

		token, err := parser.ParseV4Public(publicKey, *w.LicenseKey, nil)

		// silent error
		if err != nil {
			log.Printf("error validating license token: %s", err.Error())
			return nil
		}

		claims := token.Claims()

		if usq, ok := claims["usq"]; ok {
			// convert string to int64
			quota, err := strconv.ParseInt(usq.(string), 10, 64)
			if err != nil {
				log.Printf("error parsing license usq: %s", err.Error())
			} else {
				w.LicenseInfo.UserSegmentsQuota = quota
			}
		}

		if dlo90, ok := claims["dlo90"]; ok {
			// convert string to int64
			quota90days, err := strconv.ParseInt(dlo90.(string), 10, 64)
			if err != nil {
				log.Printf("error parsing license dlo90: %s", err.Error())
			} else {
				w.LicenseInfo.DataLogsOver90Days = quota90days
			}
		}

		if ar, ok := claims["ar"]; ok {
			// convert string to bool
			hasAdminRoles, err := strconv.ParseBool(ar.(string))
			if err != nil {
				log.Printf("error parsing license ar: %s", err.Error())
			} else {
				w.LicenseInfo.HasAdminRoles = hasAdminRoles
			}
		}
	}

	return nil
}

func (p *Workspace) Validate() error {

	// sanitize
	p.ID = strings.TrimSpace(p.ID)
	p.Name = strings.TrimSpace(p.Name)
	p.WebsiteURL = strings.TrimSpace(p.WebsiteURL)
	p.PrivacyPolicyURL = strings.TrimSpace(p.PrivacyPolicyURL)

	if p.ID == "" {
		return ErrWorkspaceIDRequired
	}
	// reserved table names
	if govalidator.IsIn(p.ID, "system") {
		return ErrWorkspaceIDReserved
	}
	if !strings.HasPrefix(p.ID, p.OrganizationID+"_") {
		return ErrWorkspaceIDPrefixInvalid
	}
	// check that ID has only one underscore
	if strings.Count(p.ID, "_") != 1 {
		return ErrWorkspaceIDInvalid
	}
	if p.Name == "" {
		return ErrWorkspaceNameRequired
	}
	if p.CreatedAt.IsZero() {
		p.CreatedAt = time.Now().UTC()
	}
	if p.UpdatedAt.IsZero() {
		p.UpdatedAt = time.Now().UTC()
	}
	if !govalidator.IsRequestURL(p.WebsiteURL) {
		return ErrWorkspaceURLInvalid
	}
	if !govalidator.IsRequestURL(p.PrivacyPolicyURL) {
		return ErrWorkspacePrivacyURLInvalid
	}
	if !govalidator.IsIn(p.Industry, WorkspaceIndustries...) {
		return ErrWorkspaceIndustryInvalid
	}
	// if !govalidator.IsIn(p.SetupStep, WorkspaceSetupStepNew, WorkspaceSetupStepDone) {
	// 	return eris.New("workspace setupStep is not valid")
	// }
	if !govalidator.IsIn(p.Currency, common.CurrenciesCodes...) {
		return ErrWorkspaceCurrency
	}
	if p.OrganizationID == "" {
		return ErrWorkspaceOrganizationIDInvalid
	}
	// init to avoid null in JSON db field
	if p.UserReconciliationKeys == nil {
		p.UserReconciliationKeys = []string{}
	}
	if p.Domains == nil {
		p.Domains = []*Domain{}
	}
	if p.Channels == nil {
		p.Channels = []*Channel{}
	}
	if p.ChannelGroups == nil {
		p.ChannelGroups = []*ChannelGroup{}
	}
	if p.InstalledApps == nil {
		p.InstalledApps = InstalledApps{}
	}
	// if p.ObservabilityGroups == nil {
	// 	p.ObservabilityGroups = DefaultObservabilityGroups
	// }
	if !p.HasOrders && !p.HasLeads {
		log.Printf("%+v", p)
		return ErrWorkspaceHasNoConversions
	}
	if p.HasLeads {

		if len(p.LeadStages) == 0 {
			return ErrWorkspaceLeadStagesRequired
		}

		for _, stage := range p.LeadStages {
			if err := stage.Validate(); err != nil {
				return err
			}

			if stage.DeletedAt != nil {
				if stage.MigrateToID == nil {
					return ErrLeadStageMigrateIDRequired
				}

				// check if migration stage id exists
				var destinationStage *LeadStage
				for _, x := range p.LeadStages {
					if x.ID != stage.ID && x.ID == *stage.MigrateToID {
						destinationStage = x
					}
				}

				if destinationStage == nil {
					return ErrLeadStageMigrateIDInvalid
				}
			}

		}
	}
	if len(p.UserReconciliationKeys) == 0 {
		return ErrWorkspaceUserReconciliationKeysRequired
	}
	if p.SecretKeys == nil || len(p.SecretKeys) == 0 {
		return ErrWorkspaceSecretKeyRequired
	}
	if !govalidator.IsIn(p.DefaultUserTimezone, common.Timezones...) {
		return ErrWorkspaceUserTimezoneInvalid
	}
	if !govalidator.IsIn(p.DefaultUserCountry, common.CountriesCodes...) {
		return ErrWorkspaceUserCountryInvalid
	}
	if !govalidator.IsIn(p.DefaultUserLanguage, common.LanguageCodes...) {
		return ErrWorkspaceUserLanguageInvalid
	}
	if !govalidator.IsIn(p.UserIDSigning, UserIDSigningNone, UserIDSigningAuthenticated, UserIDSigningAll) {
		return ErrWorkspaceSignUserIDInvalid
	}
	if p.SessionTimeout <= 0 {
		return ErrWorkspaceSessionTimeoutInvalid
	}
	for _, key := range p.UserReconciliationKeys {
		// extra app key
		if strings.HasPrefix(key, "app_") || strings.HasPrefix(key, "appx_") {
			// regex validate custom column name
			re := regexp.MustCompile(CustomColumnNameRegex)

			if !re.MatchString(key) {
				return ErrWorkspaceUserReconciliationKeyInvalid
			}
			continue
		}

		// native key
		if !govalidator.IsIn(key, NativeReconciliationKeys...) {
			return ErrWorkspaceUserReconciliationKeyInvalid
		}
	}

	return nil
}

// verify a user HMAC signature against workspace secret keys
func (p *Workspace) VerifyUserHmac(userId string, userHmac string) (int, error) {

	for _, k := range p.SecretKeys {

		h := hmac.New(sha256.New, []byte(k.Key))

		// Write Data to it
		h.Write([]byte(userId))

		hmacString := fmt.Sprintf("%x", h.Sum(nil))

		// check at least 8 characters are valid to shorten hmacs in URLs
		if !strings.Contains(hmacString, userHmac) {
			return 400, ErrUserHMACInvalid
		}
	}

	return 200, nil
}

func (ws *Workspace) GetFxRateForCurrency(currency string) (float64, error) {
	if ws.Currency == currency {
		return 1, nil
	}
	if ws.FxRates == nil {
		return 0, ErrFxRatesOutdated
	}
	rate := 1.0

	// base currency is EUR
	if ws.Currency == "EUR" {
		if fxRate, ok := ws.FxRates.Rates[currency]; ok {
			// invert rate
			rate = 1 / fxRate
			return rate, nil
		} else {
			return 0, ErrFxRatesOutdated
		}
	} else {
		// base currency is not EUR
		// check if we have a rate for the base currency
		if fxRate, ok := ws.FxRates.Rates[ws.Currency]; ok {
			rate = fxRate
		} else {
			return 0, ErrWorkspaceFxRateNotFound
		}

		// check if we have a rate for the target currency
		if currency != "EUR" {
			if fxRate, ok := ws.FxRates.Rates[currency]; ok {
				rate = rate / fxRate
				return rate, nil
			} else {
				return 0, eris.Wrapf(ErrTargetFxRateNotFound, "(%v)", currency)
			}
		}
		return rate, nil
	}
}

// get active secret key
func (ws *Workspace) GetActiveSecretKey(cfgSecretKey string) (activeKey *SecretKey, err error) {
	if err = ws.DecryptSecretKeys(cfgSecretKey); err != nil {
		return
	}
	return ws.SecretKeys.GetActive(), nil
}

// decrypts encrypted secret keys
func (p *Workspace) DecryptSecretKeys(cfgSecretKey string) error {

	for i, k := range p.SecretKeys {
		// decrypt key
		decryptedKey, err := common.DecryptFromHexString(k.EncryptedKey, cfgSecretKey)

		if err != nil {
			return eris.Wrap(err, "DecryptSecretKeys error")
		}

		k.Key = decryptedKey
		p.SecretKeys[i] = k
	}

	return nil
}

func (p *Workspace) GetAppTableDefinition(kind string) *AppTableManifest {
	for _, app := range p.InstalledApps {
		for _, t := range app.AppTables {
			if t.Name == kind {
				return t
			}
		}
	}
	return nil
}

func (w *Workspace) FindAppTableDefinitionForItem(itemKind string) (tableFound *AppTableManifest) {

	for _, app := range w.InstalledApps {
		for _, table := range app.AppTables {
			if table.Name == itemKind {
				return table
			}
		}
	}

	return nil
}

func (w *Workspace) FindExtraColumnsForItemKind(itemKind string) (columnsFound []*TableColumn) {
	columnsFound = []*TableColumn{}
	for _, app := range w.InstalledApps {
		for _, augTable := range app.ExtraColumns {
			if augTable.Kind == itemKind {
				for _, col := range augTable.Columns {
					columnsFound = append(columnsFound, col)
				}
			}
		}
	}
	return columnsFound
}

type UserReconciliationKeys []string

func (x *UserReconciliationKeys) Scan(val interface{}) error {

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

func (x UserReconciliationKeys) Value() (driver.Value, error) {
	return json.Marshal(x)
}

type SecretKey struct {
	EncryptedKey string     `json:"encrypted_key"`
	CreatedAt    time.Time  `json:"created_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`
	// not persisted in DB
	Key string `json:"-"`
}

type SecretKeys []*SecretKey

func (x *SecretKeys) Scan(val interface{}) error {

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

func (x SecretKeys) Value() (driver.Value, error) {
	return json.Marshal(x)
}

func (keys SecretKeys) GetActive() (keyFound *SecretKey) {
	for _, k := range keys {
		if k.DeletedAt == nil {
			keyFound = k
		}
	}
	return
}

func GenerateDemoWorkspace(workspaceID string, demoKind string, organizationID string, cfgSecretKey string) (workspace *Workspace, err error) {

	workspace = &Workspace{
		// use a determined ID to avoid creating multiple demos per org and polluting the DB
		ID:                     workspaceID,
		IsDemo:                 true,
		DemoKind:               &demoKind,
		Currency:               "USD",
		OrganizationID:         organizationID,
		DefaultUserTimezone:    "America/New_York",
		DefaultUserCountry:     "US",
		DefaultUserLanguage:    "en",
		UserReconciliationKeys: DefaultUserReconciliationKeys,
		UserIDSigning:          UserIDSigningNone,
		SessionTimeout:         1800,
		ChannelGroups:          ChannelGroups{},
		Channels:               Channels{},
		Domains:                Domains{},
		// ObservabilityGroups:    DefaultObservabilityGroups,
		HasOrders:     false,
		HasLeads:      false,
		InstalledApps: InstalledApps{},
		DataHooks:     DataHooks{},
	}

	// add default channels & groups
	now := time.Now().UTC()

	for _, gr := range DefaultChannelGroups {
		gr.CreatedAt = now
		gr.UpdatedAt = now
		workspace.ChannelGroups = append(workspace.ChannelGroups, gr)
	}

	for _, ch := range DefaultChannels {
		ch.CreatedAt = now
		ch.UpdatedAt = now
		if err := ch.Validate(workspace.Channels, workspace.ChannelGroups); err != nil {
			return nil, eris.Wrap(err, "WorkspaceCreateOrResetDemo")
		}
		workspace.Channels = append(workspace.Channels, ch)
	}

	if demoKind == WorkspaceDemoOrder {

		workspace.Name = "Demo Apple Store"
		workspace.WebsiteURL = "https://www.apple.com/"
		workspace.PrivacyPolicyURL = "https://www.apple.com/privacy/privacy-policy"
		workspace.Industry = WorkspaceIndustries[5]
		workspace.HasOrders = true

		// ecommerce channels mapping

		couponDiscountDescription := "15% discount"

		gizmodo := &Channel{
			ID:   "gizmodo",
			Name: "Gizmodo.com",
			Origins: []*ChannelOrigin{
				{
					ID:            "gizmodo.com / referral",
					MatchOperator: OriginMatchOperatorEqual,
					UTMSource:     "gizmodo.com",
					UTMMedium:     "referral",
				},
			},
			VoucherCodes: []*VoucherCode{
				{
					Code:        "GIZ15",
					OriginID:    "gizmodo.com / referral",
					Description: &couponDiscountDescription,
				},
			},
			GroupID:   "referral",
			CreatedAt: now,
			UpdatedAt: now,
		}

		if err := gizmodo.Validate(workspace.Channels, workspace.ChannelGroups); err != nil {
			return nil, eris.Wrap(err, "WorkspaceCreateOrResetDemo")
		}

		workspace.Channels = append(workspace.Channels, gizmodo)

		couponShippingDescription := "Free shipping"
		techcrunch := &Channel{
			ID:   "techcrunch",
			Name: "Techcrunch.com",
			Origins: []*ChannelOrigin{
				{
					ID:            "techcrunch.com / referral",
					MatchOperator: OriginMatchOperatorEqual,
					UTMSource:     "techcrunch.com",
					UTMMedium:     "referral",
				},
			},
			VoucherCodes: []*VoucherCode{
				{
					Code:        "SHIPSHIP",
					OriginID:    "techcrunch.com / referral",
					Description: &couponShippingDescription,
				},
			},
			GroupID:   "referral",
			CreatedAt: now,
			UpdatedAt: now,
		}

		if err := techcrunch.Validate(workspace.Channels, workspace.ChannelGroups); err != nil {
			return nil, eris.Wrap(err, "WorkspaceCreateOrResetDemo")
		}

		workspace.Channels = append(workspace.Channels, techcrunch)

		engadget := &Channel{
			ID:   "engadget",
			Name: "Engadget.com",
			Origins: []*ChannelOrigin{
				{
					ID:            "engadget.com / referral",
					MatchOperator: OriginMatchOperatorEqual,
					UTMSource:     "engadget.com",
					UTMMedium:     "referral",
				},
			},
			VoucherCodes: []*VoucherCode{},
			GroupID:      "referral",
			CreatedAt:    now,
			UpdatedAt:    now,
		}

		if err := engadget.Validate(workspace.Channels, workspace.ChannelGroups); err != nil {
			return nil, eris.Wrap(err, "WorkspaceCreateOrResetDemo")
		}

		workspace.Channels = append(workspace.Channels, engadget)

		retailMeNot := &Channel{
			ID:   "retailmenot",
			Name: "Retailmenot.com",
			Origins: []*ChannelOrigin{
				{
					ID:            "retailmenot.com / referral",
					MatchOperator: OriginMatchOperatorEqual,
					UTMSource:     "retailmenot.com",
					UTMMedium:     "referral",
				},
			},
			VoucherCodes: []*VoucherCode{},
			GroupID:      "coupon",
			CreatedAt:    now,
			UpdatedAt:    now,
		}
		if err := retailMeNot.Validate(workspace.Channels, workspace.ChannelGroups); err != nil {
			return nil, eris.Wrap(err, "WorkspaceCreateOrResetDemo")
		}

		workspace.Channels = append(workspace.Channels, retailMeNot)

		adroll := &Channel{
			ID:   "adroll",
			Name: "Adroll.com",
			Origins: []*ChannelOrigin{
				{
					ID:            "adroll.com / referral",
					MatchOperator: OriginMatchOperatorEqual,
					UTMSource:     "adroll.com",
					UTMMedium:     "referral",
				},
			},
			VoucherCodes: []*VoucherCode{},
			GroupID:      "retargeting",
			CreatedAt:    now,
			UpdatedAt:    now,
		}
		if err := adroll.Validate(workspace.Channels, workspace.ChannelGroups); err != nil {
			return nil, eris.Wrap(err, "WorkspaceCreateOrResetDemo")
		}

		workspace.Channels = append(workspace.Channels, adroll)

		// domains
		appleWeb := &Domain{
			ID:   "apple-web",
			Type: DomainWeb,
			Name: "Apple web",
			Hosts: []*DomainHost{
				{Host: "www.apple.com"},
				{Host: "support.apple.com", PathPrefix: StringPtr("support~")},
				{Host: "appleid.apple.com", PathPrefix: StringPtr("id~")},
				{Host: "icloud.com", PathPrefix: StringPtr("icloud~")},
			},
			ParamsWhitelist: []string{"q", "page"},
			CreatedAt:       now,
			UpdatedAt:       now,
		}
		if err := appleWeb.Validate(); err != nil {
			return nil, eris.Wrap(err, "WorkspaceCreateOrResetDemo")
		}

		workspace.Domains = append(workspace.Domains, appleWeb)

		ios := &Domain{
			ID:        "store-ios",
			Type:      DomainApp,
			Name:      "Store iOS",
			CreatedAt: now,
			UpdatedAt: now,
		}
		if err := ios.Validate(); err != nil {
			return nil, eris.Wrap(err, "WorkspaceCreateOrResetDemo")
		}

		workspace.Domains = append(workspace.Domains, ios)

		android := &Domain{
			ID:        "store-android",
			Type:      DomainApp,
			Name:      "Store Android",
			CreatedAt: now,
			UpdatedAt: now,
		}
		if err := android.Validate(); err != nil {
			return nil, eris.Wrap(err, "WorkspaceCreateOrResetDemo")
		}

		workspace.Domains = append(workspace.Domains, android)
	}

	// Generate a secret key
	key := common.RandomString(32)

	// hardcode testing secret key
	if workspace.ID == "testing" {
		key = WorkspaceTestingSecretKey
	}

	encryptedKey, err := common.EncryptString(key, cfgSecretKey)

	if err != nil {
		return nil, eris.Wrap(err, "WorkspaceCreateOrResetDemo")
	}

	secretKey := &SecretKey{
		Key:          key,
		EncryptedKey: encryptedKey,
		CreatedAt:    time.Now(),
	}

	workspace.SecretKeys = []*SecretKey{secretKey}

	if err := workspace.Validate(); err != nil {
		return nil, eris.Wrap(err, "WorkspaceCreateOrResetDemo")
	}

	return workspace, nil
}

// brand_keywords JSON NOT NULL,
// brand_keywords_as_direct BOOLEAN NOT NULL,

var WorkspaceSchema string = `CREATE ROWSTORE TABLE IF NOT EXISTS workspace (
	id VARCHAR(128) NOT NULL,
	name VARCHAR(50) NOT NULL,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
	deleted_at DATETIME,	
	is_demo BOOLEAN DEFAULT FALSE,
	demo_kind VARCHAR(20),
	website_url VARCHAR(128) NOT NULL,
	privacy_policy_url VARCHAR(128) NOT NULL,
	industry VARCHAR(64) NOT NULL,
	currency VARCHAR(3) NOT NULL,
	organization_id VARCHAR(64) NOT NULL,
	secret_keys JSON NOT NULL,
	default_user_timezone VARCHAR(64) NOT NULL,
	default_user_country VARCHAR(3) NOT NULL,
	default_user_language VARCHAR(3) NOT NULL,
	user_reconciliation_keys JSON NOT NULL,
	user_id_signing VARCHAR(20) NOT NULL,
	session_timeout SMALLINT UNSIGNED NOT NULL,
    abandoned_carts_processed_until DATETIME,
	domains JSON NOT NULL,
	has_orders BOOLEAN DEFAULT FALSE,
	has_leads BOOLEAN DEFAULT FALSE,
	lead_stages JSON NOT NULL,
	channels JSON NOT NULL,
	channel_groups JSON NOT NULL,
	installed_apps JSON NOT NULL,
	outdated_conversions_attribution BOOLEAN NOT NULL DEFAULT FALSE,
	fx_rates JSON,
	data_hooks JSON,
	license_key VARCHAR(1024),
	
	PRIMARY KEY (id),
    SHARD KEY (id)
  ) COLLATE utf8mb4_general_ci;
`
var WorkspaceSchemaMYSQL string = `CREATE TABLE IF NOT EXISTS workspace (
	id VARCHAR(128) NOT NULL,
	name VARCHAR(50) NOT NULL,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
	deleted_at DATETIME,	
	is_demo BOOLEAN DEFAULT FALSE,
	demo_kind VARCHAR(20),
	website_url VARCHAR(128) NOT NULL,
	privacy_policy_url VARCHAR(128) NOT NULL,
	industry VARCHAR(64) NOT NULL,
	currency VARCHAR(3) NOT NULL,
	organization_id VARCHAR(64) NOT NULL,
	secret_keys JSON NOT NULL,
	default_user_timezone VARCHAR(64) NOT NULL,
	default_user_country VARCHAR(3) NOT NULL,
	default_user_language VARCHAR(3) NOT NULL,
	user_reconciliation_keys JSON NOT NULL,
	user_id_signing VARCHAR(20) NOT NULL,
	session_timeout SMALLINT UNSIGNED NOT NULL,
    abandoned_carts_processed_until DATETIME,
	domains JSON NOT NULL,
	has_orders BOOLEAN DEFAULT FALSE,
	has_leads BOOLEAN DEFAULT FALSE,
	lead_stages JSON NOT NULL,
	channels JSON NOT NULL,
	channel_groups JSON NOT NULL,
	installed_apps JSON NOT NULL,
	outdated_conversions_attribution BOOLEAN NOT NULL DEFAULT FALSE,
	fx_rates JSON,
	data_hooks JSON,
	license_key VARCHAR(1024),
	
	PRIMARY KEY (id)
    -- SHARD KEY (id)
  ) COLLATE utf8mb4_general_ci;
`

var WorkspaceIndustries = []string{
	"arts-entertainment",
	"automotive",
	"beauty-fitness",
	"books-literature",
	"business-industrial-markets",
	"computer-electronics",
	"finance",
	"food-drink",
	"games",
	"healthcare",
	"hobbies-leisure",
	"home-garden",
	"internet-telecom",
	"jobs-education",
	"law-government",
	"news",
	"online-communities",
	"people-society",
	"pets-animals",
	"real-estate",
	"science",
	"shopping",
	"sports",
	"travel",
	"other",
}
