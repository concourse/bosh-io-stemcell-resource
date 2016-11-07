package boshio

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"
)

func NewHTTPClient(host string, wait time.Duration) HTTPClient {
	return HTTPClient{
		Host: host,
		Wait: wait,
		Client: &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,

				Dial: (&net.Dialer{
					Timeout:   30 * time.Second,
					KeepAlive: 0, // don't send keepalive TCP messages
				}).Dial,

				TLSHandshakeTimeout: 60 * time.Second,
				DisableKeepAlives:   true,
			},
		},
	}
}

type HTTPClient struct {
	Host   string
	Wait   time.Duration
	Client *http.Client
}

func (h HTTPClient) Do(req *http.Request) (*http.Response, error) {
	root, err := url.Parse(h.Host)
	if err != nil {
		return &http.Response{}, fmt.Errorf("failed to parse URL: %s", err)
	}

	if req.URL.Host == "" {
		req.URL.Host = root.Host
		req.URL.Scheme = root.Scheme
	}

	var resp *http.Response

	for {
		resp, err = h.Client.Do(req)
		if netErr, ok := err.(net.Error); ok {
			if netErr.Temporary() {
				time.Sleep(h.Wait)
				continue
			}
		}
		break
	}

	return resp, err
}
