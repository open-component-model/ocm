// Copyright 2022 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

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
