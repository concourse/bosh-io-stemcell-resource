package boshio_test

import (
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/concourse/bosh-io-stemcell-resource/boshio"
	"github.com/concourse/bosh-io-stemcell-resource/fakes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type EOFReader struct{}

func (e EOFReader) Read(p []byte) (int, error) {
	return 0, io.ErrUnexpectedEOF
}

var _ = Describe("Boshio", func() {
	var (
		auth         boshio.Auth
		httpClient   boshio.HTTPClient
		client       *boshio.Client
		ranger       *fakes.Ranger
		bar          *fakes.Bar
		forceRegular bool
	)

	BeforeEach(func() {
		auth = boshio.Auth{}
		ranger = &fakes.Ranger{}
		bar = &fakes.Bar{}
		forceRegular = false
		httpClient = boshio.NewHTTPClient(boshioServer.URL(), 800*time.Millisecond)
		client = boshio.NewClient(httpClient, bar, ranger, forceRegular)
	})

	Describe("GetStemcells", func() {
		It("fetches all stemcells for a given name", func() {
			boshioServer.Start()
			stemcells, err := client.GetStemcells("some-light-stemcell")
			Expect(err).NotTo(HaveOccurred())

			Expect(stemcells).To(Equal(boshio.Stemcells{
				{
					Name:    "a stemcell",
					Version: "some version",
					Light: &boshio.Metadata{
						URL:    serverPath("path/to/light-different-stemcell.tgz"),
						Size:   100,
						MD5:    "qqqq",
						SHA1:   "2222",
						SHA256: "4444",
					},
				},
			}))
		})

		Context("when an error occurs", func() {
			Context("when bosh.io responds with a non-200", func() {
				It("returns an error", func() {
					boshioServer.LightAPIHandler = func(w http.ResponseWriter, req *http.Request) {
						w.WriteHeader(http.StatusInternalServerError)
					}

					boshioServer.Start()
					_, err := client.GetStemcells("some-light-stemcell")
					Expect(err).To(MatchError("failed fetching metadata - boshio returned: 500"))
				})
			})

			Context("when the get fails", func() {
				XIt("returns an error", func() {
					_, err := client.GetStemcells("some-light-stemcell")
					Expect(err).To(MatchError(ContainSubstring("invalid URL escape")))
				})
			})

			Context("when the response is invalid json", func() {
				It("returns an error", func() {
					boshioServer.LightAPIHandler = func(w http.ResponseWriter, req *http.Request) {
						w.Write([]byte(`%%%%%`))
					}

					boshioServer.Start()
					_, err := client.GetStemcells("some-light-stemcell")
					Expect(err).To(MatchError(ContainSubstring("invalid character")))
				})
			})
		})
	})

	Describe("WriteMetadata", func() {
		var fileLocation *os.File

		BeforeEach(func() {
			var err error
			fileLocation, err = os.CreateTemp("", "")
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			err := os.Remove(fileLocation.Name())
			Expect(err).NotTo(HaveOccurred())
		})

		It("writes the url to disk", func() {
			err := client.WriteMetadata(boshio.Stemcell{Light: &boshio.Metadata{URL: "http://example.com"}}, "url", fileLocation)
			Expect(err).NotTo(HaveOccurred())

			url, err := os.ReadFile(fileLocation.Name())
			Expect(err).NotTo(HaveOccurred())
			Expect(string(url)).To(Equal("http://example.com"))
		})

		It("writes the sha1 to disk", func() {
			err := client.WriteMetadata(boshio.Stemcell{Regular: &boshio.Metadata{SHA1: "2222"}}, "sha1", fileLocation)
			Expect(err).NotTo(HaveOccurred())

			sha1, err := os.ReadFile(fileLocation.Name())
			Expect(err).NotTo(HaveOccurred())
			Expect(string(sha1)).To(Equal("2222"))
		})

		It("writes the sha256 to disk", func() {
			err := client.WriteMetadata(boshio.Stemcell{Regular: &boshio.Metadata{SHA256: "4444"}}, "sha256", fileLocation)
			Expect(err).NotTo(HaveOccurred())

			sha256, err := os.ReadFile(fileLocation.Name())
			Expect(err).NotTo(HaveOccurred())
			Expect(string(sha256)).To(Equal("4444"))
		})

		It("writes the version to disk", func() {
			err := client.WriteMetadata(boshio.Stemcell{Version: "some version", Regular: &boshio.Metadata{}}, "version", fileLocation)
			Expect(err).NotTo(HaveOccurred())

			version, err := os.ReadFile(fileLocation.Name())
			Expect(err).NotTo(HaveOccurred())
			Expect(string(version)).To(Equal("some version"))
		})

		Context("when an error occurs", func() {
			Context("when url writer fails", func() {
				It("returns an error", func() {
					err := client.WriteMetadata(boshio.Stemcell{Name: "some-heavy-stemcell", Regular: &boshio.Metadata{}}, "url", fakes.NoopWriter{})
					Expect(err).To(MatchError("explosions"))
				})
			})

			Context("when sha1 writer fails", func() {
				It("returns an error", func() {
					err := client.WriteMetadata(boshio.Stemcell{Name: "some-heavy-stemcell", Regular: &boshio.Metadata{}}, "sha1", fakes.NoopWriter{})
					Expect(err).To(MatchError("explosions"))
				})
			})

			Context("when version writer fails", func() {
				It("returns an error", func() {
					err := client.WriteMetadata(boshio.Stemcell{Name: "some-heavy-stemcell", Regular: &boshio.Metadata{}}, "version", fakes.NoopWriter{})
					Expect(err).To(MatchError("explosions"))
				})
			})
		})
	})

	Describe("DownloadStemcell", func() {
		var stubStemcell boshio.Stemcell
		BeforeEach(func() {
			ranger.BuildRangeReturns([]string{
				"0-9", "10-19", "20-29",
				"30-39", "40-49", "50-59",
				"60-69", "70-79", "80-89",
				"90-99",
			}, nil)

			stubStemcell = boshio.Stemcell{
				Name:    "different-stemcell",
				Version: "2222",
				Regular: &boshio.Metadata{
					URL:  serverPath("path/to/light-different-stemcell.tgz"),
					Size: 100,
					MD5:  "qqqq",
					SHA1: "5f8d38fd6bb6fd12fcaa284c7132b64cbb20ea4e",
				},
			}
		})

		It("writes the stemcell to the provided location", func() {
			boshioServer.Start()
			location, err := os.MkdirTemp("", "")
			Expect(err).NotTo(HaveOccurred())

			err = client.DownloadStemcell(stubStemcell, location, false, auth)
			Expect(err).NotTo(HaveOccurred())

			content, err := os.ReadFile(filepath.Join(location, "stemcell.tgz"))
			Expect(err).NotTo(HaveOccurred())

			Expect(string(content)).To(Equal("this string is definitely not long enough to be 100 bytes but we get it there with a little bit of.."))
		})

		It("uses the stemcell filename from bosh.io when the preserveFileName param is set to true", func() {
			boshioServer.Start()
			location, err := os.MkdirTemp("", "")
			Expect(err).NotTo(HaveOccurred())

			err = client.DownloadStemcell(stubStemcell, location, true, auth)
			Expect(err).NotTo(HaveOccurred())

			content, err := os.ReadFile(filepath.Join(location, "light-different-stemcell.tgz"))
			Expect(err).NotTo(HaveOccurred())

			Expect(string(content)).To(Equal("this string is definitely not long enough to be 100 bytes but we get it there with a little bit of.."))
		})

		Context("when using auth", func() {
			BeforeEach(func() {
				auth = boshio.Auth{
					AccessKey: "access key",
					SecretKey: "secret key",
				}
			})

			It("writes the stemcell to the provided location", func() {
				stubStemcell.Regular.URL = serverPath("bucket_name/path/to/heavy-stemcell.tgz")
				boshioServer.Start()
				location, err := os.MkdirTemp("", "")
				Expect(err).NotTo(HaveOccurred())

				err = client.DownloadStemcell(stubStemcell, location, false, auth)
				Expect(err).NotTo(HaveOccurred())

				content, err := os.ReadFile(filepath.Join(location, "stemcell.tgz"))
				Expect(err).NotTo(HaveOccurred())

				Expect(string(content)).To(Equal("this string is definitely not long enough to be 100 bytes but we get it there with a little bit of.."))
			})

			Context("when the metadata cannot be fetched", func() {
				BeforeEach(func() {
					stubStemcell.Regular.URL = serverPath("bucket_name/path/to/nothing.tgz")
					boshioServer.Start()
				})

				It("returns an error", func() {
					err := client.DownloadStemcell(stubStemcell, "", false, auth)
					Expect(err).To(MatchError(ContainSubstring("failed to fetch object metadata:")))
				})
			})
		})
	})

	Context("when an error occurs", func() {
		var stubStemcell boshio.Stemcell
		BeforeEach(func() {
			stubStemcell = boshio.Stemcell{
				Name:    "different-stemcell",
				Version: "2222",
				Regular: &boshio.Metadata{
					URL:  serverPath("path/to/light-different-stemcell.tgz"),
					Size: 100,
					MD5:  "qqqq",
					SHA1: "2222",
				},
			}
		})

		Context("when an io error occurs", func() {
			It("retries the request", func() {
				httpClient := &fakes.HTTPClient{}
				ranger.BuildRangeReturns([]string{"0-9"}, nil)

				var (
					responses  []*http.Response
					httpErrors []error
				)

				httpClient.DoStub = func(req *http.Request) (*http.Response, error) {
					count := httpClient.DoCallCount() - 1
					return responses[count], httpErrors[count]
				}

				responses = []*http.Response{
					{StatusCode: http.StatusOK, Body: nil, ContentLength: 10, Request: &http.Request{URL: &url.URL{Scheme: "https", Host: "example.com", Path: "hello"}}},
					{StatusCode: http.StatusPartialContent, Body: io.NopCloser(EOFReader{})},
					{StatusCode: http.StatusPartialContent, Body: io.NopCloser(strings.NewReader("hello good"))},
				}

				httpErrors = []error{nil, nil, nil}

				client = boshio.NewClient(httpClient, bar, ranger, forceRegular)

				location, err := os.MkdirTemp("", "")
				Expect(err).NotTo(HaveOccurred())

				stubStemcell := boshio.Stemcell{
					Name:    "different-stemcell",
					Version: "2222",
					Regular: &boshio.Metadata{
						URL:  serverPath("path/to/light-different-stemcell.tgz"),
						Size: 100,
						MD5:  "qqqq",
						SHA1: "1c36c7afa4e21e2ccc0c386f790560672534723a",
					},
				}

				err = client.DownloadStemcell(stubStemcell, location, false, auth)
				Expect(err).NotTo(HaveOccurred())

				content, err := os.ReadFile(filepath.Join(location, "stemcell.tgz"))
				Expect(err).NotTo(HaveOccurred())

				Expect(string(content)).To(Equal("hello good"))
			})
		})

		Context("when the HEAD request cannot be constructed", func() {
			It("returns an error", func() {
				stubStemcell := boshio.Stemcell{
					Name:    "different-stemcell",
					Version: "2222",
					Regular: &boshio.Metadata{
						URL:  "%%%%",
						Size: 100,
						MD5:  "qqqq",
						SHA1: "1c36c7afa4e21e2ccc0c386f790560672534723a",
					},
				}

				err := client.DownloadStemcell(stubStemcell, "", false, auth)
				Expect(err).To(MatchError(ContainSubstring("failed to construct HEAD request:")))
			})
		})

		Context("when the range cannot be constructed", func() {
			It("returns an error", func() {
				ranger.BuildRangeReturns([]string{}, errors.New("failed to build a range"))
				boshioServer.Start()

				err := client.DownloadStemcell(stubStemcell, "", true, auth)
				Expect(err).To(MatchError("failed to build a range"))
			})
		})

		Context("when the stemcell file cannot be created", func() {
			It("returns an error", func() {
				boshioServer.Start()
				location, err := os.CreateTemp("", "")
				Expect(err).NotTo(HaveOccurred())

				defer os.RemoveAll(location.Name())

				err = location.Close()
				Expect(err).NotTo(HaveOccurred())

				err = client.DownloadStemcell(stubStemcell, location.Name(), true, auth)
				Expect(err).To(MatchError(ContainSubstring("not a directory")))
			})
		})

		Context("when the sha1 cannot be verified", func() {
			It("returns an error", func() {
				boshioServer.Start()
				location, err := os.MkdirTemp("", "")
				Expect(err).NotTo(HaveOccurred())

				err = client.DownloadStemcell(stubStemcell, location, true, auth)
				Expect(err).To(MatchError("computed sha1 da39a3ee5e6b4b0d3255bfef95601890afd80709 did not match expected sha1 of 2222"))
			})
		})

		Context("when the sha256 cannot be verified", func() {
			It("returns an error", func() {
				stubStemcell.Regular.SHA256 = "4444"
				boshioServer.Start()
				location, err := os.MkdirTemp("", "")
				Expect(err).NotTo(HaveOccurred())

				err = client.DownloadStemcell(stubStemcell, location, true, auth)
				Expect(err).To(MatchError("computed sha256 e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855 did not match expected sha256 of 4444"))
			})
		})

		Context("when the get request is not successful", func() {
			It("returns an error", func() {
				ranger.BuildRangeReturns([]string{"0-9"}, nil)
				boshioServer.TarballHandler = func(w http.ResponseWriter, req *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
				}

				boshioServer.Start()
				location, err := os.MkdirTemp("", "")
				Expect(err).NotTo(HaveOccurred())

				err = client.DownloadStemcell(stubStemcell, location, true, auth)
				Expect(err).To(MatchError(ContainSubstring("failed to download stemcell - boshio returned 500")))
			})
		})
	})
})
