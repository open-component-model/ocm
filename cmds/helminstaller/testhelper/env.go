package testhelper

import (
	"io"

	"github.com/mandelsoft/vfs/pkg/vfs"
	clictx "ocm.software/ocm/api/cli"
	"ocm.software/ocm/api/config"
	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/helper/builder"
	"ocm.software/ocm/api/helper/env"
	"ocm.software/ocm/api/oci"
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/cmds/helminstaller/app"
	"ocm.software/ocm/cmds/helminstaller/app/driver"
	"ocm.software/ocm/cmds/helminstaller/app/driver/helm"
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
