package testhelper

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"ocm.software/ocm/api/ocm"
	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/compdesc/normalizations/jsonv1"
	"ocm.software/ocm/api/tech/signing/hasher/sha256"
)

func CompDigestSpec(d string) *metav1.DigestSpec {
	return &metav1.DigestSpec{
		HashAlgorithm:          sha256.Algorithm,
		NormalisationAlgorithm: jsonv1.Algorithm,
		Value:                  d,
	}
}

func CheckCompRef(cv ocm.ComponentVersionAccess, name string, d *metav1.DigestSpec, offsets ...int) {
	o := 1
	for _, a := range offsets {
		o += a
	}
	for _, ref := range cv.GetDescriptor().References {
		if ref.Name == name {
			ExpectWithOffset(o, ref.Digest).To(Equal(d))
			return
		}
	}
	Fail(fmt.Sprintf("ref %s not found", name), o)
}
