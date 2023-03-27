// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package install_test

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/cmds/ocm/testhelper"
	. "github.com/open-component-model/ocm/pkg/testutils"
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
		Expect(env.CatchOutput(buf).Execute("controller", "install", "-d", "-u", testServer.URL, "-a", testServer.URL)).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(`► installing ocm-controller with version latest
► got latest version "v0.0.1-test"
✔ successfully fetched install file
test: content
`))
	})

	It("install specific version", func() {
		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("controller", "install", "-d", "-u", testServer.URL, "-a", testServer.URL, "-v", "v0.1.0-test-2")).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(`► installing ocm-controller with version v0.1.0-test-2
✔ successfully fetched install file
test: content
`))
	})
})
