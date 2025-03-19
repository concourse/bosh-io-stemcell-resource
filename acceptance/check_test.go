package acceptance_test

import (
	"bytes"
	"os/exec"

	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"

	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const noVersionRequest = `
{
	"source": {
		"name": "bosh-aws-xen-hvm-ubuntu-trusty-go_agent"
	}
}`

const versionFamilyRequest = `
{
	"source": {
		"name": "bosh-aws-xen-hvm-ubuntu-trusty-go_agent",
		"version_family": "3262.latest"
	},
	"version": {
		"version":"3262"
	}
}`

const versionFamilyRequestLatest = `
{
	"source": {
		"name": "bosh-aws-xen-hvm-ubuntu-trusty-go_agent",
		"version_family": "latest"
	},
	"version": {
		"version":"3262"
	}
}`

const specificVersionRequest = `
{
	"source": {
		"name": "bosh-aws-xen-hvm-ubuntu-trusty-go_agent"
	},
	"version": {
		"version":"3262.4"
	}
}`

const oldVersionRequest = `
{
	"source": {
		"name": "bosh-aws-xen-hvm-ubuntu-trusty-go_agent"
	},
	"version": {
		"version":"3151"
	}
}`

const lightOnlyForceRegularRequest = `
{
	"source": {
		"name": "bosh-aws-xen-hvm-ubuntu-trusty-go_agent",
		"force_regular": true
	}
}`

const bothTypesForceRegularRequest = `
{
	"source": {
		"name": "bosh-aws-xen-hvm-ubuntu-trusty-go_agent",
		"force_regular": true
	}
}`

type stemcellVersion map[string]string

var _ = Describe("check", func() {
	Context("when no version is specified", func() {
		var command *exec.Cmd

		BeforeEach(func() {
			command = exec.Command(boshioCheck)
			command.Stdin = bytes.NewBufferString(noVersionRequest)
		})

		It("returns only the latest version", func() {
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			<-session.Exited
			Expect(session.ExitCode()).To(Equal(0))

			result := []stemcellVersion{}
			err = json.Unmarshal(session.Out.Contents(), &result)
			Expect(err).NotTo(HaveOccurred())

			Expect(result).To(HaveLen(1))
			Expect(result[0]["version"]).NotTo(BeEmpty())
			Expect(result[0]["version"]).NotTo(Equal("3262.7"))
			Expect(result[0]["version"]).NotTo(Equal("3262.5"))
		})
	})

	Context("when a version_family is specified", func() {
		var command *exec.Cmd

		Context("with `3262.latest`", func() {
			BeforeEach(func() {
				command = exec.Command(boshioCheck)
				command.Stdin = bytes.NewBufferString(versionFamilyRequest)
			})

			It("returns only versions that match the given semver", func() {
				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				<-session.Exited
				Expect(session.ExitCode()).To(Equal(0))

				result := []stemcellVersion{}
				err = json.Unmarshal(session.Out.Contents(), &result)
				Expect(err).NotTo(HaveOccurred())

				Expect(result).To(ContainElement(stemcellVersion{
					"version": "3262.4",
				}))
				Expect(result).To(ContainElement(stemcellVersion{
					"version": "3262.4.1",
				}))
				Expect(result).To(ContainElement(stemcellVersion{
					"version": "3262.5",
				}))
				Expect(result).NotTo(ContainElement(stemcellVersion{
					"version": "3263.14",
				}))
			})
		})

		Context("with `latest`", func() {
			BeforeEach(func() {
				command = exec.Command(boshioCheck)
				command.Stdin = bytes.NewBufferString(versionFamilyRequestLatest)
			})

			It("returns only versions that match the given semver", func() {
				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				<-session.Exited
				Expect(session.ExitCode()).To(Equal(0))

				result := []stemcellVersion{}
				err = json.Unmarshal(session.Out.Contents(), &result)
				Expect(err).NotTo(HaveOccurred())

				Expect(result).To(ContainElement(stemcellVersion{
					"version": "3262.4",
				}))
				Expect(result).To(ContainElement(stemcellVersion{
					"version": "3262.4.1",
				}))
				Expect(result).To(ContainElement(stemcellVersion{
					"version": "3312.3",
				}))
				Expect(result).To(ContainElement(stemcellVersion{
					"version": "3263.14",
				}))
			})
		})
	})

	Context("when a version is specified", func() {
		var command *exec.Cmd

		BeforeEach(func() {
			command = exec.Command(boshioCheck)
			command.Stdin = bytes.NewBufferString(specificVersionRequest)
		})

		It("that version along with all newer versions", func() {
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			<-session.Exited
			Expect(session.ExitCode()).To(Equal(0))

			result := []stemcellVersion{}
			err = json.Unmarshal(session.Out.Contents(), &result)
			Expect(err).NotTo(HaveOccurred())

			Expect(result[0]).To(Equal(stemcellVersion{
				"version": "3262.4",
			}))

			Expect(result[1]).To(Equal(stemcellVersion{
				"version": "3262.4.1",
			}))

			Expect(result[2]).To(Equal(stemcellVersion{
				"version": "3262.5",
			}))

			Expect(result[3]).To(Equal(stemcellVersion{
				"version": "3262.7",
			}))

			Expect(result).NotTo(ContainElement(stemcellVersion{
				"version": "3262.2",
			}))
		})
	})

	Context("when an older version is specified", func() {
		var command *exec.Cmd

		BeforeEach(func() {
			command = exec.Command(boshioCheck)
			command.Stdin = bytes.NewBufferString(oldVersionRequest)
		})

		It("that version along with all newer versions", func() {
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			<-session.Exited
			Expect(session.ExitCode()).To(Equal(0))

			result := []stemcellVersion{}
			err = json.Unmarshal(session.Out.Contents(), &result)
			Expect(err).NotTo(HaveOccurred())

			Expect(result[0]).To(Equal(stemcellVersion{
				"version": "3151",
			}))

			Expect(result[1]).To(Equal(stemcellVersion{
				"version": "3151.1",
			}))

			Expect(result).NotTo(ContainElement(stemcellVersion{
				"version": "3149",
			}))
		})
	})

	Context("when `force_regular` is true", func() {
		Context("and regular stemcell versions are available", func() {
			var command *exec.Cmd

			BeforeEach(func() {
				command = exec.Command(boshioCheck)
				command.Stdin = bytes.NewBufferString(bothTypesForceRegularRequest)
			})

			It("grabs the latest version with a Regular stemcell", func() {
				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				<-session.Exited
				Expect(session.ExitCode()).To(Equal(0))

				result := []stemcellVersion{}
				err = json.Unmarshal(session.Out.Contents(), &result)
				Expect(err).NotTo(HaveOccurred())

				Expect(result).To(HaveLen(1))
				Expect(result[0]["version"]).NotTo(BeEmpty())
			})
		})

		XContext("and only light stemcell versions are available", func() {
			var command *exec.Cmd

			BeforeEach(func() {
				command = exec.Command(boshioCheck)
				command.Stdin = bytes.NewBufferString(lightOnlyForceRegularRequest)
			})

			It("returns an empty version set", func() {
				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				<-session.Exited
				Expect(session.ExitCode()).To(Equal(0))

				result := []stemcellVersion{}
				err = json.Unmarshal(session.Out.Contents(), &result)
				Expect(err).NotTo(HaveOccurred())

				Expect(result).To(HaveLen(0))
			})
		})
	})

	Context("when an error occurs", func() {
		var command *exec.Cmd

		BeforeEach(func() {
			command = exec.Command(boshioCheck)
			command.Stdin = bytes.NewBufferString(specificVersionRequest)
		})

		Context("when the json cannot be read", func() {
			It("returns an error", func() {
				command.Stdin = bytes.NewBufferString("%%%%")
				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				<-session.Exited
				Expect(session.ExitCode()).To(Equal(1))

				Expect(session.Err).To(gbytes.Say("failed reading json: invalid character"))
			})
		})
	})
})
