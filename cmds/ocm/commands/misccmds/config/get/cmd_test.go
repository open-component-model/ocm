package get_test

import (
	"bytes"
	"encoding/json"

	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "ocm.software/ocm/cmds/ocm/testhelper"

	"github.com/mandelsoft/vfs/pkg/vfs"
)

var _ = Describe("Get config", func() {
	var env *TestEnv

	BeforeEach(func() {
		env = NewTestEnv()
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("provides json output", func() {
		var buf bytes.Buffer

		MustBeSuccessful(env.CatchOutput(&buf).Execute("get", "config", "-o", "json"))

		var r map[string]interface{}
		MustBeSuccessful(json.Unmarshal(buf.Bytes(), &r))
	})

	It("writes json output", func() {
		var buf bytes.Buffer

		MustBeSuccessful(env.CatchOutput(&buf).Execute("get", "config", "-o", "json", "-O", "config"))

		Expect(buf.String()).To(Equal("config written to \"config\"\n"))
		Expect(vfs.Exists(env.FileSystem(), "config")).To(BeTrue())

		data := Must(vfs.ReadFile(env.FileSystem(), "config"))
		var r map[string]interface{}
		MustBeSuccessful(json.Unmarshal(data, &r))
	})
})
