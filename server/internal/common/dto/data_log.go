package dto

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"aidanwoods.dev/go-paseto"
	"github.com/asaskevich/govalidator"
	"github.com/rimdian/rimdian/internal/api/common"
	"github.com/rimdian/rimdian/internal/common/auth"
	"github.com/rimdian/rimdian/internal/common/utils"
	"github.com/tidwall/gjson"
)

// Origin client = 0, token = 1, internal = 2
// Origin client has no signature & no authorization header
// Origin token has an authorization header
// Origin internal has a header signature computed from the request body
type DataLogOriginType = int

var HeaderSignature = "X-Rmd-Signature" // signature of the payload, used to authenticate internal data_logs
var HeaderOrigin = "X-Rmd-Origin"       // tell the collector what kind of internal origin we are sending (task, workflow, data_log)
var HeaderOriginID = "X-Rmd-Origin-Id"
var HeaderReplayID = "X-Rmd-Replay-Id"    // data_log.id of the replayed item
var HeaderAuthorization = "Authorization" // used to tell the collector that we are sending an API token

// DTO version of entity.DataLog that transits through the queue before being processed or retried by the API
const (
	DataLogOriginClient           DataLogOriginType = iota
	DataLogOriginToken                              // 1 = API token
	DataLogOriginInternalDataLog                    // 2 = Internal data_log item (ie: user_alias)
	DataLogOriginInternalWorkflow                   // 3 = Internal workflow
	DataLogOriginInternalTaskExec                   // 4 = Internal task_exec

	DoubleOptInPath       = "/double-opt-in"
	UnsubscribeEmailPath  = "/unsubscribe-email"
	OpenTrackingEmailPath = "/open-email"
)

type EmailTokenClaims struct {
	IssuedAt             time.Time `json:"iat"`
	WorkspaceID          string    `json:"wid"`
	Channel              string    `json:"ch"` // email...
	DataLogID            string    `json:"dlid"`
	MessageExternalID    string    `json:"mxid"`
	Email                string    `json:"email"`
	SubscriptionListID   string    `json:"lid"`
	SubscriptionListName string    `json:"lname"`
	AuthUID              *string   `json:"auth_uxid,omitempty"`
	AnonUID              *string   `json:"anon_uxid,omitempty"`
	// computed
	UserExternalID  string
	IsAuthenticated bool
}

func (x *EmailTokenClaims) Validate() error {
	if x.WorkspaceID == "" {
		return errors.New("missing workspace_id")
	}
	if x.DataLogID == "" {
		return errors.New("missing datalog_id")
	}
	if x.MessageExternalID == "" {
		return errors.New("missing message_external_id")
	}
	if x.Channel == "" {
		return errors.New("missing channel")
	}
	if x.Email == "" {
		return errors.New("missing email")
	}
	if x.SubscriptionListID == "" {
		return errors.New("missing list_id")
	}
	if x.SubscriptionListName == "" {
		return errors.New("missing list_name")
	}

	// set computed fields
	if x.AuthUID != nil {
		x.UserExternalID = *x.AuthUID
		x.IsAuthenticated = true
	} else if x.AnonUID != nil {
		x.UserExternalID = *x.AnonUID
		x.IsAuthenticated = false
	} else {
		return errors.New("missing auth or anon user ID")
	}

	return nil
}

type DataLogInQueue struct {
	ID       string         `json:"id"` // hash of the payload
	Origin   int            `json:"origin"`
	OriginID string         `json:"origin_id"`           // token_id, task_id, workflow_id, datalog_id...
	Context  DataLogContext `json:"context,omitempty"`   // context provides info regarding the hits
	Item     string         `json:"item"`                // raw data received from the collector
	IsReplay bool           `json:"is_replay,omitempty"` // used to identify if the data_log is a replay
}

func (x *DataLogInQueue) ComputeID(secretKey string) {
	x.ID = ComputeDataLogID(secretKey, x.Origin, x.Item)
}

