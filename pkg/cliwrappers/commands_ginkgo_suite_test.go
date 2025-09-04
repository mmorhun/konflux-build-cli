package cliwrappers

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestCliwrappers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CLI Wrappers Suite")
}
