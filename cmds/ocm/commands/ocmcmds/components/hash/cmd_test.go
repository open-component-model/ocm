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
test.de/x v1      37f7f500d87f4b0a8765649f7c047db382e272b73e042805131df57279991b2b [{"component":[{"componentReferences":[]},{"name":"test.de/x"},{"provider":"mandelsoft"},{"resources":[]},{"version":"v1"}]},{"meta":[{"schemaVersion":"v2"}]}]
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
HASH           : 49145d28aee8dfebdde5432bdbe2cf8e365d56cbd17fbfa959770057d6127edc
NORMALIZED FORM: [{"component":[{"componentReferences":[]},{"name":"test.de/x"},{"provider":"mandelsoft"},{"resources":[[{"digest":[{"hashAlgorithm":"SHA-256"},{"normalisationAlgorithm":"genericBlobDigest/v1"},{"value":"810ff2fb242a5dee4220f2cb0e6a519891fb67f2f828a6cab4ef8894633b1f50"}]},{"name":"test"},{"relation":"local"},{"type":"plainText"},{"version":"v1"}]]},{"version":"v1"}]},{"meta":[{"schemaVersion":"v2"}]}]
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
HASH           : aa8fc0f82d47ae7dd30016e0d99a720081bf1018b72d0642cc16cab7d0f6673d
NORMALIZED FORM: [{"component":[{"componentReferences":[]},{"name":"test.de/x"},{"provider":"mandelsoft"},{"resources":[[{"name":"test"},{"relation":"local"},{"type":"plainText"},{"version":"v1"}]]},{"version":"v1"}]},{"meta":[{"schemaVersion":"v2"}]}]
---`))
	})
})
