// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package testhelper

import (
	. "github.com/onsi/gomega"

	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/equivalent"
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