func NewDataLogInQueueFromEmailToken(route string, token string, ip string, apiEndpoint string, secretKey string) (row *DataLogInQueue, claims *EmailTokenClaims, code int, err error) {

	if token == "" {
		return nil, nil, http.StatusUnauthorized, errors.New("missing token")
	}

	// 1. verify token
	parser := paseto.NewParser()
	parser.AddRule(paseto.ForAudience(apiEndpoint))

	key, err := paseto.V4SymmetricKeyFromBytes([]byte(secretKey))
	if err != nil {
		return nil, nil, http.StatusUnauthorized, err
	}

	pasetoToken, err := parser.ParseV4Local(key, token, nil)

	if err != nil {
		return nil, nil, http.StatusUnauthorized, err
	}

	// extract claims
	if err := json.Unmarshal(pasetoToken.ClaimsJSON(), &claims); err != nil {
		return nil, nil, http.StatusUnauthorized, err
	}

	// if authUID == nil && anonUID == nil {
	// 	return nil, http.StatusUnauthorized, errors.New("missing auth or anon user ID in token")
	// }

	if err := claims.Validate(); err != nil {
		return nil, nil, http.StatusUnauthorized, err
	}

	var item string
	now := time.Now().Format(time.RFC3339)

	switch route {
	case DoubleOptInPath:
		// status 1 = active subscription + reset eventual comment
		item = fmt.Sprintf(`{
			"kind": "subscription_list_user",
			"subscription_list_user": {
				"subscription_list_id": "%v",
				"status": 1,
				"comment": null,
				"created_at": "%v",
				"updated_at": "%v"
			},
			"user": {
				"external_id": "%v",
				"is_authenticated": %t,
				"created_at": "%v"
			}
		}`,
			claims.SubscriptionListID,
			claims.IssuedAt.Format(time.RFC3339),
			now,
			claims.UserExternalID,
			claims.IsAuthenticated,
			now,
		)

	case UnsubscribeEmailPath:
		// status 3 = unsubscribed
		item = fmt.Sprintf(`{
			"kind": "subscription_list_user",
			"subscription_list_user": {
				"subscription_list_id": "%v",
				"status": 3,
				"comment": "unsubscribed",
				"created_at": "%v",
				"updated_at": "%v"
			},
			"user": {
				"external_id": "%v",
				"is_authenticated": %t,
				"created_at": "%v"
			}
		}`,
			claims.SubscriptionListID,
			claims.IssuedAt.Format(time.RFC3339),
			now,
			claims.UserExternalID,
			claims.IsAuthenticated,
			now,
		)

	case OpenTrackingEmailPath:
		// set first_open_at to now
		item = fmt.Sprintf(`{
			"kind": "message",
			"message": {
				"external_id": "%v",
				"created_at": "%v",
				"updated_at": "%v",
				"channel": "email",
				"first_open_at": "%v"
			},
			"user": {
				"external_id": "%v",
				"is_authenticated": %t,
				"created_at": "%v"
			}
		}`,
			claims.MessageExternalID,
			claims.IssuedAt.Format(time.RFC3339),
			now,
			now,
			claims.UserExternalID,
			claims.IsAuthenticated,
			now,
		)

	default:
		return nil, nil, http.StatusUnauthorized, errors.New("invalid route")
	}

	origin := DataLogOriginInternalDataLog

	// create the row
	row = &DataLogInQueue{
		ID:       ComputeDataLogID(secretKey, origin, item),
		Origin:   origin,
		OriginID: claims.DataLogID,
		Context: DataLogContext{
			WorkspaceID:      claims.WorkspaceID,
			ReceivedAt:       time.Now(),
			IP:               ip,
			HeadersAndParams: MapOfStrings{},
		},
		Item: item,
	}

	return row, claims, 200, nil
}

