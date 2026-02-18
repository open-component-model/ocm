package httptimeoutattr_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestHTTPTimeout(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "HTTP Timeout Attribute")
}
