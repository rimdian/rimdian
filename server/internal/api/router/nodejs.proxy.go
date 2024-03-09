package api

// import (
// 	"crypto/tls"
// 	"io"
// 	"log"
// 	"net"
// 	"net/http"
// 	"net/url"
// 	"strings"
// 	"time"

// 	"github.com/rimdian/rimdian/internal/api/entity"
// 	"go.opencensus.io/plugin/ochttp"
// )

// // NodeProxy
// func (api *API) NodeProxy(w http.ResponseWriter, r *http.Request) {

// 	ctx := r.Context()
// 	ochttp.SetRoute(ctx, "/api/node.proxy")

// 	netTransport := &http.Transport{
// 		Dial: (&net.Dialer{
// 			Timeout: 10 * time.Second,
// 		}).Dial,
// 		TLSHandshakeTimeout: 10 * time.Second,
// 	}

// 	if api.Config.ENV == entity.ENV_DEV {
// 		netTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
// 	}

// 	netClient := &http.Client{
// 		Timeout:   time.Second * 60,
// 		Transport: netTransport,
// 	}

// 	// ReturnJSON(w, http.StatusOK, result)
// 	url, err := url.Parse(api.Config.NODEJS_ENDPOINT)

// 	if err != nil {
// 		api.ReturnJSONError(w, http.StatusInternalServerError, err)
// 		return
// 	}

// 	// strip the api/node.proxy prefix
// 	url.Path = strings.ReplaceAll(r.URL.Path, "/api/node.proxy", "")

// 	url.RawQuery = r.URL.RawQuery

// 	// log.Printf("NodeProxy token: %s", token)
// 	log.Printf("NodeProxy: %s", url.String())

// 	// create an http request
// 	req, err := http.NewRequest(r.Method, url.String(), r.Body)

// 	if err != nil {
// 		api.ReturnJSONError(w, http.StatusInternalServerError, err)
// 		return
// 	}

// 	// clone the headers
// 	req.Header = r.Header
// 	// log.Printf("headers %+v", req.Header)

// 	// inject OpenCensus span context into the request
// 	req = req.WithContext(ctx)

// 	// make the request
// 	res, err := netClient.Do(req)

// 	if err != nil {
// 		api.ReturnJSONError(w, http.StatusInternalServerError, err)
// 		return
// 	}

// 	// copy the response
// 	for k, v := range res.Header {
// 		w.Header().Set(k, v[0])
// 	}

// 	w.WriteHeader(res.StatusCode)

// 	// copy the response body
// 	_, err = io.Copy(w, res.Body)

// 	if err != nil {
// 		api.ReturnJSONError(w, http.StatusInternalServerError, err)
// 		return
// 	}

// 	// close the response body
// 	err = res.Body.Close()

// 	if err != nil {
// 		api.ReturnJSONError(w, http.StatusInternalServerError, err)
// 		return
// 	}
// }
