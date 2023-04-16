// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package bootstrap_test

import (
	"bytes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/cmds/ocm/testhelper"
	. "github.com/open-component-model/ocm/pkg/testutils"

	"github.com/mandelsoft/vfs/pkg/vfs"

	"github.com/open-component-model/ocm/cmds/ocm/commands/toicmds/config/bootstrap"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	v1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/mime"
	"github.com/open-component-model/ocm/pkg/toi"
)

const ARCH = "/tmp/ctf"
const VERSION = "v1"
const COMP1 = "test.de/a"
const COMP2 = "test.de/b"
const PROVIDER = "mandelsoft"

var _ = Describe("Test Environment", func() {
	var env *TestEnv

	respkg := `
description: with config example by resource
additionalResources:
  ` + toi.AdditionalResourceConfigFile + `:
    content:
       param: value
`
	cntpkg := `
description: with example by content
additionalResources:
  ` + toi.AdditionalResourceCredentialsFile + `:
    content: |
      credentials: none
  ` + toi.AdditionalResourceConfigFile + `:
    content:
       param: value
`

	BeforeEach(func() {
		env = NewTestEnv()
		env.OCMCommonTransport(ARCH, accessio.FormatDirectory, func() {
			env.Component(COMP1, func() {
				env.Version(VERSION, func() {
					env.Provider(PROVIDER)
					env.Resource("package", VERSION, toi.TypeTOIPackage, v1.LocalRelation, func() {
						env.BlobStringData(toi.PackageSpecificationMimeType, respkg)
					})
					env.Resource("example", VERSION, toi.TypeTOIPackage, v1.LocalRelation, func() {
						env.BlobStringData(mime.MIME_YAML, "param: value")
					})
				})
			})
			env.Component(COMP2, func() {
				env.Version(VERSION, func() {
					env.Provider(PROVIDER)
					env.Resource("package", VERSION, toi.TypeTOIPackage, v1.LocalRelation, func() {
						env.BlobStringData(toi.PackageSpecificationMimeType, cntpkg)
					})
				})
			})
		})
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("config by resource", func() {
		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("bootstrap", "config", ARCH+"//"+COMP1)).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(
			`
Warning: repository is no OCI registry, consider importing it or use upload repository with option ' -X ociuploadrepo=...
found package resource "package" in test.de/a:v1

Package Description:
  with config example by resource

writing configuration template...
TOIParameters: 17 byte(s) written
no credentials template configured
`))
		data := Must(vfs.ReadFile(env.FileSystem(), bootstrap.DEFAULT_PARAMETER_FILE))
		Expect(string(data)).To(Equal(`param: value
`))
	})
	It("config by content", func() {
		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("bootstrap", "config", ARCH+"//"+COMP2)).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(
			`
Warning: repository is no OCI registry, consider importing it or use upload repository with option ' -X ociuploadrepo=...
found package resource "package" in test.de/b:v1

Package Description:
  with example by content

writing configuration template...
TOIParameters: 17 byte(s) written
writing credentials template...
TOICredentials: 18 byte(s) written
`))
		data := Must(vfs.ReadFile(env.FileSystem(), bootstrap.DEFAULT_PARAMETER_FILE))
		Expect(string(data)).To(Equal(`param: value
`))
		data = Must(vfs.ReadFile(env.FileSystem(), bootstrap.DEFAULT_CREDENTIALS_FILE))
		Expect(string(data)).To(Equal(`credentials: none
`))
	})
})
