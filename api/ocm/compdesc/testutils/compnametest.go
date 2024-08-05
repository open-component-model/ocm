package testutils

import (
	"fmt"
	"regexp"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/sirupsen/logrus"

	"ocm.software/ocm/api/utils/runtime"
)

func TestCompName(dataBytes []byte, err error) {
	ExpectWithOffset(1, err).To(Succeed())

	Context("component name validation", func() {
		var scheme map[string]interface{}
		Expect(runtime.DefaultYAMLEncoding.Unmarshal(dataBytes, &scheme)).To(Succeed())

		pattern := scheme["$defs"].(map[string]interface{})["componentName"].(map[string]interface{})["pattern"].(string)

		logrus.Infof("pattern=%s", pattern)

		expr, err := regexp.Compile(pattern)
		Expect(err).To(Succeed())

		Check := func(s string, exp bool) {
			if expr.MatchString(s) != exp {
				Fail(fmt.Sprintf("%s[%t] failed\n", s, exp), 1)
			}
		}

		It("parsed valid names", func() {
			Check("github.wdf.sap.corp/kubernetes/landscape-setup", true)
			Check("weave.works/registry/app", true)
			Check("internal.github.org/registry/app", true)
			Check("a.de/c", true)
			Check("a.de/c/d/e-f", true)
			Check("a.de/c/d/e_f", true)
			Check("a.de/c/d/e", true)
			Check("a.de/c/d/e.f", true)
		})

		It("rejects invalid names", func() {
			Check("a.de/", false)
			Check("a.de/a/", false)
			Check("a.de//a", false)
			Check("a.de/a.", false)
		})
	})
}
