package attrs_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"ocm.software/ocm/api/config"
	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/datacontext"
	local "ocm.software/ocm/api/datacontext/config/attrs"
	"ocm.software/ocm/api/utils/runtime"
)

const ATTR_KEY = "test"

func init() {
	datacontext.RegisterAttributeType(ATTR_KEY, AttributeType{})
}

type AttributeType struct{}

func (a AttributeType) Name() string {
	return ATTR_KEY
}

func (a AttributeType) Description() string {
	return `
A Test attribute.
`
}

type Attribute struct {
	Value string `json:"value"`
}

func (a AttributeType) Encode(v interface{}, marshaller runtime.Marshaler) ([]byte, error) {
	if _, ok := v.(*Attribute); !ok {
		return nil, fmt.Errorf("boolean required")
	}
	return marshaller.Marshal(v)
}

func (a AttributeType) Decode(data []byte, unmarshaller runtime.Unmarshaler) (interface{}, error) {
	var value Attribute
	err := unmarshaller.Unmarshal(data, &value)
	return &value, err
}

////////////////////////////////////////////////////////////////////////////////

var _ = Describe("generic attributes", func() {
	attribute := &Attribute{"TEST"}
	var ctx config.Context

	BeforeEach(func() {
		ctx = config.WithSharedAttributes(datacontext.New(nil)).New()
	})

	Context("applies", func() {
		It("applies later attribute config", func() {
			sub := credentials.WithConfigs(ctx).New()
			spec := local.New()
			Expect(spec.AddAttribute(ATTR_KEY, attribute)).To(Succeed())
			Expect(ctx.ApplyConfig(spec, "test")).To(Succeed())

			Expect(sub.GetAttributes().GetAttribute(ATTR_KEY, nil)).To(Equal(attribute))
		})

		It("applies earlier attribute config", func() {
			spec := local.New()
			Expect(spec.AddAttribute(ATTR_KEY, attribute)).To(Succeed())
			Expect(ctx.ApplyConfig(spec, "test")).To(Succeed())

			sub := credentials.WithConfigs(ctx).New()
			Expect(sub.GetAttributes().GetAttribute(ATTR_KEY, nil)).To(Equal(attribute))
		})
	})
})
