package acceptance_test

import (
	"testing"

	"github.com/onsi/gomega/gexec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	boshioCheck string
	boshioIn    string
)

var _ = BeforeSuite(func() {
	var err error
	boshioCheck, err = gexec.Build("github.com/concourse/bosh-io-stemcell-resource/cmd/check")
	Expect(err).NotTo(HaveOccurred())

	boshioIn, err = gexec.Build("github.com/concourse/bosh-io-stemcell-resource/cmd/in")
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
})

func TestAcceptance(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Acceptance Suite")
}
