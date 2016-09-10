package boshio_test

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"strconv"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestBoshio(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Boshio Suite")
}

var (
	boshioServer *server
)

type server struct {
	RedirectHandler http.HandlerFunc
	TarballHandler  http.HandlerFunc
	LightAPIHandler http.HandlerFunc
	HeavyAPIHandler http.HandlerFunc
	mux             *http.ServeMux
	s               *httptest.Server
}

func (s *server) Start() {
	s.mux.HandleFunc("/d/stemcells/different-stemcell", boshioServer.RedirectHandler)
	s.mux.HandleFunc("/path/to/light-different-stemcell.tgz", boshioServer.TarballHandler)
	s.mux.HandleFunc("/api/v1/stemcells/some-light-stemcell", boshioServer.LightAPIHandler)
	s.mux.HandleFunc("/api/v1/stemcells/some-heavy-stemcell", boshioServer.HeavyAPIHandler)

	s.s.Start()
}

func (s *server) Stop() {
	s.s.Close()
}

func (s *server) URL() string {
	return "http://" + s.s.Listener.Addr().String() + "/"
}

var _ = BeforeEach(func() {
	router := http.NewServeMux()
	testServer := httptest.NewUnstartedServer(router)
	boshioServer = &server{
		mux:             router,
		RedirectHandler: redirectHandler,
		TarballHandler:  tarballHandler,
		LightAPIHandler: lightAPIHandler,
		HeavyAPIHandler: heavyAPIHandler,
		s:               testServer,
	}
})

var _ = AfterEach(func() {
	boshioServer.Stop()
})

func redirectHandler(w http.ResponseWriter, req *http.Request) {
	magicURL := req.URL
	magicURL.Path = "/path/to/light-different-stemcell.tgz"

	w.Header().Add("Location", magicURL.String())
	w.WriteHeader(http.StatusMovedPermanently)
}

func tarballHandler(w http.ResponseWriter, req *http.Request) {
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
}

func lightAPIHandler(w http.ResponseWriter, req *http.Request) {
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
}

func heavyAPIHandler(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte(`[{
					"regular": {
						"url": "http://example.com/heavy",
						"size": 2000,
						"md5": "zzzz",
						"sha1": "asdf"
					}
				}]`))
}
