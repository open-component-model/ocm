package bootstrap_test

import (
	"bytes"

	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "ocm.software/ocm/cmds/ocm/testhelper"

	"github.com/mandelsoft/vfs/pkg/vfs"

	v1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/tools/toi"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/mime"
	"ocm.software/ocm/cmds/ocm/commands/toicmds/config/bootstrap"
)

const (
	ARCH     = "/tmp/ctf"
	VERSION  = "v1"
	COMP1    = "test.de/a"
	COMP2    = "test.de/b"
	PROVIDER = "mandelsoft"
)

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
