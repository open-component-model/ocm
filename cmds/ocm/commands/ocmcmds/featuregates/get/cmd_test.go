//go:build unix

package get_test

import (
	"bytes"

	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "ocm.software/ocm/cmds/ocm/testhelper"
)

const PLUGINS = "/testdata"

var _ = Describe("Test Environment", func() {
	var env *TestEnv

	BeforeEach(func() {
		env = NewTestEnv()
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("get features", func() {
		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("-X", `featuregates={ "features": { "test": {"mode": "on"}}}`, "get", "fg")).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(
			`
FEATURE ENABLED MODE SHORT
test    enabled on   <unknown>
`))
	})

	It("get wide feature", func() {
		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("-X", `featuregates={ "features": { "test": {"mode": "on", "attributes": { "attr": "value"}}}}`, "get", "fg", "-o", "wide")).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(
			`
FEATURE ENABLED MODE SHORT     DESCRIPTION ATTRIBUTES
test    enabled on   <unknown>             {"attr":"value"}
`))
	})

	It("get yaml feature", func() {
		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("-X", `featuregates={ "features": { "test": {"mode": "on", "attributes": { "attr": "value"}}}}`, "get", "fg", "-o", "yaml")).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(
			`
---
attributes:
  attr: value
description: ""
enabled: false
mode: "on"
name: test
short: <unknown>
`))
	})
})
