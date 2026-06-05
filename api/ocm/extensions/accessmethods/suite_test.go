package accessmethods_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestAliasSurface(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "access method alias surface (#1979)")
}
