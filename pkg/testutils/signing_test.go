package testutils_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/testutils"
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
