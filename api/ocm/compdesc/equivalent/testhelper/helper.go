package testhelper

import (
	. "github.com/onsi/gomega"

	"ocm.software/ocm/api/ocm/compdesc/equivalent"
)

func CheckEquivalent(eq equivalent.EqualState) {
	ExpectWithOffset(1, eq).To(Equal(equivalent.StateEquivalent()))

	Expect(eq.IsEquivalent()).To(BeTrue())
	Expect(eq.IsHashEqual()).To(BeTrue())
	Expect(eq.IsLocalHashEqual()).To(BeTrue())
	Expect(eq.IsArtifactEqual()).To(BeTrue())
	Expect(eq.IsArtifactDetectable()).To(BeTrue())
}

func CheckNotEquivalent(eq equivalent.EqualState) {
	ExpectWithOffset(1, eq).To(Equal(equivalent.StateNotEquivalent()))

	Expect(eq.IsEquivalent()).To(BeFalse())
	Expect(eq.IsHashEqual()).To(BeTrue())
	Expect(eq.IsLocalHashEqual()).To(BeTrue())
	Expect(eq.IsArtifactEqual()).To(BeTrue())
	Expect(eq.IsArtifactDetectable()).To(BeTrue())
}

func CheckNotLocalHashEqual(eq equivalent.EqualState) {
	ExpectWithOffset(1, eq).To(Equal(equivalent.StateNotLocalHashEqual()))

	Expect(eq.IsEquivalent()).To(BeFalse())
	Expect(eq.IsHashEqual()).To(BeFalse())
	Expect(eq.IsLocalHashEqual()).To(BeFalse())
	Expect(eq.IsArtifactEqual()).To(BeTrue())
	Expect(eq.IsArtifactDetectable()).To(BeTrue())
}

func CheckNotDetectable(eq equivalent.EqualState) {
	ExpectWithOffset(1, eq).To(Equal(equivalent.StateNotArtifactEqual(false)))

	Expect(eq.IsEquivalent()).To(BeFalse())
	Expect(eq.IsHashEqual()).To(BeFalse())
	Expect(eq.IsLocalHashEqual()).To(BeTrue())
	Expect(eq.IsArtifactEqual()).To(BeFalse())
	Expect(eq.IsArtifactDetectable()).To(BeFalse())
}

func CheckNotArtifactEqual(eq equivalent.EqualState) {
	ExpectWithOffset(1, eq).To(Equal(equivalent.StateNotArtifactEqual(true)))

	Expect(eq.IsEquivalent()).To(BeFalse())
	Expect(eq.IsHashEqual()).To(BeFalse())
	Expect(eq.IsLocalHashEqual()).To(BeTrue())
	Expect(eq.IsArtifactEqual()).To(BeFalse())
	Expect(eq.IsArtifactDetectable()).To(BeTrue())
}
