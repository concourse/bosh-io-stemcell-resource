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

			filter = versions.NewFilter("", stemcells, "")
		})

		It("returns the latest version", func() {
			list, err := filter.Versions()
			Expect(err).NotTo(HaveOccurred())

			Expect(list).To(Equal(versions.StemcellVersions{
				{"version": "3232.9"},
			}))
		})
	})

	Context("when the versions are out of order", func() {
		var filter versions.Filter

		BeforeEach(func() {
			stemcells := []boshio.Stemcell{
				{Version: "3232"},
				{Version: "3232.8"},
				{Version: "3232.9"},
				{Version: "3232.1"},
				{Version: "3330.3"},
				{Version: "3333"},
			}

			filter = versions.NewFilter("3232.1", stemcells, "")
		})

		It("orders them perfectly", func() {
			list, err := filter.Versions()
			Expect(err).NotTo(HaveOccurred())

			Expect(list).To(Equal(versions.StemcellVersions{
				{"version": "3232.1"},
				{"version": "3232.8"},
				{"version": "3232.9"},
				{"version": "3330.3"},
				{"version": "3333"},
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

			filter = versions.NewFilter("3232.4", stemcells, "")
		})

		It("returns all the versions newer than the provided version", func() {
			list, err := filter.Versions()
			Expect(err).NotTo(HaveOccurred())

			Expect(list).To(Equal(versions.StemcellVersions{
				{"version": "3232.4"},
				{"version": "3232.7"},
				{"version": "3232.7.1"},
				{"version": "3232.8"},
				{"version": "3232.9"},
			}))
		})
	})

	Context("when provided with a version_family", func() {
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

			filter = versions.NewFilter("", stemcells, "3232.7")
		})

		It("returns the latest version within the family", func() {
			list, err := filter.Versions()
			Expect(err).NotTo(HaveOccurred())

			Expect(list).To(Equal(versions.StemcellVersions{
				{"version": "3232.7.1"},
			}))
		})

		Context("and no stemcells match that family", func() {
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

				filter = versions.NewFilter("", stemcells, "9999")
			})

			It("returns an empty version list", func() {
				list, err := filter.Versions()
				Expect(err).NotTo(HaveOccurred())

				Expect(list).To(Equal(versions.StemcellVersions{}))
			})
		})

		Context("and an initial version", func() {
			BeforeEach(func() {
				stemcells := []boshio.Stemcell{
					{Version: "3234"},
					{Version: "3233.2.1"},
					{Version: "3233.2"},
					{Version: "3233.1"},
					{Version: "3233"},
					{Version: "3232.7.1"},
					{Version: "3232.7"},
				}

				filter = versions.NewFilter("3233.2", stemcells, "3233")
			})

			It("returns all the versions within the family >= the initial version", func() {
				list, err := filter.Versions()
				Expect(err).NotTo(HaveOccurred())

				Expect(list).To(Equal(versions.StemcellVersions{
					{"version": "3233.2"},
					{"version": "3233.2.1"},
				}))
			})
		})
	})

	Context("when passed an empty stemcell list and no initial version", func() {
		var filter versions.Filter

		BeforeEach(func() {
			stemcells := []boshio.Stemcell{}

			filter = versions.NewFilter("", stemcells, "")
		})

		It("returns an empty list", func() {
			list, err := filter.Versions()
			Expect(err).NotTo(HaveOccurred())

			Expect(list).To(BeEmpty())
		})
	})

	Context("when passed an empty stemcell list and an initial version", func() {
		var filter versions.Filter

		BeforeEach(func() {
			stemcells := []boshio.Stemcell{}

			filter = versions.NewFilter("3232.4", stemcells, "")
		})

		It("returns an empty list", func() {
			list, err := filter.Versions()
			Expect(err).NotTo(HaveOccurred())

			Expect(list).To(BeEmpty())
		})
	})
})
