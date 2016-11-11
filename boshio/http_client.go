package boshio

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
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
					Timeout: 30 * time.Second,
					// The OS determines the number of failed keepalive probes before the connection is closed.
					// The default is 9 retries on Linux.
					KeepAlive: 30 * time.Second,
				}).Dial,

				TLSHandshakeTimeout: 60 * time.Second,
				DisableKeepAlives:   true, // don't re-use TCP connections between requests
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
				fmt.Fprintf(os.Stderr, "Retrying on temporary error: %s", netErr.Error())
				time.Sleep(h.Wait)
				continue
			}
		}
		break
	}

	return resp, err
}
