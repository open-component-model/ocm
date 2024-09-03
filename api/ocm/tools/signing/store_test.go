package signing_test

import (
	"os"

	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"ocm.software/ocm/api/ocm/compdesc"
	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/compdesc/versions/ocm.software/v3alpha1"
	v2 "ocm.software/ocm/api/ocm/compdesc/versions/v2"
	"ocm.software/ocm/api/ocm/tools/signing"
	common "ocm.software/ocm/api/utils/misc"
	"ocm.software/ocm/api/utils/runtime"
)

var _ = Describe("Store", func() {
	Context("persistence", func() {
		It("", func() {
			file := Must(os.CreateTemp("", "store-*"))

			os.Remove(file.Name())
			store := Must(signing.NewVerifiedStore(file.Name()))
			Expect(store).NotTo(BeNil())

			cd1 := compdesc.DefaultComponent(&compdesc.ComponentDescriptor{
				Metadata: compdesc.Metadata{
					ConfiguredVersion: v3alpha1.SchemaVersion,
				},
				ComponentSpec: compdesc.ComponentSpec{
					ObjectMeta: metav1.ObjectMeta{
						Name:    COMPONENTA,
						Version: VERSION,
						Provider: metav1.Provider{
							Name: "acme.org",
						},
					},
				},
			})
			cd2 := compdesc.DefaultComponent(&compdesc.ComponentDescriptor{
				Metadata: compdesc.Metadata{
					ConfiguredVersion: v2.SchemaVersion,
				},
				ComponentSpec: compdesc.ComponentSpec{
					ObjectMeta: metav1.ObjectMeta{
						Name:    COMPONENTB,
						Version: VERSION,
						Provider: metav1.Provider{
							Name: "acme.org",
						},
					},
				},
			})

			store.Add(cd1, "a")
			store.Add(cd2, "b")
			store.Add(cd2, "c")

			MustBeSuccessful(store.Save())

			desc := signing.StorageDescriptor{
				ComponentVersions: map[string]*signing.StorageEntry{
					common.VersionedElementKey(cd1).String(): {
						Signatures: []string{"a"},
						Descriptor: (*compdesc.GenericComponentDescriptor)(cd1),
					},
					common.VersionedElementKey(cd2).String(): {
						Signatures: []string{"b", "c"},
						Descriptor: (*compdesc.GenericComponentDescriptor)(cd2),
					},
				},
			}

			data := Must(os.ReadFile(file.Name()))

			exp := Must(runtime.DefaultYAMLEncoding.Marshal(desc))
			Expect(data).To(YAMLEqual(exp))

			store = Must(signing.NewVerifiedStore(file.Name()))

			Expect(store.Get(cd1)).To(YAMLEqual(cd1))
			Expect(store.Get(cd1).Equal(cd1)).To(BeTrue())

			Expect(store.Get(cd2)).To(YAMLEqual(cd2))
			Expect(store.Get(cd2).Equal(cd2)).To(BeTrue())
		})
	})
})
