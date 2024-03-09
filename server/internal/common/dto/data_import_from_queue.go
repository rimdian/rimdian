package dto

// var ErrDataSignatureInvalid = eris.New("data signature invalid")
// var ErrDataBodyInvalid = eris.New("data body invalid")

// var OriginClient = "client"     // client activity, with eventually userId hmac signed
// var OriginToken = "token"       // authenticated API imports
// var OriginInternal = "internal" // internally generated DB mutations

// // The DataImportFromQueue object is created by the collector, sent into the Queue and processed by the API
// type DataImportFromQueue struct {
// 	ID         string              `json:"id"`               // HMAC512 of the Body with secret key
// 	Origin     string              `json:"origin"`           // client / token. Will determine how to trust the data
// 	Headers    map[string][]string `json:"headers"`          // headers of the incoming request into the Collector, contains IP, Host, Webhooks metadatas...
// 	ClientIP   string              `json:"client_ip"`        // client IP address
// 	Body       string              `json:"body"`             // type DataImport : payload of the data sent to the collector
// 	ReceivedAt time.Time           `json:"received_at"`      // datetime of received data in the Collector
// 	Params     map[string]string   `json:"params,omitempty"` // eventual parameters passed to the URL, used by webhooks (i.e: workspace_id=xxx)
// 	IsReplay   bool                `json:"is_replay"`        // true if the data is a replay of a previous data import
// }

// // compute a HMAC ID from its Body
// func ComputeDataImportID(cfgSecretKey string, jsonBody []byte) string {
// 	h := hmac.New(sha256.New, []byte(cfgSecretKey))
// 	h.Write(jsonBody)
// 	return fmt.Sprintf("%x", h.Sum(nil))
// }

// // compute a hash ID from its Body and compares it with existing ID
// func (data *DataImportFromQueue) VerifyIntegrity(cfgSecretKey string) error {
// 	// the data payload is different in replay mode
// 	// it contains isReplay=true, therefore the signature of data.ID is different
// 	if data.IsReplay {
// 		return nil
// 	}
// 	h := hmac.New(sha256.New, []byte(cfgSecretKey))
// 	h.Write([]byte(data.Body))
// 	if fmt.Sprintf("%x", h.Sum(nil)) != data.ID {
// 		return ErrDataSignatureInvalid
// 	}
// 	return nil
// }
