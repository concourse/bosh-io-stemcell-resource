package boshio_test

import (
	"github.com/concourse/bosh-io-stemcell-resource/boshio"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Stemcells", func() {
	Describe("Details", func() {
		It("returns regular stemcell metadata", func() {
			stemcell := boshio.Stemcell{
				Regular: &boshio.Metadata{
					URL:  "fake-url",
					Size: 2000,
					MD5:  "fake-md5",
					SHA1: "fake-sha1",
				},
			}
			metadata := stemcell.Details()
			Expect(metadata).To(Equal(boshio.Metadata{
				URL:  "fake-url",
				Size: 2000,
				MD5:  "fake-md5",
				SHA1: "fake-sha1",
			}))
		})

		It("returns regular stemcell metadata with sha256", func() {
			stemcell := boshio.Stemcell{
				Regular: &boshio.Metadata{
					URL:    "fake-url",
					Size:   2000,
					MD5:    "fake-md5",
					SHA1:   "fake-sha1",
					SHA256: "fake-sha256",
				},
			}
			metadata := stemcell.Details()
			Expect(metadata).To(Equal(boshio.Metadata{
				URL:    "fake-url",
				Size:   2000,
				MD5:    "fake-md5",
				SHA1:   "fake-sha1",
				SHA256: "fake-sha256",
			}))
		})

		It("returns light stemcell metadata", func() {
			stemcell := boshio.Stemcell{
				Light: &boshio.Metadata{
					URL:  "fake-url",
					Size: 2000,
					MD5:  "fake-md5",
					SHA1: "fake-sha1",
				},
			}
			metadata := stemcell.Details()
			Expect(metadata).To(Equal(boshio.Metadata{
				URL:  "fake-url",
				Size: 2000,
				MD5:  "fake-md5",
				SHA1: "fake-sha1",
			}))
		})

		It("returns light stemcell metadata if both types are available", func() {
			stemcell := boshio.Stemcell{
				Light: &boshio.Metadata{
					URL:  "fake-url-light",
					Size: 2000,
					MD5:  "fake-md5-light",
					SHA1: "fake-sha1-light",
				},
				Regular: &boshio.Metadata{
					URL:  "fake-url-regular",
					Size: 2001,
					MD5:  "fake-md5-regular",
					SHA1: "fake-sha1-regular",
				},
			}
			metadata := stemcell.Details()
			Expect(metadata).To(Equal(boshio.Metadata{
				URL:  "fake-url-light",
				Size: 2000,
				MD5:  "fake-md5-light",
				SHA1: "fake-sha1-light",
			}))
		})

		It("returns regular stemcell metadata if both types are available and force_regular is true", func() {
			stemcell := boshio.Stemcell{
				Light: &boshio.Metadata{
					URL:  "fake-url-light",
					Size: 2000,
					MD5:  "fake-md5-light",
					SHA1: "fake-sha1-light",
				},
				Regular: &boshio.Metadata{
					URL:  "fake-url-regular",
					Size: 2001,
					MD5:  "fake-md5-regular",
					SHA1: "fake-sha1-regular",
				},
				ForceRegular: true,
			}
			metadata := stemcell.Details()
			Expect(metadata).To(Equal(boshio.Metadata{
				URL:  "fake-url-regular",
				Size: 2001,
				MD5:  "fake-md5-regular",
				SHA1: "fake-sha1-regular",
			}))
		})
	})

	Describe("FindStemcellByVersion", func() {
		var stemcellList boshio.Stemcells

		BeforeEach(func() {
			stemcellList = boshio.Stemcells{
				{
					Name:    "some-stemcell",
					Version: "111.1",
				},
				{
					Name:    "some-other-stemcell",
					Version: "111.2",
				},
				{
					Name:    "some-other-stemcell",
					Version: "2222",
				},
			}
		})

		It("returns the stemcell matching the criteria from the list", func() {
			stemcell, ok := stemcellList.FindStemcellByVersion("111.2")
			Expect(ok).To(BeTrue(), "Expected filter to find stemcell matching version 111.2, but it did not")
			Expect(stemcell).To(Equal(boshio.Stemcell{
				Name:    "some-other-stemcell",
				Version: "111.2",
			}))
		})

		Context("when the stemcell is not in the list", func() {
			It("returns a zero-value", func() {
				_, ok := stemcellList.FindStemcellByVersion("999.9")
				Expect(ok).To(BeFalse(), "Expected filter not to find stemcell matching version 111.2, but it did")
			})
		})
	})

	Describe("FilterByType", func() {
		var stemcellList boshio.Stemcells

		Context("when only regular stemcells are available", func() {
			BeforeEach(func() {
				stemcellList = boshio.Stemcells{
					{
						Name:    "some-stemcell",
						Regular: &boshio.Metadata{},
					},
					{
						Name:    "some-other-stemcell",
						Regular: &boshio.Metadata{},
					},
					{
						Name:    "some-other-stemcell",
						Regular: &boshio.Metadata{},
					},
				}
			})

			It("returns all regular stemcells", func() {
				filteredStemcells := stemcellList.FilterByType()
				Expect(filteredStemcells).To(Equal(stemcellList))
			})
		})

		Context("when both regular and light stemcells are available", func() {
			BeforeEach(func() {
				stemcellList = boshio.Stemcells{
					{
						Name:    "some-stemcell",
						Regular: &boshio.Metadata{},
					},
					{
						Name:  "some-light-stemcell",
						Light: &boshio.Metadata{},
					},
					{
						Name:    "some-dual-stemcell",
						Regular: &boshio.Metadata{},
						Light:   &boshio.Metadata{},
					},
				}
			})

			It("returns only the light stemcells", func() {
				filteredStemcells := stemcellList.FilterByType()
				expectedStemcells := boshio.Stemcells{
					{
						Name:  "some-light-stemcell",
						Light: &boshio.Metadata{},
					},
					{
						Name:    "some-dual-stemcell",
						Light:   &boshio.Metadata{},
						Regular: &boshio.Metadata{},
					},
				}
				Expect(filteredStemcells).To(Equal(expectedStemcells))
			})

			Context("when force_regular is true", func() {
				BeforeEach(func() {
					for i := range stemcellList {
						stemcellList[i].ForceRegular = true
					}
				})

				It("returns only the regular stemcells when force_regular is true", func() {
					filteredStemcells := stemcellList.FilterByType()
					expectedStemcells := boshio.Stemcells{
						{
							Name:         "some-stemcell",
							Regular:      &boshio.Metadata{},
							ForceRegular: true,
						},
						{
							Name:         "some-dual-stemcell",
							Light:        &boshio.Metadata{},
							Regular:      &boshio.Metadata{},
							ForceRegular: true,
						},
					}
					Expect(filteredStemcells).To(Equal(expectedStemcells))
				})
			})
		})
	})
})
