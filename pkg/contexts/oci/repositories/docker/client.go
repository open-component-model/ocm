package docker

import (
	"os"

	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/config"
	"github.com/docker/cli/cli/flags"
	"github.com/moby/moby/client"
	"github.com/spf13/pflag"
)

func newDockerClient(dockerhost string) (*client.Client, error) {
	if dockerhost == "" {
		opts := flags.NewClientOptions()
		// set defaults
		opts.SetDefaultOptions(pflag.NewFlagSet("", pflag.ContinueOnError))
		configfile := config.LoadDefaultConfigFile(os.Stderr)
		c, err := command.NewAPIClientFromFlags(opts, configfile)
		if err != nil {
			return nil, err
		}
		return c.(*client.Client), nil
	}
	c, err := client.NewClientWithOpts(client.FromEnv, client.WithHost(dockerhost))
	if err != nil {
		return nil, err
	}
	url, err := client.ParseHostURL(dockerhost)
	if err == nil && url.Scheme == "unix" {
		client.WithScheme(url.Scheme)(c)
	}
	return c, nil
}
