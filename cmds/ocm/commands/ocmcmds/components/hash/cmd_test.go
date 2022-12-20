// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package hash_test

import (
	"bytes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/cmds/ocm/testhelper"
	. "github.com/open-component-model/ocm/pkg/testutils"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/resourcetypes"
	"github.com/open-component-model/ocm/pkg/mime"
)

const ARCH = "/tmp/ca"
const VERSION = "v1"
const COMP = "test.de/x"
const PROVIDER = "mandelsoft"

var _ = Describe("Test Environment", func() {
	var env *TestEnv

	BeforeEach(func() {
		env = NewTestEnv()
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("hash component archive", func() {
		env.ComponentArchive(ARCH, accessio.FormatDirectory, COMP, VERSION, func() {
			env.Provider(PROVIDER)
		})

		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("hash", "components", ARCH, "-o", "wide")).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(
			`
COMPONENT VERSION HASH                                                             NORMALIZED FORM
test.de/x v1      7aff33d597be2bd358672bd79fb904666d955c3890d75d32cde356eb86355c65 [{"component":[{"componentReferences":[]},{"name":"test.de/x"},{"provider":"mandelsoft"},{"resources":[]},{"sources":[]},{"version":"v1"}]},{"meta":[{"schemaVersion":"v2"}]}]
`))
	})

	It("hash component archive with resources", func() {
		env.ComponentArchive(ARCH, accessio.FormatDirectory, COMP, VERSION, func() {
			env.Provider(PROVIDER)
			env.Resource("test", VERSION, resourcetypes.PLAIN_TEXT, metav1.LocalRelation, func() {
				env.BlobStringData(mime.MIME_TEXT, "testdata")
			})
		})

		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("hash", "components", ARCH, "-o", "wide")).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(
			`
COMPONENT      : test.de/x
VERSION        : v1
HASH           : c3f4916f8913fe99ad0abb18106c5f4b89950bae19b57aeb342bc054eebad056
NORMALIZED FORM: [{"component":[{"componentReferences":[]},{"name":"test.de/x"},{"provider":"mandelsoft"},{"resources":[[{"digest":[{"hashAlgorithm":"sha256"},{"normalisationAlgorithm":"genericBlobDigest/v1"},{"value":"810ff2fb242a5dee4220f2cb0e6a519891fb67f2f828a6cab4ef8894633b1f50"}]},{"name":"test"},{"relation":"local"},{"type":"plainText"},{"version":"v1"}]]},{"sources":[]},{"version":"v1"}]},{"meta":[{"schemaVersion":"v2"}]}]
---`))
	})

	It("hash component archive with resources", func() {
		env.ComponentArchive(ARCH, accessio.FormatDirectory, COMP, VERSION, func() {
			env.Provider(PROVIDER)
			env.Resource("test", VERSION, resourcetypes.PLAIN_TEXT, metav1.LocalRelation, func() {
				env.BlobStringData(mime.MIME_TEXT, "testdata")
			})
		})

		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("hash", "components", ARCH, "--actual", "-o", "wide")).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(
			`
COMPONENT      : test.de/x
VERSION        : v1
HASH           : f8a0fcfc2e0501d35657aa88ca6848abc15fc527a3a2b0055f72db977cf01d6b
NORMALIZED FORM: [{"component":[{"componentReferences":[]},{"name":"test.de/x"},{"provider":"mandelsoft"},{"resources":[[{"name":"test"},{"relation":"local"},{"type":"plainText"},{"version":"v1"}]]},{"sources":[]},{"version":"v1"}]},{"meta":[{"schemaVersion":"v2"}]}]
---`))
	})
})
