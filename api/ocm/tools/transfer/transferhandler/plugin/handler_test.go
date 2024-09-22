package plugin_test

import (
	"os"
	"reflect"

	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/compdesc"
	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/ociartifact"
	"ocm.software/ocm/api/ocm/extensions/attrs/plugincacheattr"
	"ocm.software/ocm/api/ocm/extensions/repositories/ocireg"
	"ocm.software/ocm/api/ocm/internal"
	. "ocm.software/ocm/api/ocm/plugin/testutils"
	"ocm.software/ocm/api/ocm/tools/transfer/transferhandler"
	"ocm.software/ocm/api/utils/blobaccess"
	. "ocm.software/ocm/cmds/ocm/testhelper"

	"github.com/mandelsoft/goutils/generics"

	"ocm.software/ocm/api/ocm/tools/transfer"
	"ocm.software/ocm/api/ocm/tools/transfer/transferhandler/plugin"
)

const PLUGIN = "transferplugin"
const HANDLER = "demo"

var _ = Describe("Plugin Handler Test Environment", func() {
	Context("handler creation", func() {
		It("", func() {
			f := Must(transfer.NewTransferHandler(plugin.Plugin(PLUGIN), plugin.TransferHandler(HANDLER)))
			Expect(reflect.TypeOf(f)).To(Equal(generics.TypeOf[*plugin.Handler]()))
		})
	})

	Context("plugin execution", func() {
		var env *TestEnv
		var plugins TempPluginDir

		BeforeEach(func() {
			env = NewTestEnv(TestData())
			plugins = Must(ConfigureTestPlugins(env, "testdata/plugins"))
		})

		AfterEach(func() {
			plugins.Cleanup()
			env.Cleanup()
		})

		It("loads plugin", func() {
			registry := plugincacheattr.Get(env)
			//	Expect(registration.RegisterExtensions(env)).To(Succeed())
			p := registry.Get(PLUGIN)
			Expect(p).NotTo(BeNil())
			Expect(p.Error()).To(Equal(""))
		})

		It("answers question", func() {
			config := Must(os.ReadFile("testdata/config"))
			f := Must(transfer.NewTransferHandler(plugin.Plugin(PLUGIN), plugin.TransferHandler(HANDLER), plugin.TransferHandlerConfig(config)))

			cv := FakeCV(env.OCMContext())

			ra := &ResourceAccess{
				ElementMeta: compdesc.ElementMeta{
					Name:    "test",
					Version: "1.0.0",
				},
			}
			b := Must(f.TransferResource(cv, ociartifact.New("ghcr.io/open-component-model/test:v1"), ra))
			Expect(b).To(BeTrue())

			b = Must(f.TransferResource(cv, ociartifact.New("gcr.io/open-component-model/test:v1"), ra))
			Expect(b).To(BeFalse())
		})

		Context("registrations", func() {
			It("finds plugin handler by name", func() {
				h := Must(transferhandler.For(env).ByName(env, "plugin/transferplugin/demo"))
				Expect(reflect.TypeOf(h)).To(Equal(generics.TypeOf[*plugin.Handler]()))
			})
		})
	})

})

////////////////////////////////////////////////////////////////////////////////

type ComponentVersionAccess struct {
	cpi.DummyComponentVersionAccess
}

func FakeCV(ctx ocm.Context) ocm.ComponentVersionAccess {
	return &ComponentVersionAccess{
		cpi.DummyComponentVersionAccess{ctx},
	}
}

func (a *ComponentVersionAccess) Repository() ocm.Repository {
	r, _ := ocireg.NewRepository(a.GetContext(), "ghcr.io/component-model/ocm")
	r, _ = r.Dup()
	return r
}

func (a *ComponentVersionAccess) GetDescriptor() *compdesc.ComponentDescriptor {
	return &compdesc.ComponentDescriptor{
		Metadata: compdesc.Metadata{},
		ComponentSpec: compdesc.ComponentSpec{
			ObjectMeta: metav1.ObjectMeta{
				Name:         "",
				Version:      "",
				Labels:       nil,
				Provider:     metav1.Provider{},
				CreationTime: nil,
			},
			RepositoryContexts: nil,
			Sources:            nil,
			References:         nil,
			Resources:          nil,
		},
		Signatures:    nil,
		NestedDigests: nil,
	}
}

func (a *ComponentVersionAccess) GetProvider() *compdesc.Provider {
	return &a.GetDescriptor().Provider
}

////////////////////////////////////////////////////////////////////////////////

type ResourceAccess ocm.ResourceMeta

var _ ocm.ResourceAccess = (*ResourceAccess)(nil)

func (r *ResourceAccess) Meta() *ocm.ResourceMeta {
	return (*ocm.ResourceMeta)(r)
}

func (r ResourceAccess) GetComponentVersion() (internal.ComponentVersionAccess, error) {
	// TODO implement me
	panic("implement me")
}

func (r ResourceAccess) GetOCMContext() internal.Context {
	// TODO implement me
	panic("implement me")
}

func (r ResourceAccess) ReferenceHint() string {
	// TODO implement me
	panic("implement me")
}

func (r ResourceAccess) Access() (internal.AccessSpec, error) {
	// TODO implement me
	panic("implement me")
}

func (r ResourceAccess) AccessMethod() (internal.AccessMethod, error) {
	// TODO implement me
	panic("implement me")
}

func (r ResourceAccess) GlobalAccess() internal.AccessSpec {
	// TODO implement me
	panic("implement me")
}

func (r ResourceAccess) BlobAccess() (blobaccess.BlobAccess, error) {
	// TODO implement me
	panic("implement me")
}
