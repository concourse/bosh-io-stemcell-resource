package boshio_test

import (
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/concourse/bosh-io-stemcell-resource/boshio"
	"github.com/concourse/bosh-io-stemcell-resource/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Boshio", func() {
	var (
		client *boshio.Client
		ranger *fakes.Ranger
	)

	BeforeEach(func() {
		ranger = &fakes.Ranger{}
		bar := &fakes.Bar{}
		client = boshio.NewClient(bar, ranger)
		client.Host = boshioServer.URL()
	})

	Describe("GetStemcells", func() {
		It("fetches all stemcells for a given name", func() {
			boshioServer.Start()
			stemcells, err := client.GetStemcells("some-light-stemcell")
			Expect(err).NotTo(HaveOccurred())

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
				It("returns an error", func() {
					client.Host = "%%%%"
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

	Describe("Details", func() {
		It("returns stemcell metadata", func() {
			boshioServer.Start()
			stemcells, err := client.GetStemcells("some-heavy-stemcell")
			Expect(err).NotTo(HaveOccurred())

			metadata := stemcells[0].Details()
			Expect(metadata).To(Equal(boshio.Metadata{
				URL:  "http://example.com/heavy",
				Size: 2000,
				MD5:  "zzzz",
				SHA1: "asdf",
			}))
		})
	})

	Describe("FilterStemcells", func() {
		var stemcellList []boshio.Stemcell

		BeforeEach(func() {
			stemcellList = []boshio.Stemcell{
				{
					Name:    "some-stemcell",
					Version: "111.1",
				},
				{
					Name:    "some-other-stemcell",
					Version: "2222",
				},
			}
		})

		It("returns exactly one stemcell from the list", func() {
			stemcell, err := client.FilterStemcells("111.1", stemcellList)
			Expect(err).NotTo(HaveOccurred())

			Expect(stemcell).To(Equal(boshio.Stemcell{Name: "some-stemcell", Version: "111.1"}))
		})

		Context("when an error occurs", func() {
			Context("when the stemcell is not in the list", func() {
				It("returns an error", func() {
					_, err := client.FilterStemcells("aaaaa", stemcellList)
					Expect(err).To(MatchError(`failed to find stemcell matching version: "aaaaa"`))
				})
			})
		})
	})

	Describe("WriteMetadata", func() {
		var fileLocation *os.File

		BeforeEach(func() {
			var err error
			fileLocation, err = ioutil.TempFile("", "")
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			err := os.Remove(fileLocation.Name())
			Expect(err).NotTo(HaveOccurred())
		})

		It("writes the url to disk", func() {
			err := client.WriteMetadata(boshio.Stemcell{Light: &boshio.Metadata{URL: "http://example.com"}}, "url", fileLocation)
			Expect(err).NotTo(HaveOccurred())

			url, err := ioutil.ReadFile(fileLocation.Name())
			Expect(err).NotTo(HaveOccurred())
			Expect(string(url)).To(Equal("http://example.com"))
		})

		It("writes the sha1 to disk", func() {
			err := client.WriteMetadata(boshio.Stemcell{Regular: &boshio.Metadata{SHA1: "2222"}}, "sha1", fileLocation)
			Expect(err).NotTo(HaveOccurred())

			sha1, err := ioutil.ReadFile(fileLocation.Name())
			Expect(err).NotTo(HaveOccurred())
			Expect(string(sha1)).To(Equal("2222"))
		})

		It("writes the version to disk", func() {
			err := client.WriteMetadata(boshio.Stemcell{Version: "some version", Regular: &boshio.Metadata{}}, "version", fileLocation)
			Expect(err).NotTo(HaveOccurred())

			version, err := ioutil.ReadFile(fileLocation.Name())
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
					URL:  "http://example.com",
					Size: 100,
					MD5:  "qqqq",
					SHA1: "5f8d38fd6bb6fd12fcaa284c7132b64cbb20ea4e",
				},
			}
		})

		It("writes the stemcell to the provided location", func() {
			boshioServer.Start()
			location, err := ioutil.TempDir("", "")
			Expect(err).NotTo(HaveOccurred())

			err = client.DownloadStemcell(stubStemcell, location, false)
			Expect(err).NotTo(HaveOccurred())

			content, err := ioutil.ReadFile(filepath.Join(location, "stemcell.tgz"))
			Expect(err).NotTo(HaveOccurred())

			Expect(string(content)).To(Equal("this string is definitely not long enough to be 100 bytes but we get it there with a little bit of.."))
		})

		It("uses the stemcell filename from bosh.io when the preserveFileName param is set to true", func() {
			boshioServer.Start()
			location, err := ioutil.TempDir("", "")
			Expect(err).NotTo(HaveOccurred())

			err = client.DownloadStemcell(stubStemcell, location, true)
			Expect(err).NotTo(HaveOccurred())

			content, err := ioutil.ReadFile(filepath.Join(location, "light-different-stemcell.tgz"))
			Expect(err).NotTo(HaveOccurred())

			Expect(string(content)).To(Equal("this string is definitely not long enough to be 100 bytes but we get it there with a little bit of.."))
		})
	})

	Context("when an error occurs", func() {
		var stubStemcell boshio.Stemcell
		BeforeEach(func() {
			stubStemcell = boshio.Stemcell{
				Name:    "different-stemcell",
				Version: "2222",
				Regular: &boshio.Metadata{
					URL:  "http://example.com",
					Size: 100,
					MD5:  "qqqq",
					SHA1: "2222",
				},
			}
		})

		Context("when the head request is not successful", func() {
			It("returns an error", func() {
				client.Host = "%%%%"
				err := client.DownloadStemcell(stubStemcell, "", true)
				Expect(err).To(MatchError(ContainSubstring("invalid URL escape")))
			})
		})

		Context("when the range cannot be constructed", func() {
			It("returns an error", func() {
				ranger.BuildRangeReturns([]string{}, errors.New("failed to build a range"))
				boshioServer.Start()

				err := client.DownloadStemcell(stubStemcell, "", true)
				Expect(err).To(MatchError("failed to build a range"))
			})
		})

		Context("when the stemcell file cannot be created", func() {
			It("returns an error", func() {
				boshioServer.Start()
				location, err := ioutil.TempFile("", "")
				Expect(err).NotTo(HaveOccurred())

				defer os.RemoveAll(location.Name())

				err = location.Close()
				Expect(err).NotTo(HaveOccurred())

				err = client.DownloadStemcell(stubStemcell, location.Name(), true)
				Expect(err).To(MatchError(ContainSubstring("not a directory")))
			})
		})

		Context("when the sha1 cannot be verified", func() {
			It("returns an error", func() {
				boshioServer.Start()
				location, err := ioutil.TempDir("", "")
				Expect(err).NotTo(HaveOccurred())

				err = client.DownloadStemcell(stubStemcell, location, true)
				Expect(err).To(MatchError("computed sha1 da39a3ee5e6b4b0d3255bfef95601890afd80709 did not match expected sha1 of 2222"))
			})
		})

		Context("when the get request is not successful", func() {
			It("returns an error", func() {
				ranger.BuildRangeReturns([]string{"0-9"}, nil)
				boshioServer.TarballHandler = func(w http.ResponseWriter, req *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
				}

				boshioServer.Start()
				location, err := ioutil.TempDir("", "")
				Expect(err).NotTo(HaveOccurred())

				err = client.DownloadStemcell(stubStemcell, location, true)
				Expect(err).To(MatchError(ContainSubstring("failed to download stemcell - boshio returned 500")))
			})
		})
	})
})
