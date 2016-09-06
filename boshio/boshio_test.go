package boshio_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/concourse/bosh-io-stemcell-resource/boshio"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Boshio", func() {
	var (
		client *boshio.Client
		server *httptest.Server
	)

	BeforeEach(func() {
		server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			switch req.URL.Path {
			case "/api/v1/stemcells/some-light-stemcell":
				w.Write([]byte(`[{
					"name": "a stemcell",
					"version": "some version",
					"light": {
						"url": "http://example.com",
						"size": 100,
						"md5": "qqqq",
						"sha1": "2222"
					}
				}]`))
			case "/api/v1/stemcells/some-heavy-stemcell":
				w.Write([]byte(`[{
					"regular": {
						"url": "http://example.com/heavy",
						"size": 2000,
						"md5": "zzzz",
						"sha1": "asdf"
					}
				}]`))
			default:
				Fail(fmt.Sprintf("received unknown request: %s", req.URL.Path))
			}
		}))
		client = boshio.NewClient()
		client.Host = server.URL + "/"
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("GetStemcells", func() {
		It("fetches all stemcells for a given name", func() {
			stemcells := client.GetStemcells("some-light-stemcell")
			Expect(stemcells).To(Equal([]boshio.Stemcell{
				{
					Name:    "a stemcell",
					Version: "some version",
					Light: &boshio.Metadata{
						URL:  "http://example.com",
						Size: 100,
						MD5:  "qqqq",
						SHA1: "2222",
					},
				},
			}))
		})
	})

	Describe("Details", func() {
		It("retruns stemcell metadata", func() {
			stemcells := client.GetStemcells("some-heavy-stemcell")
			metadata := stemcells[0].Details()
			Expect(metadata).To(Equal(boshio.Metadata{
				URL:  "http://example.com/heavy",
				Size: 2000,
				MD5:  "zzzz",
				SHA1: "asdf",
			}))
		})
	})
})
