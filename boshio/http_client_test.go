package boshio_test

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/concourse/bosh-io-stemcell-resource/boshio"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type tempError struct {
	error
}

func (te tempError) Temporary() bool {
	return true
}

func (te tempError) Timeout() bool {
	return false
}

type fakeTransport struct {
	count int
}

func (f *fakeTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if f.count != 1 {
		f.count++
		return nil, tempError{errors.New("boom")}
	}

	return &http.Response{StatusCode: http.StatusOK}, nil
}

var _ = Describe("HTTPClient", func() {
	Describe("Do", func() {
		It("makes an http request", func() {
			var (
				receivedRequest *http.Request
				requestBody     []byte
			)

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				receivedRequest = req

				var err error
				requestBody, err = ioutil.ReadAll(req.Body)
				Expect(err).NotTo(HaveOccurred())
			}))

			client := boshio.HTTPClient{Host: server.URL, Wait: 500 * time.Millisecond, Client: http.DefaultClient}

			request, err := http.NewRequest("POST", "/more/path", strings.NewReader(`{"test": "something"}`))
			Expect(err).NotTo(HaveOccurred())

			request.Header.Add("something", "some-value")

			response, err := client.Do(request)
			Expect(err).NotTo(HaveOccurred())

			Expect(response.StatusCode).To(Equal(http.StatusOK))

			Expect(receivedRequest.Method).To(Equal("POST"))
			Expect(receivedRequest.URL.String()).To(Equal("/more/path"))
			Expect(receivedRequest.Header.Get("something")).To(Equal("some-value"))

			Expect(requestBody).To(MatchJSON(`{"test": "something"}`))
		})

		Context("when the request already has its host sett", func() {
			It("doesn't modify the host", func() {
				stemcells := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {}))
				amazon := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					w.WriteHeader(http.StatusTeapot)
				}))

				client := boshio.HTTPClient{Host: stemcells.URL, Wait: 500 * time.Millisecond, Client: http.DefaultClient}

				request, err := http.NewRequest("POST", amazon.URL, strings.NewReader(`{"test": "something"}`))
				Expect(err).NotTo(HaveOccurred())

				response, err := client.Do(request)
				Expect(err).NotTo(HaveOccurred())

				Expect(response.StatusCode).To(Equal(http.StatusTeapot))
			})
		})

		Context("when the request has a temporary error", func() {
			It("retries the request", func() {
				client := boshio.HTTPClient{Host: "http://www.example.com", Wait: 100 * time.Millisecond, Client: &http.Client{Transport: &fakeTransport{}}}

				request, err := http.NewRequest("GET", "/different/path", nil)
				Expect(err).NotTo(HaveOccurred())

				response, err := client.Do(request)
				Expect(err).NotTo(HaveOccurred())

				Expect(response.StatusCode).To(Equal(http.StatusOK))
			})
		})

		Context("when an error occurs", func() {
			Context("when the host cannot be parsed", func() {
				It("returns an error", func() {
					client := boshio.HTTPClient{Host: "%%%%%%", Wait: 100 * time.Millisecond, Client: http.DefaultClient}

					_, err := client.Do(&http.Request{})
					Expect(err).To(MatchError(ContainSubstring("failed to parse URL")))
				})
			})
		})
	})
})
