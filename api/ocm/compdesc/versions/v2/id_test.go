package v2_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	me "ocm.software/ocm/api/ocm/compdesc/versions/v2"
)

var _ = Describe("Extra Identity Test Environment", func() {
	It("handles complete defaulting", func() {
		resources := me.Resources{
			me.Resource{
				ElementMeta: me.ElementMeta{
					Name:    "res",
					Version: "v1",
				},
			},
			me.Resource{
				ElementMeta: me.ElementMeta{
					Name:    "res",
					Version: "v2",
				},
			},
		}
		Expect(resources[0].GetIdentity(resources)).To(Equal(metav1.NewIdentity("res", me.SystemIdentityVersion, "v1")))
		Expect(resources[1].GetIdentity(resources)).To(Equal(metav1.NewIdentity("res", me.SystemIdentityVersion, "v2")))
	})

	It("handles partial defaulting", func() {
		resources := me.Resources{
			me.Resource{
				ElementMeta: me.ElementMeta{
					Name:          "res",
					Version:       "v1",
					ExtraIdentity: metav1.NewExtraIdentity(me.SystemIdentityVersion, "v1"),
				},
			},
			me.Resource{
				ElementMeta: me.ElementMeta{
					Name:    "res",
					Version: "v2",
				},
			},
		}
		Expect(resources[0].GetIdentity(resources)).To(Equal(metav1.NewIdentity("res", me.SystemIdentityVersion, "v1")))
		Expect(resources[1].GetIdentity(resources)).To(Equal(metav1.NewIdentity("res", me.SystemIdentityVersion, "v2")))
	})
})
