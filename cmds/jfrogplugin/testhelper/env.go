package testhelper

import (
	"bytes"
	"io"

	"gopkg.in/yaml.v3"

	"ocm.software/ocm/api/helper/builder"
	"ocm.software/ocm/api/helper/env"
	"ocm.software/ocm/api/ocm/plugin/ppi"
	"ocm.software/ocm/api/ocm/plugin/ppi/cmds"
	"ocm.software/ocm/cmds/jfrogplugin/config"
	jfrogppi "ocm.software/ocm/cmds/jfrogplugin/ppi"
)

type CLI struct {
	ppi.Plugin
	*cmds.PluginCommand

	config string

	output *bytes.Buffer
}

func NewCLI() (*CLI, error) {
	plugin, err := jfrogppi.Plugin()
	if err != nil {
		return nil, err
	}
	cmd := cmds.NewPluginCommand(plugin)
	cli := &CLI{Plugin: plugin, PluginCommand: cmd, output: &bytes.Buffer{}}
	cmd.Command().SetOut(cli.output)
	return cli, nil
}

func (cli *CLI) Execute(args ...string) error {
	cli.output.Reset()
	if cli.config != "" {
		args = append(args, "--config", cli.config)
	}
	return cli.PluginCommand.Execute(args)
}

func (cli *CLI) SetConfig(cfg *config.Config) error {
	var data bytes.Buffer
	if err := yaml.NewEncoder(&data).Encode(cfg); err != nil {
		return err
	}

	cli.config = data.String()

	return nil
}

func (cli *CLI) GetOutput() []byte {
	return cli.output.Bytes()
}

func (cli *CLI) SetInput(data io.Reader) {
	cli.PluginCommand.Command().SetIn(data)
}

type TestEnv struct {
	*builder.Builder
	*CLI
}

func NewTestEnv(opts ...env.Option) (*TestEnv, error) {
	b := builder.NewBuilder(opts...)

	cli, err := NewCLI()
	if err != nil {
		return nil, err
	}

	return &TestEnv{
		Builder: b,
		CLI:     cli,
	}, nil
}
