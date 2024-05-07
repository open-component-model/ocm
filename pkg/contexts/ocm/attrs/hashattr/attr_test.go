package hashattr_test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/pkg/testutils"

	"github.com/open-component-model/ocm/pkg/contexts/config"
	"github.com/open-component-model/ocm/pkg/contexts/datacontext"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/attrs/hashattr"
	"github.com/open-component-model/ocm/pkg/runtime"
	"github.com/open-component-model/ocm/pkg/signing/hasher/sha512"
)

const NAME = "test"

var _ = Describe("attribute", func() {
	var cfgctx config.Context
	var ocmctx ocm.Context

	BeforeEach(func() {
		ocmctx = ocm.New(datacontext.MODE_EXTENDED)
		cfgctx = ocmctx.ConfigContext()
	})

	It("marshal/unmarshal", func() {
		cfg := hashattr.New(sha512.Algorithm)
		data := Must(json.Marshal(cfg))

		r := &hashattr.Config{}
		Expect(json.Unmarshal(data, r)).To(Succeed())
		Expect(r).To(Equal(cfg))
	})

	It("decode", func() {
		attr := &hashattr.Attribute{
			DefaultHasher: sha512.Algorithm,
		}

		r := Must(hashattr.AttributeType{}.Decode([]byte(sha512.Algorithm), runtime.DefaultYAMLEncoding))
		Expect(r).To(Equal(attr))
	})

	It("applies string", func() {
		MustBeSuccessful(cfgctx.GetAttributes().SetAttribute(hashattr.ATTR_KEY, sha512.Algorithm))
		attr := hashattr.Get(ocmctx)
		Expect(attr.GetHasher(ocmctx)).To(Equal(sha512.Handler{}))
	})

	It("applies config", func() {
		cfg := hashattr.New(sha512.Algorithm)

		MustBeSuccessful(cfgctx.ApplyConfig(cfg, "from test"))
		Expect(hashattr.Get(ocmctx).GetHasher(ocmctx)).To(Equal(sha512.Handler{}))
	})
})
