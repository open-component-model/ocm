// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

//go:build unix

package describe_test

import (
	"bytes"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/open-component-model/ocm/cmds/ocm/testhelper"
	. "github.com/open-component-model/ocm/pkg/testutils"
)

const PLUGINS = "/testdata"

var _ = Describe("Test Environment", func() {
	var env *TestEnv
	var path string

	BeforeEach(func() {
		env = NewTestEnv(TestData())

		p, err := filepath.Abs("testdata")
		Expect(err).To(Succeed())
		path = p
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("get plugins", func() {
		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("-X", "plugindir="+path, "describe", "plugins")).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(
			`
Plugin Name:      action
Plugin Version:   v1
Path:             ` + path + `/action
Status:           valid
Capabilities:     Actions
Source:           manually installed
Description: 
      a test plugin with action oci.repository.prepare
Actions:
- Name: oci.repository.prepare
    Prepare the usage of a repository in an OCI registry.
  Info:
    test action
  Versions:
  - v1 (best matching)
  Handler accepts standard credentials
----------------------
Plugin Name:      test
Plugin Version:   v1
Path:             ` + path + `/test
Status:           valid
Capabilities:     Access Methods
Source:           manually installed
Description: 
      a test plugin with access method test
Access Methods:
- Name: test
  Versions:
  - Version: v1
*** found 2 plugins
`))
	})
})
