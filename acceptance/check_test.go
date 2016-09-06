package acceptance_test

import (
	"bytes"
	"os/exec"

	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const noVersionRequest = `
{
	"source": {
		"name": "bosh-aws-xen-hvm-ubuntu-trusty-go_agent"
	},
	"version": {}
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

			Eventually(session, "10s").Should(gexec.Exit(0))
			Expect(session.Out).NotTo(gbytes.Say(`{"version":"3262.7"}`))
			Expect(session.Out).NotTo(gbytes.Say(`{"version":"3262.5"}`))
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

			Eventually(session, "10s").Should(gexec.Exit(0))
			Expect(session.Out).To(gbytes.Say(`{"version":"3262.7"}`))
			Expect(session.Out).To(gbytes.Say(`{"version":"3262.5"}`))
			Expect(session.Out).To(gbytes.Say(`{"version":"3262.4.1"}`))
			Expect(session.Out).NotTo(gbytes.Say(`{"version":"3262.2"}`))
		})
	})
})