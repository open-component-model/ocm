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

package testhelper

import (
	"io"

	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/spf13/cobra"

	"github.com/open-component-model/ocm/cmds/ocm/app"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
	"github.com/open-component-model/ocm/pkg/contexts/config"
	"github.com/open-component-model/ocm/pkg/contexts/credentials"
	"github.com/open-component-model/ocm/pkg/contexts/datacontext"
	"github.com/open-component-model/ocm/pkg/contexts/oci"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/env"
	"github.com/open-component-model/ocm/pkg/env/builder"
)

type CLI struct {
	clictx.Context
}

func NewCLI(ctx clictx.Context) *CLI {
	if ctx == nil {
		ctx = clictx.DefaultContext()
	}
	return &CLI{ctx}
}

func (c *CLI) Execute(args ...string) error {
	cmd := app.NewCliCommand(c)
	cmd.SetArgs(args)
	return cmd.Execute()
}

func (c *CLI) ExecuteModified(mod func(ctx clictx.Context, cmd *cobra.Command), args ...string) error {
	cmd := app.NewCliCommand(c, mod)
	cmd.SetArgs(args)
	return cmd.Execute()
}

type TestEnv struct {
	*builder.Builder
	CLI
}

func NewTestEnv(opts ...env.Option) *TestEnv {
	b := builder.NewBuilder(env.NewEnvironment(opts...))
	ctx := clictx.WithOCM(b.OCMContext()).WithSharedAttributes(datacontext.New(nil)).New()
	return &TestEnv{
		Builder: b,
		CLI:     *NewCLI(ctx),
	}
}

func (e *TestEnv) ApplyOption(opts accessio.Options) error {
	return e.Builder.ApplyOption(opts)
}

func (e *TestEnv) ConfigContext() config.Context {
	return e.Builder.ConfigContext()
}

func (e *TestEnv) CredentialsContext() credentials.Context {
	return e.Builder.CredentialsContext()
}

func (e *TestEnv) OCMContext() ocm.Context {
	return e.Builder.OCMContext()
}

func (e *TestEnv) OCIContext() oci.Context {
	return e.Builder.OCIContext()
}

func (e *TestEnv) FileSystem() vfs.FileSystem {
	return e.Builder.FileSystem()
}

func (e *TestEnv) ReadTextFile(path string) (string, error) {
	data, err := e.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (e TestEnv) CatchOutput(w io.Writer) *TestEnv {
	e.Context = e.Context.WithStdIO(nil, w, nil)
	return &e
}

func (e TestEnv) CatchErrorOutput(w io.Writer) *TestEnv {
	e.Context = e.Context.WithStdIO(nil, nil, w)
	return &e
}

func (e TestEnv) WithInput(r io.Reader) *TestEnv {
	e.Context = e.Context.WithStdIO(r, nil, nil)
	return &e
}
