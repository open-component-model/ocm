// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package show_test

import (
	"bytes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/cmds/ocm/testhelper"
	. "github.com/open-component-model/ocm/pkg/testutils"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/mime"
)

const ARCH = "/tmp/ctf"
const NAMESAPCE = "mandelsoft/test"
const V13 = "v1.3"
const V131 = "v1.3.1"
const V132 = "v1.3.2"
const V132x = "v1.3.2-beta.1"
const V14 = "v1.4"
const V2 = "v2.0"
const OTHERVERS = "sometag"

var _ = Describe("Show OCI Tags", func() {
	var env *TestEnv
	BeforeEach(func() {
		env = NewTestEnv()

		env.OCICommonTransport(ARCH, accessio.FormatDirectory, func() {
			env.Namespace(NAMESAPCE, func() {
				env.Manifest(V13, func() {
					env.Tags(V131, OTHERVERS)
					env.Config(func() {
						env.BlobStringData(mime.MIME_JSON, "{}")
					})
					env.Layer(func() {
						env.BlobStringData(mime.MIME_TEXT, "data131")
					})
				})
				env.Manifest(V132, func() {
					env.Tags(V132x)
					env.Config(func() {
						env.BlobStringData(mime.MIME_JSON, "{}")
					})
					env.Layer(func() {
						env.BlobStringData(mime.MIME_TEXT, "data132")
					})
				})
				env.Manifest(V14, func() {
					env.Config(func() {
						env.BlobStringData(mime.MIME_JSON, "{}")
					})
					env.Layer(func() {
						env.BlobStringData(mime.MIME_TEXT, "data14")
					})
				})
				env.Manifest(V2, func() {
					env.Config(func() {
						env.BlobStringData(mime.MIME_JSON, "{}")
					})
					env.Layer(func() {
						env.BlobStringData(mime.MIME_TEXT, "data2")
					})
				})
			})
		})
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("lists tags", func() {
		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("oci", "tags", "show", "--repo", ARCH, NAMESAPCE)).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(`
sometag
v1.3
v1.3.1
v1.3.2
v1.3.2-beta.1
v1.4
v2.0
`))
	})

	It("lists tags for same artifact", func() {
		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("oci", "tags", "show", "--repo", ARCH, NAMESAPCE+":"+V13)).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(`
sometag
v1.3
v1.3.1
`))
	})

	It("lists semver tags", func() {
		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("oci", "tags", "show", "--semver", "--repo", ARCH, NAMESAPCE)).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(`
v1.3
v1.3.1
v1.3.2-beta.1
v1.3.2
v1.4
v2.0
`))
	})

	It("lists semver tags for same artifact", func() {
		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("oci", "tags", "show", "--semver", "--repo", ARCH, NAMESAPCE+":"+V13)).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(`
v1.3
v1.3.1
`))
	})
})
