package boshio_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestBoshio(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Boshio Suite")
}
