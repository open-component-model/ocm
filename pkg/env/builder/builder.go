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
	"github.com/gardener/ocm/pkg/common/accessio"
	"github.com/gardener/ocm/pkg/env"
	"github.com/gardener/ocm/pkg/oci"
	"github.com/gardener/ocm/pkg/ocm/compdesc"
	"github.com/gardener/ocm/pkg/ocm/cpi"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type element interface {
	SetBuilder(b *Builder)
	Type() string
	Close() error
	Set()
}

type State struct {
}
type base struct {
	*Builder
}

func (e *base) SetBuilder(b *Builder) {
	e.Builder = b
}

type Builder struct {
	*env.Environment
	stack []element

	ocm_repo cpi.Repository
	ocm_comp cpi.ComponentAccess
	ocm_vers cpi.ComponentVersionAccess
	ocm_rsc  *compdesc.ResourceMeta
	ocm_src  *compdesc.SourceMeta
	ocm_meta *compdesc.ElementMeta
	ocm_acc  *compdesc.AccessSpec

	blob *accessio.BlobAccess

	oci_repo   oci.Repository
	oci_nsacc  oci.NamespaceAccess
	oci_artacc oci.ArtefactAccess
	oci_tags   *[]string
}

func NewBuilder(t *env.Environment) *Builder {
	return &Builder{Environment: t}
}

func (b *Builder) require(typ string) {
	Expect(b.peek().Type()).To(Equal(typ))
}

func (b *Builder) set() {
	b.ocm_repo = nil
	b.ocm_comp = nil
	b.ocm_vers = nil
	b.ocm_rsc = nil
	b.ocm_src = nil
	b.ocm_meta = nil
	b.ocm_acc = nil

	b.blob = nil

	b.oci_repo = nil
	b.oci_nsacc = nil
	b.oci_artacc = nil
	b.oci_tags = nil

	if len(b.stack) > 0 {
		b.peek().Set()
	}

}

func (b *Builder) expect(p interface{}, msg string, tests ...func() bool) {
	if p == nil {
		Fail(msg+" required", 2)
	}
	for _, f := range tests {
		if !f() {
			Fail(msg+" required", 2)
		}
	}
}

func (b *Builder) failOn(err error, callerSkip ...int) {
	if err != nil {
		skip := 2
		if len(callerSkip) > 0 {
			skip = callerSkip[0]
		}
		Fail(err.Error(), skip)
	}
}

func (b *Builder) peek() element {
	Expect(len(b.stack) > 0).To(BeTrue())
	return b.stack[len(b.stack)-1]
}

func (b *Builder) pop() element {
	Expect(len(b.stack) > 0).To(BeTrue())
	e := b.stack[len(b.stack)-1]
	b.stack = b.stack[:len(b.stack)-1]
	b.set()
	return e
}

func (b *Builder) push(e element) {
	b.stack = append(b.stack, e)
	b.set()
}

func (b *Builder) configure(e element, funcs []func()) {
	e.SetBuilder(b)
	b.push(e)
	for _, f := range funcs {
		if f != nil {
			f()
		}
	}
	err := b.pop().Close()
	if err != nil {
		Fail(err.Error(), 2)
	}
}

////////////////////////////////////////////////////////////////////////////////

func (b *Builder) BlobStringData(mime string, data string) {
	b.expect(b.blob, T_OCMACCESS)
	if b.ocm_acc != nil && *b.ocm_acc != nil {
		Fail("access already set", 1)
	}
	*(b.blob) = accessio.BlobAccessForData(mime, []byte(data))
}
