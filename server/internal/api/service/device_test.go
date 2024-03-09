package service

// func TestServiceImpl_ParseUserAgent(t *testing.T) {

// 	t.Run("should call nodejs to parse a user agent", func(t *testing.T) {
// 		ua := "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:101.0) Gecko/20100101 Firefox/101.0"

// 		// netClientMock := &httpClient.HTTPClientMock{
// 		// 	DoFunc: func(req *http.Request) (*http.Response, error) {
// 		// 		// return a mock response
// 		// 		return &http.Response{
// 		// 			StatusCode: http.StatusOK,
// 		// 			Body:       ioutil.NopCloser(strings.NewReader(`{"browser":{"name":"Chromium","version":"15.0.874.106"},"device":{},"os":{"name":"Ubuntu","version":"11.10"}}`)),
// 		// 			// Body:       ioutil.NewBody(),
// 		// 		}, nil

// 		// 	},
// 		// }

// 		currentPath, _ := os.Getwd()

// 		svc := &ServiceImpl{
// 			Config: &entity.Config{},
// 			Repo:   &repository.RepositoryMock{},
// 			Mailer: nil,
// 			// NetClient: netClientMock,
// 		}

// 		_, err := svc.ParseUserAgent(ua)

// 		if err != nil {
// 			t.Errorf("should not return error, got: %v", err)
// 		}

// 		// check that the nodejs client was called
// 		// if len(netClientMock.DoCalls()) != 1 {
// 		// 	t.Errorf("should call nodejs client, got: %v", netClientMock.DoCalls())
// 		// }
// 	})
// }
