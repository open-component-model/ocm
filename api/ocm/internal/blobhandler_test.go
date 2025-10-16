package internal_test

import (
	"fmt"

	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"ocm.software/ocm/api/ocm/internal"
	"ocm.software/ocm/api/utils/mime"
	"ocm.software/ocm/api/utils/registrations"
)

const REPO = "repo"

var (
	IMPL = internal.ImplementationRepositoryType{internal.CONTEXT_TYPE, REPO}
	ART  = "myType"
)

type BlobHandler struct {
	name string
}

var _ internal.BlobHandler = (*BlobHandler)(nil)

func (b BlobHandler) StoreBlob(blob internal.BlobAccess, artType string, hint string, global internal.AccessSpec, ctx internal.StorageContext) (internal.AccessSpec, error) {
	return nil, fmt.Errorf(b.name)
}

type TestRegistrationHandler struct {
	name       string
	registered map[string]interface{}
}

func NewTestRegistrationHandler(name string) *TestRegistrationHandler {
	return &TestRegistrationHandler{
		name:       name,
		registered: map[string]interface{}{},
	}
}

func (t *TestRegistrationHandler) RegisterByName(handler string, ctx internal.Context, config internal.BlobHandlerConfig, opts ...internal.BlobHandlerOption) (bool, error) {
	path := registrations.NewNamePath(handler)
	if len(path) < 1 || path[0] != "match" {
		return false, nil
	}
	t.registered[handler] = nil
	return true, nil
}

func (t *TestRegistrationHandler) GetHandlers(ctx internal.Context) registrations.HandlerInfos {
	return nil
}

