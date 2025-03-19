package content_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"testing"
)

func TestContent(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Content Suite")
}
