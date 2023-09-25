// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package testhelper

import (
	"io"

	"github.com/mandelsoft/vfs/pkg/vfs"

	"github.com/open-component-model/ocm/cmds/helminstaller/app"
	"github.com/open-component-model/ocm/cmds/helminstaller/app/driver"
	"github.com/open-component-model/ocm/cmds/helminstaller/app/driver/helm"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
	"github.com/open-component-model/ocm/pkg/contexts/config"
	"github.com/open-component-model/ocm/pkg/contexts/credentials"
	"github.com/open-component-model/ocm/pkg/contexts/oci"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/env"
	"github.com/open-component-model/ocm/pkg/env/builder"
)

type CLI struct {
	clictx.Context
	Driver driver.Driver
}

func NewCLI(ctx clictx.Context) *CLI {
	if ctx == nil {
		ctx = clictx.DefaultContext()
	}
	return &CLI{Context: ctx, Driver: helm.New()}
}

func (c *CLI) Execute(args ...string) error {
	cmd := app.NewCliCommand(clictx.DefaultContext(), c.Driver)
	cmd.SetArgs(args)
	return cmd.Execute()
}

type TestEnv struct {
	*builder.Builder
	CLI
}

func NewTestEnv(opts ...env.Option) *TestEnv {
	b := builder.NewBuilder(opts...)
	ctx := clictx.WithOCM(b.OCMContext()).New()
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
