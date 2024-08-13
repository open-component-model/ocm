package signing_test

import (
	"os"

	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/compdesc/versions/ocm.software/v3alpha1"
	v2 "ocm.software/ocm/api/ocm/compdesc/versions/v2"
	common "ocm.software/ocm/api/utils/misc"

	"ocm.software/ocm/api/ocm/compdesc"
	"ocm.software/ocm/api/ocm/tools/signing"
)

var _ = Describe("Store", func() {
	Context("persistence", func() {
		It("", func() {
			file := Must(os.CreateTemp("", "store-*"))

			os.Remove(file.Name())
			store := Must(signing.NewVerifiedStore(file.Name()))
			Expect(store).NotTo(BeNil())

			cd1 := &compdesc.ComponentDescriptor{
				Metadata: compdesc.Metadata{
					ConfiguredVersion: v3alpha1.SchemaVersion,
				},
				ComponentSpec: compdesc.ComponentSpec{
					ObjectMeta: metav1.ObjectMeta{
						Name:    COMPONENTA,
						Version: VERSION,
					},
				},
			}
			cd2 := &compdesc.ComponentDescriptor{
				Metadata: compdesc.Metadata{
					ConfiguredVersion: v2.SchemaVersion,
				},
				ComponentSpec: compdesc.ComponentSpec{
					ObjectMeta: metav1.ObjectMeta{
						Name:    COMPONENTA,
						Version: VERSION,
					},
				},
			}

			store.Add(cd1, "a")
			store.Add(cd2, "b")
			store.Add(cd2, "c")

			MustBeSuccessful(store.Save())

			desc := signing.StorageDescriptor{
				ComponentVersions: map[string]*signing.StorageEntry{
					common.VersionedElementKey(cd1).String(): &signing.StorageEntry{
						Signatures: []string{"a"},
						Descriptor: (*compdesc.GenericComponentDescriptor)(cd1),
					},
					common.VersionedElementKey(cd2).String(): &signing.StorageEntry{
						Signatures: []string{"b", "c"},
						Descriptor: (*compdesc.GenericComponentDescriptor)(cd2),
					},
				},
			}

			data := Must(os.ReadFile(file.Name()))

			Expect(data).To(YAMLEqual(desc))

			store = Must(signing.NewVerifiedStore(file.Name()))

			Expect(store.Get(cd1)).To(YAMLEqual(cd1))
			Expect(store.Get(cd1).Equal(cd1)).To(BeTrue())

			Expect(store.Get(cd2)).To(YAMLEqual(cd2))
			Expect(store.Get(cd2).Equal(cd2)).To(BeTrue())

		})
	})
})
