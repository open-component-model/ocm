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

package builder

import (
	. "github.com/onsi/gomega"
)

const T_OCMMETA = "element with metadata"

////////////////////////////////////////////////////////////////////////////////

func (b *Builder) Label(name string, value interface{}) {
	b.expect(b.ocm_meta, T_OCMMETA)

	ExpectWithOffset(1, b.ocm_meta.Labels.Set(name, value)).To(Succeed())
}

////////////////////////////////////////////////////////////////////////////////

func (b *Builder) RemoveLabel(name string) {
	b.expect(b.ocm_meta, T_OCMMETA)

	b.ocm_meta.Labels.Remove(name)
}

////////////////////////////////////////////////////////////////////////////////

func (b *Builder) ClearLabels() {
	b.expect(b.ocm_meta, T_OCMMETA)

	b.ocm_meta.Labels = nil
}
