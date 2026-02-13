package ocireg_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestOCIReg(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "OCIReg")
}
