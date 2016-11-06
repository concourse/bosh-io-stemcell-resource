package boshio

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"
)

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
			break
		}
		break
	}

	return resp, err
}
