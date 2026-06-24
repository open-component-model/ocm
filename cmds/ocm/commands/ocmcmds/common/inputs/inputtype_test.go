package inputs_test

import (
	"encoding/json"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"ocm.software/ocm/api/utils/mime"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs/types/binary"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs/types/directory"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs/types/docker"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs/types/dockermulti"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs/types/file"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs/types/git"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs/types/helm"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs/types/maven"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs/types/npm"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs/types/ociartifact"
	ocminput "ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs/types/ocm"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs/types/spiff"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs/types/utf8"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs/types/wget"
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

	// Every input type registers a canonical name and a `/v1`-suffixed variant;
	// the types that already had a Title-cased alias on main additionally got a
	// `Title/v1` variant. A missing or typo'd entry in one of the init() blocks
	// would silently make that variant un-decodable, so the table below exercises
	// each registered literal end-to-end through the global scheme.
	DescribeTable("every registered alias decodes through the default scheme",
		func(typ string, prototype inputs.InputSpec) {
			decodeAs(typ, prototype)
		},
		Entry("binary", "binary", &binary.Spec{}),
		Entry("binary/v1", "binary/v1", &binary.Spec{}),

		Entry("dir", "dir", &directory.Spec{}),
		Entry("dir/v1", "dir/v1", &directory.Spec{}),
		Entry("Dir", "Dir", &directory.Spec{}),
		Entry("Dir/v1", "Dir/v1", &directory.Spec{}),

		Entry("docker", "docker", &docker.Spec{}),
		Entry("docker/v1", "docker/v1", &docker.Spec{}),

		Entry("dockermulti", "dockermulti", &dockermulti.Spec{}),
		Entry("dockermulti/v1", "dockermulti/v1", &dockermulti.Spec{}),

		Entry("file", "file", &file.Spec{}),
		Entry("file/v1", "file/v1", &file.Spec{}),
		Entry("File", "File", &file.Spec{}),
		Entry("File/v1", "File/v1", &file.Spec{}),

		Entry("git", "git", &git.Spec{}),
		Entry("git/v1", "git/v1", &git.Spec{}),
		Entry("Git", "Git", &git.Spec{}),
		Entry("Git/v1", "Git/v1", &git.Spec{}),

		Entry("maven", "maven", &maven.Spec{}),
		Entry("maven/v1", "maven/v1", &maven.Spec{}),
		Entry("Maven", "Maven", &maven.Spec{}),
		Entry("Maven/v1", "Maven/v1", &maven.Spec{}),

		Entry("npm", "npm", &npm.Spec{}),
		Entry("npm/v1", "npm/v1", &npm.Spec{}),
		Entry("NPM", "NPM", &npm.Spec{}),
		Entry("NPM/v1", "NPM/v1", &npm.Spec{}),

		Entry("ocm", "ocm", &ocminput.Spec{}),
		Entry("ocm/v1", "ocm/v1", &ocminput.Spec{}),

		Entry("spiff", "spiff", &spiff.Spec{}),
		Entry("spiff/v1", "spiff/v1", &spiff.Spec{}),

		Entry("utf8", "utf8", &utf8.Spec{}),
		Entry("utf8/v1", "utf8/v1", &utf8.Spec{}),
		Entry("UTF8", "UTF8", &utf8.Spec{}),
		Entry("UTF8/v1", "UTF8/v1", &utf8.Spec{}),

		Entry("wget", "wget", &wget.Spec{}),
		Entry("wget/v1", "wget/v1", &wget.Spec{}),
		Entry("Wget", "Wget", &wget.Spec{}),
		Entry("Wget/v1", "Wget/v1", &wget.Spec{}),
	)
})