func NewDataLogInQueueFromRequest(r *http.Request, secretKey string) (rows []*DataLogInQueue, code int, err error) {

	receivedAt := time.Now()

	// Read body
	body, err := io.ReadAll(r.Body)

	if err != nil {
		log.Println(err)
		return nil, http.StatusUnprocessableEntity, errors.New("cannot read body")
	}

	bodyString := strings.TrimSpace(string(body))

	if bodyString == "" {
		return nil, http.StatusBadRequest, errors.New("empty body")
	}

	// workspace_id might be passed as a query param (for webhooks) or in the body (web hits / API import)
	workspaceID := r.FormValue("workspace_id")

	// parse body to get workspace id
	if workspaceID == "" {
		result := gjson.Get(bodyString, "workspace_id")
		if !result.Exists() {
			return nil, http.StatusBadRequest, errors.New("workspace_id is required")
		}
		workspaceID = strings.TrimSpace(result.String())
	}

	if workspaceID == "" {
		return nil, http.StatusBadRequest, errors.New("workspace_id is required")
	}

	// some webhooks pass data data in headers (shopify, sparkpost...) that we need to extract
	// we strip out generic headers here but keep Host, IP etc...
	headersAndParams := MapOfStrings{}

	rows = []*DataLogInQueue{}
	isReplay := false
	replayID := ""
	origin := DataLogOriginClient
	originID := ""
	var tokenClaims *auth.AccountTokenClaims
	signature := ""
	verifySignature := false

	for k, headerValue := range r.Header {

		if strings.EqualFold(k, HeaderSignature) {
			signature = strings.TrimSpace(headerValue[0])
			if signature == "" {
				return nil, http.StatusBadRequest, errors.New("signature is empty")
			}
		}

		if strings.EqualFold(k, HeaderOrigin) {

			// always verify signature for internal data_logs
			verifySignature = true

			// convert header to int
			intOrigin, err := strconv.Atoi(headerValue[0])
			if err != nil {
				return nil, http.StatusBadRequest, fmt.Errorf("invalid origin: %v", err)
			}

			switch intOrigin {
			// header origin is also used for replays, to specify the source origin of the replayed item
			case DataLogOriginClient:
				origin = DataLogOriginClient
			case DataLogOriginInternalTaskExec:
				origin = DataLogOriginInternalTaskExec
			case DataLogOriginInternalWorkflow:
				origin = DataLogOriginInternalWorkflow
			case DataLogOriginInternalDataLog:
				origin = DataLogOriginInternalDataLog
			default:
				return nil, http.StatusBadRequest, fmt.Errorf("invalid origin: %v", intOrigin)
			}
		}

		if strings.EqualFold(k, HeaderOriginID) {
			originID = strings.TrimSpace(headerValue[0])
			if originID == "" {
				return nil, http.StatusBadRequest, errors.New("origin_id is empty")
			}
		}

		// token signature is already verified by the auth middleware
		if strings.EqualFold(k, HeaderAuthorization) {
			origin = DataLogOriginToken

			// Might have a valid token
			eventualAccountToken := r.Context().Value(auth.AccountTokenContextKey)

			if eventualAccountToken == nil {
				return nil, http.StatusUnauthorized, errors.New("missing token")
			}
			tokenClaims, err = auth.GetAccountTokenClaimsFromContext(r.Context())
			if err != nil {
				return nil, http.StatusUnauthorized, err
			}
		}

		// is it a replay? force lower-case, because proxies change cases...
		if strings.EqualFold(k, HeaderReplayID) {
			isReplay = true
			// extract data_log.id from the header
			replayID = strings.TrimSpace(headerValue[0])
			if replayID == "" {
				log.Printf("replay header requires an ID")
				return nil, http.StatusBadRequest, errors.New("replay header requires an ID")
			}
		}

		// keep only interesting headers (X-Geo-Country...)
		// they might contain useful data from 3rd party webhooks (shopify...)
		if !govalidator.IsIn(k,
			"Authorization",
			"Accept-Encoding",
			"Accept",
			"Accept-Language",
			"Connection",
			"Content-Length",
			"Cache-Control",
			"Cookies",
			"Cookie",
			"Dnt",
			"Forwarded",
			"Sec-Ch-Ua",
			"Sec-Ch-Ua-Mobile",
			"Sec-Ch-Ua-Platform",
			"Sec-Fetch-Dest",
			"Sec-Fetch-Mode",
			"Sec-Fetch-Site",
			"Traceparent",
			"Baggage",
			"Sentry-Trace",
			"Via",
			"Priority",
			"Alt-Used",
			HeaderOrigin,
			HeaderOriginID,
			HeaderReplayID,
			HeaderSignature,
			"X-Cloud-Trace-Context",
			"X-Forwarded-For",
			"Real-Ip",
			"X-Forwarded-Proto",
			"User-Agent",
			"Pragma",
			"Content-Type") {
			headersAndParams[k] = strings.Join(headerValue, ",")
		}
	}

	if verifySignature {
		// verify internal HMAC256 signature
		h := hmac.New(sha256.New, []byte(secretKey))
		h.Write(body)
		signatureComputed := fmt.Sprintf("%x", h.Sum(nil))

		// reject if signatures dont match
		if signatureComputed != signature {
			log.Printf("invalid signature: %v != %v", signatureComputed, signature)
			return nil, http.StatusUnauthorized, errors.New("invalid signature for internal/replay origin")
		}
	}

	// extract items from the JSON body
	// log.Printf("body: %v", bodyString)

	result := gjson.Get(bodyString, "items")
	if !result.Exists() {
		return nil, http.StatusBadRequest, errors.New("items is required")
	}

	items := result.Array()
	if len(items) == 0 {
		return nil, http.StatusBadRequest, errors.New("items is empty")
	}

	if isReplay {

		// has one item
		if len(items) != 1 {
			return nil, http.StatusBadRequest, fmt.Errorf("replay header requires one item, got %v", len(items))
		}

		if replayID == "" {
			log.Printf("replay header requires an ID")
			return nil, http.StatusBadRequest, errors.New("replay header requires an ID")
		}

		rows = append(rows, &DataLogInQueue{
			ID:     replayID,
			Origin: origin,
			Context: DataLogContext{
				WorkspaceID:      workspaceID,
				IP:               r.RemoteAddr,
				ReceivedAt:       receivedAt,
				HeadersAndParams: headersAndParams,
			},
			Item:     items[0].Raw,
			IsReplay: true,
		})

		return rows, 200, nil
	}

	realIP := utils.GetIPAdress(r)

	for _, item := range items {

		row := &DataLogInQueue{
			ID:     ComputeDataLogID(secretKey, origin, item.Raw),
			Origin: origin,
			Context: DataLogContext{
				WorkspaceID:      workspaceID,
				IP:               realIP,
				ReceivedAt:       receivedAt,
				HeadersAndParams: headersAndParams,
			},
			Item: item.Raw,
		}

		switch origin {
		case DataLogOriginInternalTaskExec, DataLogOriginInternalWorkflow, DataLogOriginInternalDataLog:
			if originID == "" {
				return nil, http.StatusBadRequest, errors.New("origin_id is required")
			}
			row.OriginID = originID

		case DataLogOriginClient:
			row.OriginID = r.Host

			// extract data_sent_at from the JSON body
			result := gjson.Get(bodyString, "context.data_sent_at")
			if !result.Exists() {
				return nil, http.StatusBadRequest, errors.New("context.data_sent_at is required")
			}
			dataSentAt := strings.TrimSpace(result.String())
			if dataSentAt == "" {
				return nil, http.StatusBadRequest, errors.New("context.data_sent_at is empty")
			}
			if t, err := time.Parse(time.RFC3339Nano, dataSentAt); err == nil {
				row.Context.DataSentAt = &t
			} else {
				return nil, http.StatusBadRequest, fmt.Errorf("context.data_sent_at is invalid: %v", err)
			}

			// compute clock difference
			if row.Context.DataSentAt != nil {
				clientClock := *row.Context.DataSentAt
				collectorClock := row.Context.ReceivedAt
				row.Context.ClockDifference = collectorClock.Sub(clientClock)
			}

		case DataLogOriginToken:
			// extract the token ID from the claims
			if tokenClaims == nil {
				return nil, http.StatusBadRequest, errors.New("missing token claims")
			}
			row.OriginID = tokenClaims.AccountID
		default:
			log.Printf("origin_id not implemented: %v", origin)
		}

		// log.Printf("data_log: %v", row)

		rows = append(rows, row)
	}

	return rows, 200, nil
}

