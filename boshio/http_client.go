package boshio

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"
)

type HTTPClient struct {
	host    string
	timeout time.Duration
	wait    time.Duration
	client  *http.Client
}

func NewHTTPClient(host string, timeout time.Duration) HTTPClient {
	return HTTPClient{
		host: host,
		wait: 1 * time.Second,
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

func (h HTTPClient) Do(req *http.Request) (*http.Response, error) {
	root, err := url.Parse(h.host)
	if err != nil {
		return &http.Response{}, fmt.Errorf("failed to parse URL: %s", err)
	}

	if req.URL.Host == "" {
		req.URL.Host = root.Host
		req.URL.Scheme = root.Scheme
	}

	var resp *http.Response

	for {
		resp, err = h.client.Do(req)
		if netErr, ok := err.(net.Error); ok {
			if netErr.Temporary() {
				time.Sleep(h.wait)
				continue
			}
			break
		}
		break
	}

	return resp, err
}
