package versions_test

import (
	"github.com/concourse/bosh-io-stemcell-resource/boshio"
	"github.com/concourse/bosh-io-stemcell-resource/versions"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Versions", func() {
	Context("when provided with no starting version", func() {
		var filter versions.Filter

		BeforeEach(func() {
			stemcells := []boshio.Stemcell{
				{Version: "3232.9"},
				{Version: "3232.8"},
				{Version: "3232.1"},
				{Version: "3232"},
			}

			filter = versions.NewFilter("", stemcells)
		})

		It("returns the latest version", func() {
			list, err := filter.Versions()
			Expect(err).NotTo(HaveOccurred())

			Expect(list).To(Equal([]versions.List{
				{"version": "3232.9"},
			}))
		})
	})

	Context("when provided with a starting version", func() {
		var filter versions.Filter

		BeforeEach(func() {
			stemcells := []boshio.Stemcell{
				{Version: "3232.9"},
				{Version: "3232.8"},
				{Version: "3232.7.1"},
				{Version: "3232.7"},
				{Version: "3232.4"},
				{Version: "3232.3"},
				{Version: "3232.2"},
				{Version: "3232"},
			}

			filter = versions.NewFilter("3232.4", stemcells)
		})

		It("returns all the versions newer than the provided version", func() {
			list, err := filter.Versions()
			Expect(err).NotTo(HaveOccurred())

			Expect(list).To(Equal([]versions.List{
				{"version": "3232.9"},
				{"version": "3232.8"},
				{"version": "3232.7.1"},
				{"version": "3232.7"},
				{"version": "3232.4"},
			}))
		})
	})
})
