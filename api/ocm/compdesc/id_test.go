package compdesc_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"ocm.software/ocm/api/ocm/compdesc"
	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
)

var _ = Describe("Extra Identity Test Environment", func() {
	It("handles complete defaulting", func() {
		resources := compdesc.Resources{
			compdesc.Resource{
				ResourceMeta: compdesc.ResourceMeta{
					ElementMeta: compdesc.ElementMeta{
						Name:    "res",
						Version: "v1",
					},
				},
			},
			compdesc.Resource{
				ResourceMeta: compdesc.ResourceMeta{
					ElementMeta: compdesc.ElementMeta{
						Name:    "res",
						Version: "v2",
					},
				},
			},
		}
		Expect(resources[0].GetIdentity(resources)).To(Equal(metav1.NewIdentity("res", compdesc.SystemIdentityVersion, "v1")))
		Expect(resources[1].GetIdentity(resources)).To(Equal(metav1.NewIdentity("res", compdesc.SystemIdentityVersion, "v2")))
	})

	It("handles partial defaulting", func() {
		resources := compdesc.Resources{
			compdesc.Resource{
				ResourceMeta: compdesc.ResourceMeta{
					ElementMeta: compdesc.ElementMeta{
						Name:          "res",
						Version:       "v1",
						ExtraIdentity: metav1.NewExtraIdentity(compdesc.SystemIdentityVersion, "v1"),
					},
				},
			},
			compdesc.Resource{
				ResourceMeta: compdesc.ResourceMeta{
					ElementMeta: compdesc.ElementMeta{
						Name:    "res",
						Version: "v2",
					},
				},
			},
		}
		Expect(resources[0].GetIdentity(resources)).To(Equal(metav1.NewIdentity("res", compdesc.SystemIdentityVersion, "v1")))
		Expect(resources[1].GetIdentity(resources)).To(Equal(metav1.NewIdentity("res", compdesc.SystemIdentityVersion, "v2")))
	})
})
