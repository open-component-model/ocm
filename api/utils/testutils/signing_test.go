package testutils_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	common "ocm.software/ocm/api/utils/misc"
	"ocm.software/ocm/api/utils/testutils"
)

var _ = Describe("normalization", func() {
	It("compares with substitution variables", func() {
		exp := "A ${TEST}."
		res := "A testcase."
		vars := common.Properties{
			"TEST": "testcase",
		}
		Expect(res).To(testutils.StringEqualTrimmedWithContext(exp, common.Properties{}, vars))
		Expect(res).To(testutils.StringEqualTrimmedWithContext(exp, vars, common.Properties{}))
	})
})
