package boshio_test

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/concourse/bosh-io-stemcell-resource/boshio"
	"github.com/concourse/bosh-io-stemcell-resource/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Boshio", func() {
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
		It("returns stemcell metadata", func() {
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
			err := client.WriteMetadata("some-light-stemcell", "some version", "url", fileLocation)
			Expect(err).NotTo(HaveOccurred())

			url, err := ioutil.ReadFile(fileLocation.Name())
			Expect(err).NotTo(HaveOccurred())
			Expect(string(url)).To(Equal("http://example.com"))
		})

		It("writes the sha1 to disk", func() {
			err := client.WriteMetadata("some-light-stemcell", "some version", "sha1", fileLocation)
			Expect(err).NotTo(HaveOccurred())

			sha1, err := ioutil.ReadFile(fileLocation.Name())
			Expect(err).NotTo(HaveOccurred())
			Expect(string(sha1)).To(Equal("2222"))
		})

		It("writes the version to disk", func() {
			err := client.WriteMetadata("some-light-stemcell", "some version", "version", fileLocation)
			Expect(err).NotTo(HaveOccurred())

			version, err := ioutil.ReadFile(fileLocation.Name())
			Expect(err).NotTo(HaveOccurred())
			Expect(string(version)).To(Equal("some version"))
		})

		Context("when an error occurs", func() {
			Context("when the stemcell cannot be found", func() {
				It("returns an error", func() {
					err := client.WriteMetadata("some-heavy-stemcell", "some version", "url", fakes.NoopWriter{})
					Expect(err).To(MatchError(`Failed to find stemcell: "some-heavy-stemcell"`))
				})
			})

			Context("when url writer fails", func() {
				It("returns an error", func() {
					err := client.WriteMetadata("some-light-stemcell", "some version", "url", fakes.NoopWriter{})
					Expect(err).To(MatchError("explosions"))
				})
			})

			Context("when sha1 writer fails", func() {
				It("returns an error", func() {
					err := client.WriteMetadata("some-light-stemcell", "some version", "sha1", fakes.NoopWriter{})
					Expect(err).To(MatchError("explosions"))
				})
			})

			Context("when version writer fails", func() {
				It("returns an error", func() {
					err := client.WriteMetadata("some-light-stemcell", "some version", "version", fakes.NoopWriter{})
					Expect(err).To(MatchError("explosions"))
				})
			})
		})
	})

	Describe("DownloadStemcell", func() {
		BeforeEach(func() {
			ranger.BuildRangeCall.Returns.Ranges = []string{
				"0-9", "10-19", "20-29",
				"30-39", "40-49", "50-59",
				"60-69", "70-79", "80-89",
				"90-99",
			}
		})

		It("writes the stemcell to the provided location", func() {
			location, err := ioutil.TempDir("", "")
			Expect(err).NotTo(HaveOccurred())

			err = client.DownloadStemcell("different-stemcell", "2222", location, false)
			Expect(err).NotTo(HaveOccurred())

			content, err := ioutil.ReadFile(filepath.Join(location, "stemcell.tgz"))
			Expect(err).NotTo(HaveOccurred())

			Expect(string(content)).To(Equal("this string is definitely not long enough to be 100 bytes but we get it there with a little bit of.."))
		})

		It("uses the stemcell filename from bosh.io when the preserveFileName paream is set to true", func() {
			location, err := ioutil.TempDir("", "")
			Expect(err).NotTo(HaveOccurred())

			err = client.DownloadStemcell("different-stemcell", "2222", location, true)
			Expect(err).NotTo(HaveOccurred())

			content, err := ioutil.ReadFile(filepath.Join(location, "light-different-stemcell.tgz"))
			Expect(err).NotTo(HaveOccurred())

			Expect(string(content)).To(Equal("this string is definitely not long enough to be 100 bytes but we get it there with a little bit of.."))
		})
	})
})
