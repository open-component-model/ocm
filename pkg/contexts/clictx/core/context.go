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

package core

import (
	"context"
	"io"
	"reflect"

	"github.com/mandelsoft/logging"
	"github.com/mandelsoft/vfs/pkg/vfs"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/config"
	cfgcpi "github.com/open-component-model/ocm/pkg/contexts/config/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/credentials"
	"github.com/open-component-model/ocm/pkg/contexts/datacontext"
	"github.com/open-component-model/ocm/pkg/contexts/datacontext/attrs/vfsattr"
	"github.com/open-component-model/ocm/pkg/contexts/oci"
	ctfoci "github.com/open-component-model/ocm/pkg/contexts/oci/repositories/ctf"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	ctfocm "github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/ctf"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/out"
)

const CONTEXT_TYPE = "ocm.cmd.context.gardener.cloud"

type OCI interface {
	Context() oci.Context
	OpenCTF(path string) (oci.Repository, error)
}

type OCM interface {
	Context() ocm.Context
	OpenCTF(path string) (ocm.Repository, error)
}

type FileSystem struct {
	vfs.FileSystem
}

func (f *FileSystem) ApplyOption(options accessio.Options) error {
	options.SetPathFileSystem(f.FileSystem)
	return nil
}

type Context interface {
	datacontext.Context

	AttributesContext() datacontext.AttributesContext

	ConfigContext() config.Context
	CredentialsContext() credentials.Context
	OCIContext() oci.Context
	OCMContext() ocm.Context

	FileSystem() *FileSystem

	OCI() OCI
	OCM() OCM

	ApplyOption(options accessio.Options) error

	out.Context
	WithStdIO(r io.Reader, o io.Writer, e io.Writer) Context
}

var key = reflect.TypeOf(_context{})

// DefaultContext is the default context initialized by init functions.
var DefaultContext = Builder{}.New(datacontext.MODE_SHARED)

// ForContext returns the Context to use for context.Context.
// This is either an explicit context or the default context.
// The returned context incorporates the given context.
func ForContext(ctx context.Context) Context {
	c, _ := datacontext.ForContextByKey(ctx, key, DefaultContext)
	return c.(Context)
}

func DefinedForContext(ctx context.Context) (Context, bool) {
	c, ok := datacontext.ForContextByKey(ctx, key, DefaultContext)
	if c != nil {
		return c.(Context), ok
	}
	return nil, ok
}

////////////////////////////////////////////////////////////////////////////////

type _context struct {
	datacontext.Context
	updater cfgcpi.Updater

	sharedAttributes datacontext.AttributesContext

	config      config.Context
	credentials credentials.Context
	oci         *_oci
	ocm         *_ocm

	out out.Context
}

var _ Context = &_context{}

func newContext(shared datacontext.AttributesContext, ocmctx ocm.Context, outctx out.Context, fs vfs.FileSystem, logger logging.Context) Context {
	if outctx == nil {
		outctx = out.New()
	}
	if shared == nil {
		shared = ocmctx.AttributesContext()
	}
	c := &_context{
		sharedAttributes: shared,
		credentials:      ocmctx.CredentialsContext(),
		config:           ocmctx.CredentialsContext().ConfigContext(),
		out:              outctx,
	}
	c.Context = datacontext.NewContextBase(c, CONTEXT_TYPE, key, ocmctx.GetAttributes(), logger)
	c.updater = cfgcpi.NewUpdater(ocmctx.CredentialsContext().ConfigContext(), c)
	c.oci = newOCI(c, ocmctx)
	c.ocm = newOCM(c, ocmctx)
	if fs != nil {
		vfsattr.Set(c.AttributesContext(), fs)
	}
	return c
}

func (c *_context) Update() error {
	return c.updater.Update()
}

func (c *_context) AttributesContext() datacontext.AttributesContext {
	return c.sharedAttributes
}

func (c *_context) ConfigContext() config.Context {
	return c.config
}

func (c *_context) CredentialsContext() credentials.Context {
	return c.credentials
}

func (c *_context) OCIContext() oci.Context {
	return c.oci.Context()
}

func (c *_context) OCMContext() ocm.Context {
	return c.ocm.Context()
}

func (c *_context) FileSystem() *FileSystem {
	return &FileSystem{vfsattr.Get(c)}
}

func (c *_context) OCI() OCI {
	return c.oci
}

func (c *_context) OCM() OCM {
	return c.ocm
}

func (c *_context) ApplyOption(options accessio.Options) error {
	options.SetPathFileSystem(c.FileSystem())
	return nil
}

func (c *_context) StdOut() io.Writer {
	return c.out.StdOut()
}

func (c *_context) StdErr() io.Writer {
	return c.out.StdErr()
}

func (c *_context) StdIn() io.Reader {
	return c.out.StdIn()
}

func (c *_context) WithStdIO(r io.Reader, o io.Writer, e io.Writer) Context {
	return &_view{
		Context: c,
		out:     out.NewFor(out.WithStdIO(c.out, r, o, e)),
	}
}

////////////////////////////////////////////////////////////////////////////////

type _view struct {
	Context
	out out.Context
}

func (c *_view) StdOut() io.Writer {
	return c.out.StdOut()
}

func (c *_view) StdErr() io.Writer {
	return c.out.StdErr()
}

func (c *_view) StdIn() io.Reader {
	return c.out.StdIn()
}

func (c *_view) WithStdIO(r io.Reader, o io.Writer, e io.Writer) Context {
	return &_view{
		Context: c.Context,
		out:     out.NewFor(out.WithStdIO(c.out, r, o, e)),
	}
}

////////////////////////////////////////////////////////////////////////////////
// the coding for _oci and _ocm is identical except the package used:
// _oci uses oci and ctfoci
// _ocm uses ocm and ctfocm

type _oci struct {
	*_context
	ctx   oci.Context
	repos map[string]oci.RepositorySpec
}

func newOCI(ctx *_context, ocmctx ocm.Context) *_oci {
	return &_oci{
		_context: ctx,
		ctx:      ocmctx.OCIContext(),
		repos:    map[string]oci.RepositorySpec{},
	}
}

func (c *_oci) Context() oci.Context {
	return c.ctx
}

func (c *_oci) OpenCTF(path string) (oci.Repository, error) {
	ok, err := vfs.Exists(c.FileSystem(), path)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.ErrNotFound("file", path)
	}
	return ctfoci.Open(c.ctx, accessobj.ACC_WRITABLE, path, 0, accessio.PathFileSystem(c.FileSystem()))
}

////////////////////////////////////////////////////////////////////////////////

type _ocm struct {
	*_context
	ctx   ocm.Context
	repos map[string]ocm.RepositorySpec
}

func newOCM(ctx *_context, ocmctx ocm.Context) *_ocm {
	return &_ocm{
		_context: ctx,
		ctx:      ocmctx,
		repos:    map[string]ocm.RepositorySpec{},
	}
}

func (c *_ocm) Context() ocm.Context {
	return c.ctx
}

func (c *_ocm) OpenCTF(path string) (ocm.Repository, error) {
	ok, err := vfs.Exists(c.FileSystem(), path)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.ErrNotFound("file", path)
	}
	return ctfocm.Open(c.ctx, accessobj.ACC_WRITABLE, path, 0, c)
}
