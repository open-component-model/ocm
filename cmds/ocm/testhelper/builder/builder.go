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
	"github.com/gardener/ocm/cmds/ocm/testhelper"
	"github.com/gardener/ocm/pkg/common/accessio"
	"github.com/gardener/ocm/pkg/common/accessobj"
	"github.com/gardener/ocm/pkg/errors"
	"github.com/gardener/ocm/pkg/ocm/compdesc"
	metav1 "github.com/gardener/ocm/pkg/ocm/compdesc/meta/v1"
	"github.com/gardener/ocm/pkg/ocm/cpi"
	"github.com/gardener/ocm/pkg/ocm/repositories/comparch/comparch"
	"github.com/gardener/ocm/pkg/ocm/repositories/ctf"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type element interface {
	SetBuilder(b *Builder)
	Type() string
	Close() error
	Set()
}

type base struct {
	*Builder
}

func (e *base) SetBuilder(b *Builder) {
	e.Builder = b
}

type Builder struct {
	*testhelper.TestEnv
	stack []element
	repo  cpi.Repository
	comp  cpi.ComponentAccess
	vers  cpi.ComponentVersionAccess
	rsc   *compdesc.ResourceMeta
	src   *compdesc.SourceMeta
	acc   *compdesc.AccessSpec
	blob  *accessio.BlobAccess
}

func NewBuilder(t *testhelper.TestEnv) *Builder {
	return &Builder{TestEnv: t}
}

func (b *Builder) require(typ string) {
	Expect(b.peek().Type()).To(Equal(typ))
}

func (b *Builder) set() {
	b.repo = nil
	b.comp = nil
	b.vers = nil
	b.rsc = nil
	b.src = nil
	b.acc = nil
	b.blob = nil
	if len(b.stack) > 0 {
		b.peek().Set()
	}
}

func (b *Builder) expect(p interface{}, msg string) {
	if p == nil {
		Fail(msg+" required", 2)
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
	b.pop().Close()
}

////////////////////////////////////////////////////////////////////////////////

const T_REPOSITORY = "repository"

type repository struct {
	base
	kind string
	cpi.Repository
}

func (r *repository) Type() string {
	if r.kind != "" {
		return r.kind
	}
	return T_REPOSITORY
}

func (r *repository) Set() {
	r.Builder.repo = r.Repository
}

const T_COMPONENT = "component"

type component struct {
	base
	kind string
	cpi.ComponentAccess
}

func (r *component) Type() string {
	if r.kind != "" {
		return r.kind
	}
	return T_COMPONENT
}

func (r *component) Set() {
	r.Builder.comp = r.ComponentAccess
}

const T_VERSION = "component version"

type version struct {
	base
	kind string
	cpi.ComponentVersionAccess
}

func (r *version) Type() string {
	if r.kind != "" {
		return r.kind
	}
	return T_VERSION
}

func (r *version) Set() {
	r.Builder.vers = r.ComponentVersionAccess
}

////////////////////////////////////////////////////////////////////////////////

const T_COMPARCH = "component archive"

func (b *Builder) ComponentArchive(path string, fmt accessio.FileFormat, name, vers string, f ...func()) {
	r, err := comparch.Open(b.OCMContext(), accessobj.ACC_WRITABLE|accessobj.ACC_CREATE, path, 0777)
	b.failOn(err)
	r.SetName(name)
	r.SetVersion(vers)
	r.GetDescriptor().Provider = metav1.ProviderType("ACME")

	b.configure(&version{ComponentVersionAccess: r, kind: T_COMPARCH}, f)
}

const T_CTF = "common transport format"

func (b *Builder) CommonTransport(path string, fmt accessio.FileFormat, f ...func()) {
	r, err := ctf.Open(b.OCMContext(), accessobj.ACC_WRITABLE|accessobj.ACC_CREATE, path, 0777)
	b.failOn(err)
	b.configure(&repository{Repository: r, kind: T_CTF}, f)
}

func (b *Builder) Component(name string, f ...func()) {
	b.expect(b.repo, T_REPOSITORY)
	c, err := b.repo.LookupComponent(name)
	b.failOn(err)
	b.configure(&component{ComponentAccess: c}, f)
}

func (b *Builder) Version(name string, f ...func()) {
	b.expect(b.comp, T_COMPONENT)
	v, err := b.comp.LookupVersion(name)
	if err != nil {
		if errors.IsErrNotFound(err) {
			v, err = b.comp.NewVersion(name)
		}
	}
	b.failOn(err)
	v.GetDescriptor().Provider = metav1.ProviderType("ACME")
	b.configure(&version{ComponentVersionAccess: v}, f)
}

func (b *Builder) Provider(name string) {
	b.expect(b.vers, T_VERSION)
	b.vers.GetDescriptor().Provider = metav1.ProviderType(name)
}

////////////////////////////////////////////////////////////////////////////////

type resource struct {
	base

	meta   compdesc.ResourceMeta
	access compdesc.AccessSpec
	blob   accessio.BlobAccess
}

const T_RESOURCE = "resource"

func (r *resource) Type() string {
	return T_RESOURCE
}

func (r *resource) Set() {
	r.Builder.rsc = &r.meta
	r.Builder.acc = &r.access
	r.Builder.blob = &r.blob
}

func (r *resource) Close() error {
	switch {
	case r.acc != nil:
		Expect(r.Builder.vers.SetResource(&r.meta, r.access)).To(Succeed())
	case r.blob != nil:
		Expect(r.Builder.vers.SetResourceBlob(&r.meta, r.blob, "", nil)).To(Succeed())
	default:
		Fail("access or blob", 3)
	}
	return nil
}

func (b *Builder) Resource(name, vers, typ string, relation metav1.ResourceRelation, f ...func()) {
	b.expect(b.vers, T_VERSION)
	r := &resource{}
	r.meta.Name = name
	r.meta.Version = vers
	r.meta.Type = typ
	r.meta.Relation = relation
	b.configure(r, f)
}

type source struct {
	base

	meta   compdesc.SourceMeta
	access compdesc.AccessSpec
	blob   accessio.BlobAccess
}

const SOURCE = "source"

func (r *source) Type() string {
	return SOURCE
}

func (r *source) Set() {
	r.Builder.src = &r.meta
	r.Builder.acc = &r.access
	r.Builder.blob = &r.blob
}

func (r *source) Close() error {
	switch {
	case r.acc != nil:
		Expect(r.Builder.vers.SetSource(&r.meta, r.access)).To(Succeed())
	case r.blob != nil:
		Expect(r.Builder.vers.SetSourceBlob(&r.meta, r.blob, "", nil)).To(Succeed())
	default:
		Fail("access or blob", 3)
	}
	return nil
}

func (b *Builder) Source(name, vers, typ string, f ...func()) {
	b.expect(b.vers, T_VERSION)
	r := &source{}
	r.meta.Name = name
	r.meta.Type = typ
	r.meta.Version = vers
	b.configure(r, f)
}

func (b *Builder) BlobStringData(mime string, data string) {
	b.expect(b.blob, ACCESS)
	if b.acc != nil && *b.acc != nil {
		Fail("access already set", 1)
	}
	*(b.blob) = accessio.BlobAccessForData(mime, []byte(data))
}

func (b *Builder) Access(acc compdesc.AccessSpec) {
	b.expect(b.acc, ACCESS)
	if b.blob != nil && *b.blob != nil {
		Fail("access already set", 1)
	}

	*(b.acc) = acc
}

////////////////////////////////////////////////////////////////////////////////

const ACCESS = "access"

type access struct {
	access cpi.AccessSpec
}

func (r *access) Type() string {
	return ACCESS
}
