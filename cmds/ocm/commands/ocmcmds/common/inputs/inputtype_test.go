package inputs_test

import (
	"encoding/json"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"ocm.software/ocm/api/utils/mime"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs/types/file"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs/types/helm"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs/types/ociartifact"
)

var _ = Describe("Blob Inputs", func() {
	scheme := inputs.DefaultInputTypeScheme
	spec := file.New("test", mime.MIME_TEXT, false)

	It("simple decode", func() {
		data, err := json.Marshal(spec)
		Expect(err).To(Succeed())

		s, err := scheme.DecodeInputSpec(data, nil)
		Expect(err).To(Succeed())
		Expect(s).To(Equal(spec))
	})

	It("generic eval", func() {
		gen, err := inputs.ToGenericInputSpec(spec)
		Expect(err).To(Succeed())

		Expect(gen.Evaluate(scheme)).To(Equal(spec))
	})

	It("generic marshal effective", func() {
		gen, err := inputs.ToGenericInputSpec(spec)
		Expect(err).To(Succeed())

		data, err := json.Marshal(gen)
		Expect(err).To(Succeed())

		s, err := scheme.DecodeInputSpec(data, nil)
		Expect(err).To(Succeed())
		Expect(s).To(Equal(spec))
	})

	It("generic marshal effective", func() {
		gen, err := inputs.ToGenericInputSpec(spec)
		Expect(err).To(Succeed())
		Expect(gen.Evaluate(scheme)).To(Equal(spec))

		data, err := json.Marshal(gen)
		Expect(err).To(Succeed())

		s, err := scheme.DecodeInputSpec(data, nil)
		Expect(err).To(Succeed())
		Expect(s).To(Equal(spec))
	})

	It("generic unmarshal", func() {
		gen := inputs.GenericInputSpec{}

		data, err := json.Marshal(spec)
		Expect(err).To(Succeed())

		Expect(json.Unmarshal(data, &gen)).To(Succeed())

		Expect(gen.Evaluate(scheme)).To(Equal(spec))
	})
})

var _ = Describe("Versioned input type aliases", func() {
	scheme := inputs.DefaultInputTypeScheme

	// decodeAs verifies that a type literal decodes to the expected concrete spec type.
	decodeAs := func(typ string, prototype inputs.InputSpec) {
		data := []byte(fmt.Sprintf(`{"type":%q}`, typ))
		s, err := scheme.DecodeInputSpec(data, nil)
		Expect(err).To(Succeed(), "type %q should decode", typ)
		Expect(s).To(BeAssignableToTypeOf(prototype), "type %q should produce %T", typ, prototype)
	}

	Context("helm", func() {
		It("decodes both unversioned and /v1 forms", func() {
			decodeAs("helm", &helm.Spec{})
			decodeAs("helm/v1", &helm.Spec{})
			decodeAs("Helm", &helm.Spec{})
			decodeAs("Helm/v1", &helm.Spec{})
		})

		It("constructor still emits the unversioned canonical form", func() {
			data, err := json.Marshal(helm.New("./chart"))
			Expect(err).To(Succeed())

			var probe struct {
				Type string `json:"type"`
			}
			Expect(json.Unmarshal(data, &probe)).To(Succeed())
			Expect(probe.Type).To(Equal("helm"))
		})
	})

	Context("ociartifact", func() {
		It("decodes both canonical and legacy aliases incl. /v1", func() {
			decodeAs("ociArtifact", &ociartifact.Spec{})
			decodeAs("ociArtifact/v1", &ociartifact.Spec{})
			decodeAs("ociImage", &ociartifact.Spec{})
			decodeAs("ociImage/v1", &ociartifact.Spec{})
		})
	})
})
