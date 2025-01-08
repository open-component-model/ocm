package runtime_test

import (
	"encoding/json"
	"reflect"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mandelsoft/logging"

	ocmlog "ocm.software/ocm/api/utils/logging"
	"ocm.software/ocm/api/utils/runtime"
)

func getOutput(log logging.Logger, in runtime.TypedObject, encoding runtime.Encoding) (runtime.TypedObject, string, error) {
	t := reflect.TypeOf(in)
	log.Info("in", "type", t)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	var p reflect.Value

	if t.Kind() == reflect.Map {
		p = reflect.New(t)
		m := reflect.MakeMap(t)
		log.Info("pointer", "type", p.Type())
		p.Elem().Set(m)
	} else {
		p = reflect.New(t)
	}
	out := p.Interface().(runtime.TypedObject)

	log.Info("out", "out", out)
	data, err := encoding.Marshal(in)
	if err != nil {
		return nil, "", err
	}
	err = encoding.Unmarshal(data, out)
	return out, string(data), err
}

var _ = Describe("*** unstructured", func() {
	result := "{\"type\":\"test\"}"
	log := ocmlog.Logger()

	It("unmarshal simple unstructured", func() {
		un := runtime.NewEmptyUnstructured("test")
		data, err := json.Marshal(un)
		Expect(err).To(Succeed())
		Expect(string(data)).To(Equal("{\"type\":\"test\"}"))

		un = &runtime.UnstructuredTypedObject{}
		log.Info("out", "object", un)
		err = json.Unmarshal(data, un)
		Expect(err).To(Succeed())
		Expect(un.GetType()).To(Equal("test"))
	})

	It("unmarshal json test", func() {
		out, data, err := getOutput(log, runtime.NewEmptyUnstructured("test"), runtime.DefaultJSONEncoding)
		Expect(err).To(Succeed())
		Expect(out.GetType()).To(Equal("test"))
		Expect(data).To(Equal(result))

		out, data, err = getOutput(log, runtime.NewEmptyUnstructuredVersioned("test"), runtime.DefaultJSONEncoding)
		Expect(err).To(Succeed())
		Expect(out.GetType()).To(Equal("test"))
		Expect(data).To(Equal(result))
	})

	It("unmarshal yaml test", func() {
		out, data, err := getOutput(log, runtime.NewEmptyUnstructured("test"), runtime.DefaultYAMLEncoding)
		Expect(err).To(Succeed())
		Expect(out.GetType()).To(Equal("test"))
		Expect(data).To(Equal("type: test\n"))

		out, data, err = getOutput(log, runtime.NewEmptyUnstructuredVersioned("test"), runtime.DefaultYAMLEncoding)
		Expect(err).To(Succeed())
		Expect(out.GetType()).To(Equal("test"))
		Expect(data).To(Equal("type: test\n"))
	})

	Context("unstructured match", func() {
		It("matches complete", func() {
			a := runtime.UnstructuredMap{
				"map": map[string]interface{}{
					"alice": 25,
					"bob":   "husband",
					"peter": true,
					"sally": 3.14,
				},
				"array": []interface{}{
					"string",
					true,
					25,
					3.14,
				},
			}
			b := runtime.UnstructuredMap{
				"map": map[string]interface{}{
					"alice": 25,
					"bob":   "husband",
					"peter": true,
					"sally": 3.14,
				},
				"array": []interface{}{
					"string",
					true,
					25,
					3.14,
				},
			}
			Expect(a.Match(b)).To(BeTrue())
			Expect(b.Match(a)).To(BeTrue())
		})
		It("matches initial", func() {
			a := runtime.UnstructuredMap{
				"map":    map[string]interface{}{},
				"array":  []interface{}{},
				"string": "",
				"bool":   false,
				"int":    0,
				"float":  0.0,
			}
			b := runtime.UnstructuredMap{}
			Expect(a.Match(b)).To(BeTrue())
			Expect(b.Match(a)).To(BeTrue())
		})

		It("handles string mismatch", func() {
			a := runtime.UnstructuredMap{
				"map":    map[string]interface{}{},
				"array":  []interface{}{},
				"string": "x",
				"bool":   false,
				"int":    0,
				"float":  0.0,
			}
			b := runtime.UnstructuredMap{}
			Expect(a.Match(b)).To(BeFalse())
			Expect(b.Match(a)).To(BeFalse())
		})
		It("handles bool mismatch", func() {
			a := runtime.UnstructuredMap{
				"map":    map[string]interface{}{},
				"array":  []interface{}{},
				"string": "",
				"bool":   true,
				"int":    0,
				"float":  0.0,
			}
			b := runtime.UnstructuredMap{}
			Expect(a.Match(b)).To(BeFalse())
			Expect(b.Match(a)).To(BeFalse())
		})
		It("handles int mismatch", func() {
			a := runtime.UnstructuredMap{
				"map":    map[string]interface{}{},
				"array":  []interface{}{},
				"string": "",
				"bool":   false,
				"int":    1,
				"float":  0.0,
			}
			b := runtime.UnstructuredMap{}
			Expect(a.Match(b)).To(BeFalse())
			Expect(b.Match(a)).To(BeFalse())
		})
		It("handles float mismatch", func() {
			a := runtime.UnstructuredMap{
				"map":    map[string]interface{}{},
				"array":  []interface{}{},
				"string": "",
				"bool":   false,
				"int":    0,
				"float":  3.14,
			}
			b := runtime.UnstructuredMap{}
			Expect(a.Match(b)).To(BeFalse())
			Expect(b.Match(a)).To(BeFalse())
		})
		It("handles map mismatch", func() {
			a := runtime.UnstructuredMap{
				"map":    map[string]interface{}{"alice": 25},
				"array":  []interface{}{},
				"string": "",
				"bool":   false,
				"int":    0,
				"float":  0.0,
			}
			b := runtime.UnstructuredMap{}
			Expect(a.Match(b)).To(BeFalse())
			Expect(b.Match(a)).To(BeFalse())
		})
		It("handles array mismatch", func() {
			a := runtime.UnstructuredMap{
				"map":    map[string]interface{}{},
				"array":  []interface{}{"alice"},
				"string": "",
				"bool":   false,
				"int":    0,
				"float":  0.0,
			}
			b := runtime.UnstructuredMap{}
			Expect(a.Match(b)).To(BeFalse())
			Expect(b.Match(a)).To(BeFalse())
		})
		It("handles structure mismatch", func() {
			a := runtime.UnstructuredMap{
				"map": []interface{}{},
			}
			b := runtime.UnstructuredMap{
				"map": map[string]interface{}{},
			}
			Expect(a.Match(b)).To(BeFalse())
			Expect(b.Match(a)).To(BeFalse())
		})
		It("handles type mismatch", func() {
			a := runtime.UnstructuredMap{
				"string": "alice",
			}
			b := runtime.UnstructuredMap{
				"string": 25,
			}
			Expect(a.Match(b)).To(BeFalse())
			Expect(b.Match(a)).To(BeFalse())
		})
	})
})
