package compdesc_test

import (
	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"ocm.software/ocm/api/ocm/compdesc"
	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
)

const (
	VERSION    = "v1"
	COMPONENT  = "github.com/mandelsoft/test"
	COMPONENT2 = "github.com/mandelsoft/test2"
	COMPONENT3 = "github.com/mandelsoft/test3"

	EXTRA_ATTR = "platform"
	EXTRA_VAL  = "test"
)

func CheckResourceRef(cv *compdesc.ComponentDescriptor, resolver compdesc.ComponentVersionResolver, comp string, rsc metav1.Identity, path ...metav1.Identity) {
	ref := metav1.NewNestedResourceRef(rsc, path)

	r, cd := Must2(compdesc.ResolveResourceReference(cv, ref, resolver))
	ExpectWithOffset(1, r).NotTo(BeNil())
	ExpectWithOffset(1, cd).NotTo(BeNil())

	ExpectWithOffset(1, cd.Name).To(Equal(comp))
	ExpectWithOffset(1, r.Name).To(Equal(rsc.Get(compdesc.SystemIdentityName)))
	ExpectWithOffset(1, r.ExtraIdentity).To(Equal(rsc.ExtraIdentity()))
}

var _ = Describe("resolving local resource references", func() {
	var set *compdesc.ComponentVersionSet

	BeforeEach(func() {
		set = compdesc.NewComponentVersionSet()

		cd := compdesc.New(COMPONENT, VERSION)
		cd.Resources = append(cd.Resources, compdesc.Resource{
			ResourceMeta: compdesc.ResourceMeta{
				ElementMeta: compdesc.ElementMeta{
					Name:    "testata",
					Version: "",
				},
				Type:     "PlainText",
				Relation: metav1.LocalRelation,
			},
		})
		set.AddVersion(cd)

		cd = compdesc.New(COMPONENT2, VERSION)
		cd.Resources = append(cd.Resources, compdesc.Resource{
			ResourceMeta: compdesc.ResourceMeta{
				ElementMeta: compdesc.ElementMeta{
					Name:    "otherdata",
					Version: "",
				},
				Type:     "PlainText",
				Relation: metav1.LocalRelation,
			},
		})
		cd.Resources = append(cd.Resources, compdesc.Resource{
			ResourceMeta: compdesc.ResourceMeta{
				ElementMeta: compdesc.ElementMeta{
					Name:          "otherdata",
					Version:       "",
					ExtraIdentity: metav1.NewExtraIdentity(EXTRA_ATTR, EXTRA_VAL),
				},
				Type:     "PlainText",
				Relation: metav1.LocalRelation,
			},
		})
		cd.References = append(cd.References, compdesc.Reference{
			ElementMeta: compdesc.ElementMeta{
				Name:    "ref",
				Version: VERSION,
			},
			ComponentName: COMPONENT,
		})
		set.AddVersion(cd)

		cd = compdesc.New(COMPONENT3, VERSION)
		cd.Resources = append(cd.Resources, compdesc.Resource{
			ResourceMeta: compdesc.ResourceMeta{
				ElementMeta: compdesc.ElementMeta{
					Name:    "topdata",
					Version: "",
				},
				Type:     "PlainText",
				Relation: metav1.LocalRelation,
			},
		})
		cd.References = append(cd.References, compdesc.Reference{
			ElementMeta: compdesc.ElementMeta{
				Name:    "nested",
				Version: VERSION,
			},
			ComponentName: COMPONENT2,
		})
		cd.References = append(cd.References, compdesc.Reference{
			ElementMeta: compdesc.ElementMeta{
				Name:          "nested",
				Version:       VERSION,
				ExtraIdentity: metav1.NewExtraIdentity(EXTRA_ATTR, EXTRA_VAL),
			},
			ComponentName: COMPONENT2,
		})
		set.AddVersion(cd)
	})

	It("resolves a direct local resource", func() {
		CheckResourceRef(Must(set.LookupComponentVersion(COMPONENT3, VERSION)), set, COMPONENT3, metav1.NewIdentity("topdata"))
	})

	It("resolves an indirect resource", func() {
		CheckResourceRef(Must(set.LookupComponentVersion(COMPONENT3, VERSION)), set, COMPONENT2, metav1.NewIdentity("otherdata"), metav1.NewIdentity("nested"))
	})
	It("resolves an indirect resource with extra id", func() {
		CheckResourceRef(Must(set.LookupComponentVersion(COMPONENT3, VERSION)), set, COMPONENT2, metav1.NewIdentity("otherdata", EXTRA_ATTR, EXTRA_VAL), metav1.NewIdentity("nested"))
	})

	It("fails resolving an indirect resource with non existing extra id", func() {
		ref := metav1.NewNestedResourceRef(metav1.NewIdentity("otherdata", EXTRA_ATTR, "dummy"), []metav1.Identity{metav1.NewIdentity("nested")})
		ExpectError(compdesc.ResolveResourceReference(Must(set.LookupComponentVersion(COMPONENT3, VERSION)), ref, set)).To(
			MatchError("not found"))
	})

	It("skips an intermediate component version", func() {
		CheckResourceRef(Must(set.LookupComponentVersion(COMPONENT3, VERSION)), set, COMPONENT, metav1.NewIdentity("testata"), metav1.NewIdentity("nested"), metav1.NewIdentity("ref"))
	})

	It("skips an intermediate component version with extra id", func() {
		CheckResourceRef(Must(set.LookupComponentVersion(COMPONENT3, VERSION)), set, COMPONENT, metav1.NewIdentity("testata"), metav1.NewIdentity("nested", EXTRA_ATTR, EXTRA_VAL), metav1.NewIdentity("ref"))
	})

	It("fails resolving an indirect resource with non existing intermediate ref", func() {
		ref := metav1.NewNestedResourceRef(metav1.NewIdentity("testdata"), []metav1.Identity{metav1.NewIdentity("nested", EXTRA_ATTR, "dummy")})
		ExpectError(compdesc.ResolveResourceReference(Must(set.LookupComponentVersion(COMPONENT3, VERSION)), ref, set)).To(
			MatchError("github.com/mandelsoft/test3:v1: component reference \"\"name\"=\"nested\",\"platform\"=\"dummy\"\" not found"))
	})
})