var _ = Describe("blob handler registry test", func() {
	Context("registration registry", func() {
		var reg internal.BlobHandlerRegistrationRegistry

		var ha *TestRegistrationHandler
		var hab *TestRegistrationHandler
		var habc *TestRegistrationHandler
		var habd *TestRegistrationHandler
		var habe *TestRegistrationHandler
		var hb *TestRegistrationHandler

		BeforeEach(func() {
			reg = internal.NewBlobHandlerRegistrationRegistry()
			ha = NewTestRegistrationHandler("a")
			hab = NewTestRegistrationHandler("a/b")
			habc = NewTestRegistrationHandler("a/b/c")
			habd = NewTestRegistrationHandler("a/b/d")
			habe = NewTestRegistrationHandler("a/b/e")
			hb = NewTestRegistrationHandler("b")
		})

		It("registers ordered prefixes", func() {
			reg.RegisterRegistrationHandler("a", ha)
			reg.RegisterRegistrationHandler("a/b/d", habd)
			reg.RegisterRegistrationHandler("a/b/c", habc)
			reg.RegisterRegistrationHandler("a/b/e", habe)
			reg.RegisterRegistrationHandler("a/b", hab)
			reg.RegisterRegistrationHandler("b", hb)

			Expect(reg.GetRegistrationHandlers("a/b/c")).To(Equal([]*internal.RegistrationHandlerInfo{
				internal.NewRegistrationHandlerInfo("a/b/c", habc),
				internal.NewRegistrationHandlerInfo("a/b", hab),
				internal.NewRegistrationHandlerInfo("a", ha),
			}))
		})

		It("registers ordered prefixes", func() {
			reg.RegisterRegistrationHandler("a", ha)
			reg.RegisterRegistrationHandler("a/b/d", habd)
			reg.RegisterRegistrationHandler("a/b/c", habc)
			reg.RegisterRegistrationHandler("a/b/e", habe)
			reg.RegisterRegistrationHandler("a/b", hab)
			reg.RegisterRegistrationHandler("b", hb)

			_, err := reg.RegisterByName("a/b/c/d", nil, nil)
			MustFailWithMessage(err, "no registration handler found for a/b/c/d")
			Expect(Must(reg.RegisterByName("a/b/c/match/d", nil, nil))).To(BeTrue())
			Expect(Must(reg.RegisterByName("a/b/c/match", nil, nil))).To(BeTrue())

			Expect(ha.registered).To(Equal(map[string]interface{}{}))
			Expect(hb.registered).To(Equal(map[string]interface{}{}))
			Expect(hab.registered).To(Equal(map[string]interface{}{}))
			Expect(habd.registered).To(Equal(map[string]interface{}{}))
			Expect(habe.registered).To(Equal(map[string]interface{}{}))
			Expect(habc.registered).To(Equal(map[string]interface{}{"match/d": nil, "match": nil}))
		})

		It("registers ordered prefixes", func() {
			reg.RegisterRegistrationHandler("a/b/d", habd)
			reg.RegisterRegistrationHandler("a/b/c", habc)
			reg.RegisterRegistrationHandler("a/b/e", habe)
			reg.RegisterRegistrationHandler("a", ha)

			derived := internal.NewBlobHandlerRegistrationRegistry(reg)
			derived.RegisterRegistrationHandler("a/b", hab)
			derived.RegisterRegistrationHandler("b", hb)

			list := derived.GetRegistrationHandlers("a/b/e")
			Expect(list).To(Equal([]*internal.RegistrationHandlerInfo{
				internal.NewRegistrationHandlerInfo("a/b/e", habe),
				internal.NewRegistrationHandlerInfo("a/b", hab),
				internal.NewRegistrationHandlerInfo("a", ha),
			}))

			_, err := reg.RegisterByName("a/b/e/d", nil, nil)
			MustFailWithMessage(err, "no registration handler found for a/b/e/d")
			Expect(Must(reg.RegisterByName("a/b/e/match/d", nil, nil))).To(BeTrue())
			Expect(Must(reg.RegisterByName("a/b/e/match", nil, nil))).To(BeTrue())

			Expect(ha.registered).To(Equal(map[string]interface{}{}))
			Expect(hb.registered).To(Equal(map[string]interface{}{}))
			Expect(hab.registered).To(Equal(map[string]interface{}{}))
			Expect(habd.registered).To(Equal(map[string]interface{}{}))
			Expect(habc.registered).To(Equal(map[string]interface{}{}))
			Expect(habe.registered).To(Equal(map[string]interface{}{"match/d": nil, "match": nil}))
		})
	})

	////////////////////////////////////////////////////////////////////////////

	Context("blob handler registry", func() {
		var reg internal.BlobHandlerRegistry
		var ext internal.BlobHandlerRegistry

		BeforeEach(func() {
			reg = internal.NewBlobHandlerRegistry()
			ext = internal.NewBlobHandlerRegistry(reg)
		})

		DescribeTable("prioritizes complete specs",
			func(eff *internal.BlobHandlerRegistry) {
				reg.Register(&BlobHandler{"mine"}, internal.ForMimeType(mime.MIME_TEXT))
				reg.Register(&BlobHandler{"repo"}, internal.ForRepo(internal.CONTEXT_TYPE, REPO))
				reg.Register(&BlobHandler{"art"}, internal.ForArtifactType(ART))
				reg.Register(&BlobHandler{"all"}, internal.ForRepo(internal.CONTEXT_TYPE, REPO), internal.ForArtifactType(ART), internal.ForMimeType(mime.MIME_TEXT))
				reg.Register(&BlobHandler{"repomime"}, internal.ForRepo(internal.CONTEXT_TYPE, REPO), internal.ForMimeType(mime.MIME_TEXT))

				h := (*eff).LookupHandler(IMPL, ART, mime.MIME_TEXT)
				Expect(h).NotTo(BeNil())
				_, err := h.StoreBlob(nil, "", "", nil, nil)
				Expect(err).To(MatchError(fmt.Errorf("all")))
			},
			Entry("plain", &reg),
			Entry("extended", &ext),
		)

		DescribeTable("prioritizes mime",
			func(eff *internal.BlobHandlerRegistry) {
				reg.Register(&BlobHandler{"mine"}, internal.ForMimeType(mime.MIME_TEXT))
				reg.Register(&BlobHandler{"repo"}, internal.ForRepo(internal.CONTEXT_TYPE, REPO))
				reg.Register(&BlobHandler{"repomime"}, internal.ForRepo(internal.CONTEXT_TYPE, REPO), internal.ForMimeType(mime.MIME_TEXT))

				h := (*eff).LookupHandler(IMPL, ART, mime.MIME_TEXT)
				Expect(h).NotTo(BeNil())
				_, err := h.StoreBlob(nil, "", "", nil, nil)
				Expect(err).To(MatchError(fmt.Errorf("repomime")))
			},
			Entry("plain", &reg),
			Entry("extended", &ext),
		)

		DescribeTable("prioritizes mime",
			func(eff *internal.BlobHandlerRegistry) {
				reg.Register(&BlobHandler{"mine"}, internal.ForMimeType(mime.MIME_TEXT))
				reg.Register(&BlobHandler{"repo"}, internal.ForRepo(internal.CONTEXT_TYPE, REPO))
				reg.Register(&BlobHandler{"repomine"}, internal.ForRepo(internal.CONTEXT_TYPE, REPO), internal.ForMimeType(mime.MIME_TEXT))
				reg.Register(&BlobHandler{"repoart"}, internal.ForRepo(internal.CONTEXT_TYPE, REPO), internal.ForArtifactType(ART))

				h := (*eff).LookupHandler(IMPL, ART, mime.MIME_TEXT)
				Expect(h).NotTo(BeNil())
				_, err := h.StoreBlob(nil, "", "", nil, nil)
				Expect(err).To(MatchError(fmt.Errorf("repoart")))
			},
			Entry("plain", &reg),
			Entry("extended", &ext),
		)

		DescribeTable("prioritizes prio",
			func(eff *internal.BlobHandlerRegistry) {
				reg.Register(&BlobHandler{"mine"}, internal.ForMimeType(mime.MIME_TEXT))
				reg.Register(&BlobHandler{"repo"}, internal.ForRepo(internal.CONTEXT_TYPE, REPO))
				reg.Register(&BlobHandler{"all"}, internal.ForRepo(internal.CONTEXT_TYPE, REPO), internal.ForMimeType(mime.MIME_TEXT))
				reg.Register(&BlobHandler{"high"}, internal.WithPrio(internal.DEFAULT_BLOBHANDLER_PRIO+1))

				h := (*eff).LookupHandler(IMPL, ART, mime.MIME_TEXT)
				Expect(h).NotTo(BeNil())
				_, err := h.StoreBlob(nil, "", "", nil, nil)
				Expect(err).To(MatchError(fmt.Errorf("high")))
			},
			Entry("plain", &reg),
			Entry("extended", &ext),
		)

		DescribeTable("copies registries",
			func(eff *internal.BlobHandlerRegistry) {
				mine := &BlobHandler{"mine"}
				repo := &BlobHandler{"repo"}
				reg.Register(mine, internal.ForMimeType(mime.MIME_TEXT))
				reg.Register(repo, internal.ForRepo(internal.CONTEXT_TYPE, REPO))

				h := (*eff).LookupHandler(internal.ImplementationRepositoryType{internal.CONTEXT_TYPE, REPO}, ART, mime.MIME_OCTET)
				Expect(h).To(Equal(internal.MultiBlobHandler{repo}))

				copy := (*eff).Copy()
				new := &BlobHandler{"repo2"}
				copy.Register(new, internal.ForRepo(internal.CONTEXT_TYPE, REPO))

				h = (*eff).LookupHandler(internal.ImplementationRepositoryType{internal.CONTEXT_TYPE, REPO}, ART, mime.MIME_OCTET)
				Expect(h).To(Equal(internal.MultiBlobHandler{repo}))

				h = copy.LookupHandler(internal.ImplementationRepositoryType{internal.CONTEXT_TYPE, REPO}, ART, mime.MIME_OCTET)
				Expect(h).To(Equal(internal.MultiBlobHandler{new}))
			},
			Entry("plain", &reg),
			Entry("extended", &ext),
		)
	})
})
