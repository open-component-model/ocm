package grammar

import (
	"regexp"
	"testing"

	tool "github.com/mandelsoft/goutils/regexutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestConfig(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "OCI Test Suite")
}

func CheckURI(ref string, parts ...string) {
	Check(ref, TypedURIRegexp, parts...)
}

func Check(ref string, exp *regexp.Regexp, parts ...string) {
	spec := exp.FindSubmatch([]byte(ref))
	if len(parts) == 0 {
		Expect(spec).To(BeNil())
	} else {
		result := make([]string, len(spec))
		for i, v := range spec {
			result[i] = string(v)
		}
		Expect(result).To(Equal(append([]string{ref}, parts...)))
	}
}

func Type(t string) string {
	if t == "" {
		return t
	}
	return t + "::"
}

func Sub(t string) string {
	if t == "" {
		return t
	}
	return "/" + t
}

func Vers(t string) string {
	if t == "" {
		return t
	}
	return ":" + t
}

var _ = Describe("ref matching", func() {
	Context("parts", func() {
		It("path port", func() {
			Check("/some/path/docker.sock:100", tool.Anchored(tool.Capture(PathPortRegexp)), "/some/path/docker.sock:100")
		})

		It("host port", func() {
			Check("github:100", tool.Anchored(tool.Capture(HostPortRegexp)), "github:100")
		})

		It("IP port", func() {
			Check("100.1.2.10:100", tool.Anchored(tool.Capture(HostPortRegexp)), "100.1.2.10:100")
		})
	})

	Context("types refs", func() {
		t := "DockerDaemon"
		s := "unix"
		p := "/some/path/docker.sock:100"
		r := "repo"
		v := "test"

		It("fails", func() {
			CheckURI("DockerDaemon::unix:///some/path/docker.sock:100//repo:test", t, s, p, r, v, "")
			CheckURI("DockerDaemon::unix:///some/path/docker.sock:100", t, s, p, "", "", "")
			CheckURI("DockerDaemon::unix://some/path/docker.sock:100//repo:test", t, s, p[1:], r, v, "")
		})
	})
})
