package stream

import (
	"fmt"
	"net/http"
)

type HttpClient struct {
	httpClient *http.Client
}

func NewHttp() IHTTPClient {
	httpClient := &http.Client{}
	return &HttpClient{
		httpClient: httpClient,
	}
}

// Perform a GET request to the given URL with the given headers
func (h *HttpClient) Get(url string, headers map[string]string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	for k, v := range headers {
		req.Header.Add(k, v)
	}

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error performing request: %v", err)
	}

	return resp, nil
}
