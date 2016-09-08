package boshio_test

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strconv"

	"github.com/concourse/bosh-io-stemcell-resource/boshio"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

type noopWriter struct{}

func (no noopWriter) Write(b []byte) (n int, err error) {
	return 0, errors.New("explosions")
}

func TestBoshio(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Boshio Suite")
}

var (
	client *boshio.Client
	server *httptest.Server
)

var _ = BeforeEach(func() {
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		switch req.URL.Path {
		case "/d/stemcells/different-stemcell":
			if req.Method == "HEAD" {
				w.Header().Add("Content-Length", "100")
				return
			}

			ex := regexp.MustCompile(`bytes=(\d+)-(\d+)`)
			matches := ex.FindStringSubmatch(req.Header.Get("Range"))

			start, err := strconv.Atoi(matches[1])
			if err != nil {
				Fail(err.Error())
			}

			end, err := strconv.Atoi(matches[2])
			if err != nil {
				Fail(err.Error())
			}

			content := []byte("this string is definitely not long enough to be 100 bytes but we get it there with a little bit of..")
			w.Write(content[start : end+1])
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

var _ = AfterEach(func() {
	server.Close()
})