// compute a HMAC ID from its origin+item
func ComputeDataLogID(cfgSecretKey string, origin int, item string) string {
	h := hmac.New(sha256.New, []byte(cfgSecretKey))
	h.Write([]byte(fmt.Sprintf("%v_%v", origin, item)))
	return fmt.Sprintf("%x", h.Sum(nil))
}

// result of a data log import
type ResponseForTaskQueue struct {
	HasError         bool   `json:"has_error"`
	Error            string `json:"error,omitempty"`
	QueueShouldRetry bool   `json:"queue_should_retry,omitempty"`
	// QueueShouldReplay bool   `json:"queue_should_replay,omitempty"`
}

func (x *ResponseForTaskQueue) SetError(err string, shouldRetry bool) {
	x.HasError = true
	x.Error = err
	x.QueueShouldRetry = shouldRetry
	// TODO: change datalog status to "error"
}

// the context holds metadatas required by some items
// restricted fields, automatically set by CM while processing requests:
type DataLogContext struct {
	WorkspaceID      string        `json:"workspace_id"`
	IP               string        `json:"ip,omitempty"`
	ReceivedAt       time.Time     `json:"received_at"`                  // date of received batch on the collector
	HeadersAndParams MapOfStrings  `json:"headers_and_params,omitempty"` // eventual parameters & headers passed to the URL, used by webhooks (i.e: workspace_id=xxx)
	DataSentAt       *time.Time    `json:"data_sent_at,omitempty"`       // browser sending time, used to detect browser clock skew by comparing with server-side "receivedAt" time
	ClockDifference  time.Duration `json:"clock_difference,omitempty"`   // difference between client and server clocks
}

func (dataCtx *DataLogContext) Validate(workspaceSecretKey string, anonymizeIP bool) (errFieldName string, err error) {

	// country is extracted from headers during API processing
	// if dataCtx.Country != nil && !govalidator.IsIn(*dataCtx.Country, common.CountriesCodes...) {
	// 	return "context.country", eris.Errorf("country is not valid (%s)", *dataCtx.Country)
	// }

	if anonymizeIP && dataCtx.IP != "" {
		// use a fraction of signed hmac to anonymize ip
		signedIP := common.ComputeHMAC256([]byte(dataCtx.IP), workspaceSecretKey)
		// keep the 10 first characters of the signed hash
		runes := []rune(signedIP)
		dataCtx.IP = string(runes[0:10])
	}

	return "", nil
}

func (x *DataLogContext) Scan(val interface{}) error {

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

func (x DataLogContext) Value() (driver.Value, error) {
	return json.Marshal(x)
}

type MapOfStrings map[string]string

func (x *MapOfStrings) Scan(val interface{}) error {

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

func (x MapOfStrings) Value() (driver.Value, error) {
	return json.Marshal(x)
}
