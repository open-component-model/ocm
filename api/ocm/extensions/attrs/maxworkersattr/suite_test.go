package maxworkersattr_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestMaxWorkers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "MaxWorkers Attribute")
}
