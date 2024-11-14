package downloaderoption_test

import (
	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mandelsoft/goutils/generics"
	"github.com/spf13/pflag"

	"ocm.software/ocm/api/ocm"
	me "ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/options/downloaderoption"
)

var _ = Describe("Downloader Option Test Environment", func() {
	var o *me.Option
	var fs *pflag.FlagSet

	BeforeEach(func() {
		o = me.New(ocm.DefaultContext())
		fs = &pflag.FlagSet{}
		o.AddFlags(fs)
	})

	It("handles all parts", func() {
		MustBeSuccessful(fs.Parse([]string{"--downloader", `bla/blub:a:b:10={"k":"v"}`}))
		MustBeSuccessful(o.Configure(nil))
		Expect(len(o.Registrations)).To(Equal(1))
		Expect(o.Registrations[0].Prio).To(Equal(generics.Pointer(10)))
		Expect(o.Registrations[0].Name).To(Equal("bla/blub"))
		Expect(o.Registrations[0].ArtifactType).To(Equal("a"))
		Expect(o.Registrations[0].MediaType).To(Equal("b"))
		Expect(o.Registrations[0].Config).To(Equal([]byte(`{"k":"v"}`)))
		Expect("").To(Equal(""))
	})

	It("handles empty parts", func() {
		MustBeSuccessful(fs.Parse([]string{"--downloader", `bla/blub:::10={"k":"v"}`}))
		MustBeSuccessful(o.Configure(nil))
		Expect(len(o.Registrations)).To(Equal(1))
		Expect(o.Registrations[0].Prio).To(Equal(generics.Pointer(10)))
		Expect(o.Registrations[0].Name).To(Equal("bla/blub"))
		Expect(o.Registrations[0].ArtifactType).To(Equal(""))
		Expect(o.Registrations[0].MediaType).To(Equal(""))
		Expect(o.Registrations[0].Config).To(Equal([]byte(`{"k":"v"}`)))
	})

	It("handles art/media/config", func() {
		MustBeSuccessful(fs.Parse([]string{"--downloader", `bla/blub:a:b={"k":"v"}`}))
		MustBeSuccessful(o.Configure(nil))
		Expect(len(o.Registrations)).To(Equal(1))
		Expect(o.Registrations[0].Prio).To(BeNil())
		Expect(o.Registrations[0].Name).To(Equal("bla/blub"))
		Expect(o.Registrations[0].ArtifactType).To(Equal("a"))
		Expect(o.Registrations[0].MediaType).To(Equal("b"))
		Expect(o.Registrations[0].Config).To(Equal([]byte(`{"k":"v"}`)))
	})

	It("handles art/media/empty config", func() {
		MustBeSuccessful(fs.Parse([]string{"--downloader", `bla/blub:a:b`}))
		MustBeSuccessful(o.Configure(nil))
		Expect(len(o.Registrations)).To(Equal(1))
		Expect(o.Registrations[0].Prio).To(BeNil())
		Expect(o.Registrations[0].Name).To(Equal("bla/blub"))
		Expect(o.Registrations[0].ArtifactType).To(Equal("a"))
		Expect(o.Registrations[0].MediaType).To(Equal("b"))
		Expect(o.Registrations[0].Config).To(BeNil())
	})

	It("handles empty config", func() {
		MustBeSuccessful(fs.Parse([]string{"--downloader", `bla/blub`}))
		MustBeSuccessful(o.Configure(nil))
		Expect(len(o.Registrations)).To(Equal(1))
		Expect(o.Registrations[0].Prio).To(BeNil())
		Expect(o.Registrations[0].Name).To(Equal("bla/blub"))
		Expect(o.Registrations[0].ArtifactType).To(Equal(""))
		Expect(o.Registrations[0].MediaType).To(Equal(""))
		Expect(o.Registrations[0].Config).To(BeNil())
	})
})
