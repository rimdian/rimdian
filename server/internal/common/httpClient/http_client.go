//go:generate moq -out http_client_mock.go . HTTPClient

package httpClient

import "net/http"

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type HTTPClientImpl struct {
	NetClient *http.Client
}

func NewHTTPClient(client *http.Client) HTTPClient {
	return &HTTPClientImpl{
		NetClient: client,
	}
}

func (client *HTTPClientImpl) Do(req *http.Request) (*http.Response, error) {
	return client.NetClient.Do(req)
}
