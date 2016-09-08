package boshio_test

import (
	"io/ioutil"
	"os"

	"github.com/concourse/bosh-io-stemcell-resource/boshio"

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
		var (
			dataLocations map[string]*os.File
		)

		BeforeEach(func() {
			dataLocations = map[string]*os.File{
				"version": nil,
				"sha1":    nil,
				"url":     nil,
			}

			for key := range dataLocations {
				fileLocation, err := ioutil.TempFile("", "")
				Expect(err).NotTo(HaveOccurred())

				dataLocations[key] = fileLocation
			}
		})

		AfterEach(func() {
			for _, file := range dataLocations {
				err := os.Remove(file.Name())
				Expect(err).NotTo(HaveOccurred())
			}
		})

		It("writes the url, sha1, and version to disk", func() {
			err := client.WriteMetadata("some-light-stemcell", "some version", dataLocations)
			Expect(err).NotTo(HaveOccurred())

			url, err := ioutil.ReadFile(dataLocations["url"].Name())
			Expect(err).NotTo(HaveOccurred())
			Expect(string(url)).To(Equal("http://example.com"))

			sha1, err := ioutil.ReadFile(dataLocations["sha1"].Name())
			Expect(err).NotTo(HaveOccurred())
			Expect(string(sha1)).To(Equal("2222"))

			version, err := ioutil.ReadFile(dataLocations["version"].Name())
			Expect(err).NotTo(HaveOccurred())
			Expect(string(version)).To(Equal("some version"))
		})

		Context("when an error occurs", func() {
			// Context("when url writer fails", func() {
			// 	It("returns an error", func() {
			// 		dataLocations["url"] =
			// 		err := client.WriteMetadata("some-light-stemcell", "some version", dataLocations)
			// 		Expect(err).To(MatchError("explosions"))
			// 	})
			// })

			// Context("when sha1 writer fails", func() {
			// 	It("returns an error", func() {
			// 		dataLocations["sha1"] = noopWriter{}
			// 		err := client.WriteMetadata("some-light-stemcell", "some version", dataLocations)
			// 		Expect(err).To(MatchError("explosions"))
			// 	})
			// })

			// Context("when version writer fails", func() {
			// 	It("returns an error", func() {
			// 		dataLocations["version"] = noopWriter{}
			// 		err := client.WriteMetadata("some-light-stemcell", "some version", dataLocations)
			// 		Expect(err).To(MatchError("explosions"))
			// 	})
			// })
		})
	})

	Describe("DownloadStemcell", func() {
		It("writes the stemcell to the provided location", func() {
			file, err := ioutil.TempFile("", "")
			Expect(err).NotTo(HaveOccurred())

			err = client.DownloadStemcell("different-stemcell", "2222", file)
			Expect(err).NotTo(HaveOccurred())

			content, err := ioutil.ReadFile(file.Name())
			Expect(err).NotTo(HaveOccurred())

			Expect(string(content)).To(Equal("this string is definitely not long enough to be 100 bytes but we get it there with a little bit of.."))
		})
	})
})
