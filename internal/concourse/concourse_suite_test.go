package concourse_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestConcourse(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Concourse Suite")
}
