// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package inputs_test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/inputs"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/inputs/types/file"
	"github.com/open-component-model/ocm/pkg/mime"
)

var _ = Describe("Blob Inputs", func() {
	scheme := inputs.DefaultInputTypeScheme
	spec := file.New("test", mime.MIME_TEXT, false)

	It("simple decode", func() {
		data, err := json.Marshal(spec)
		Expect(err).To(Succeed())

		s, err := scheme.DecodeInputSpec(data, nil)
		Expect(err).To(Succeed())
		Expect(s).To(Equal(spec))
	})

	It("generic eval", func() {
		gen, err := inputs.ToGenericInputSpec(spec)
		Expect(err).To(Succeed())

		Expect(gen.Evaluate(scheme)).To(Equal(spec))
	})

	It("generic marshal effective", func() {
		gen, err := inputs.ToGenericInputSpec(spec)
		Expect(err).To(Succeed())

		data, err := json.Marshal(gen)
		Expect(err).To(Succeed())

		s, err := scheme.DecodeInputSpec(data, nil)
		Expect(err).To(Succeed())
		Expect(s).To(Equal(spec))
	})

	It("generic marshal effective", func() {
		gen, err := inputs.ToGenericInputSpec(spec)
		Expect(err).To(Succeed())
		Expect(gen.Evaluate(scheme)).To(Equal(spec))

		data, err := json.Marshal(gen)
		Expect(err).To(Succeed())

		s, err := scheme.DecodeInputSpec(data, nil)
		Expect(err).To(Succeed())
		Expect(s).To(Equal(spec))
	})

	It("generic unmarshal", func() {
		gen := inputs.GenericInputSpec{}

		data, err := json.Marshal(spec)
		Expect(err).To(Succeed())

		Expect(json.Unmarshal(data, &gen)).To(Succeed())

		Expect(gen.Evaluate(scheme)).To(Equal(spec))
	})
})
