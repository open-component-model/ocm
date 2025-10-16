package labelsel_test

import (
	"bytes"

	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mikefarah/yq/v4/pkg/yqlib"
	"sigs.k8s.io/yaml"

	v1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/selectors"
	"ocm.software/ocm/api/ocm/selectors/labelsel"
)

func Parse(data []byte) (*yqlib.CandidateNode, error) {
	decoder := yqlib.NewYamlDecoder(yqlib.YamlPreferences{})
	err := decoder.Init(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	return decoder.Decode()
}

var _ = Describe("yq label values", func() {
	data := `
people:
- name: alice
  age: 25
- name: bob
  age: 26
data:
  attr: value
`
	Context("yq", func() {
		It("", func() {
			doc := Must(Parse([]byte(data)))
			eval := yqlib.NewAllAtOnceEvaluator()
			r := Must(eval.EvaluateNodes(".people[0].name", doc))
			Expect(r.Len()).To(Equal(1))
			e := r.Front()
			v := e.Value.(*yqlib.CandidateNode)
			data := Must(v.MarshalJSON())
			Expect(data).To(YAMLEqual("alice"))
		})
	})

	Context("labels", func() {
		var datav map[string]interface{}
		labels := v1.Labels{}
		MustBeSuccessful(yaml.Unmarshal([]byte(data), &datav))
		MustBeSuccessful(labels.SetValue("data", datav))
		MustBeSuccessful(labels.SetValue("string", "some data"))

		It("check complex data", func() {
			m := labelsel.YQExpression(".data", datav["data"])
			Expect(m).NotTo(BeNil())
			Expect(selectors.ValidateSelectors(m)).NotTo(HaveOccurred())

			Expect(m.MatchLabel(&labels[0])).To(BeTrue())
			Expect(m.MatchLabel(&labels[1])).To(BeFalse())
		})

		It("check complex expression", func() {
			m := labelsel.YQExpression(".people[0].name", "alice")
			Expect(m).NotTo(BeNil())
			Expect(selectors.ValidateSelectors(m)).NotTo(HaveOccurred())

			Expect(m.MatchLabel(&labels[0])).To(BeTrue())
			Expect(m.MatchLabel(&labels[1])).To(BeFalse())
		})

		It("detects error", func() {
			m := labelsel.YQExpression(".people[0]].name", "alice")
			Expect(m).NotTo(BeNil())
			Expect(selectors.ValidateSelectors(m)).To(MatchError("error in selector list: YQExpression selector: bad expression, could not find matching `)`"))
		})

		It("detects error in expressions", func() {
			m := labelsel.YQExpression(".people[0]].name", "alice")
			Expect(m).NotTo(BeNil())
			Expect(selectors.ValidateSelectors(labelsel.Or(m))).To(MatchError("error in selector list: or: YQExpression selector: bad expression, could not find matching `)`"))
			Expect(selectors.ValidateSelectors(labelsel.And(m))).To(MatchError("error in selector list: and: YQExpression selector: bad expression, could not find matching `)`"))
			Expect(selectors.ValidateSelectors(labelsel.Not(m))).To(MatchError("error in selector list: not: YQExpression selector: bad expression, could not find matching `)`"))
		})
	})
})
