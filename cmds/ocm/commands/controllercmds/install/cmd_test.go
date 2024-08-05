package install_test

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"

	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "ocm.software/ocm/cmds/ocm/testhelper"

	"github.com/mandelsoft/filepath/pkg/filepath"
)

var _ = Describe("Test Environment", func() {
	var (
		env        *TestEnv
		testServer *httptest.Server
	)

	BeforeEach(func() {
		env = NewTestEnv()
		testServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.Contains(r.URL.String(), "download") {
				content, err := os.ReadFile(filepath.Join("testdata", "install.yaml"))
				if err != nil {
					fmt.Fprintf(w, "failed")
					return
				}

				fmt.Fprintf(w, string(content))
				return
			}

			fmt.Fprintf(w, `{
	"tag_name": "v0.0.1-test"
}
`)
		}))
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("install latest version", func() {
		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("controller", "install", "-d", "-s", "-u", testServer.URL, "-a", testServer.URL)).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(`test: content
`))
	})

	It("install specific version", func() {
		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("controller", "install", "-d", "-s", "-u", testServer.URL, "-a", testServer.URL, "-v", "v0.1.0-test-2")).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(`test: content
`))
	})
})
