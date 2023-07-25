// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

//go:build unix

package describe_test

import (
	"bytes"

	"github.com/mandelsoft/filepath/pkg/filepath"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/v2/cmds/ocm/testhelper"
)

const PLUGINS = "/testdata"

var _ = Describe("Test Environment", func() {
	var env *TestEnv
	var path string

	BeforeEach(func() {
		env = NewTestEnv(TestData())

		// use os filesystem here
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

    The hostname of the target repository is used as selector. The action should
    assure, that the requested repository is available on the target OCI registry.
    
    Spec version v1 uses the following specification fields:
    - «hostname» *string*: The  hostname of the OCI registry.
    - «repository» *string*: The OCI repository name.
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
