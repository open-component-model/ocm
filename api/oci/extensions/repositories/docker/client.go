// taken from "github.com/containers/image/v5", unfortunately this is private and cannot be used via import

package docker

import (
	"os"

	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/config"
	cliflags "github.com/docker/cli/cli/flags"
	dockerclient "github.com/docker/docker/client"
	mlog "github.com/mandelsoft/logging"
	"github.com/spf13/pflag"
	"ocm.software/ocm/api/utils/logging"
)

func newDockerClient(dockerhost string, logger mlog.UnboundLogger) (*dockerclient.Client, error) {
	if dockerhost == "" {
		opts := cliflags.NewClientOptions()
		// set defaults
		opts.SetDefaultOptions(pflag.NewFlagSet("", pflag.ContinueOnError))
		configfile := config.LoadDefaultConfigFile(os.Stderr)
		c, err := command.NewAPIClientFromFlags(opts, configfile)
		if err != nil {
			return nil, err
		}
		return c.(*dockerclient.Client), nil
	}
	c, err := dockerclient.NewClientWithOpts(dockerclient.FromEnv, dockerclient.WithHost(dockerhost))
	if err != nil {
		return nil, err
	}
	url, err := dockerclient.ParseHostURL(dockerhost)
	if err == nil && url.Scheme == "unix" {
		if err := dockerclient.WithScheme(url.Scheme)(c); err != nil {
			return nil, err
		}
	}
	clnt := c.HTTPClient()
	clnt.Transport = logging.NewRoundTripper(clnt.Transport, logger)
	if err := dockerclient.WithHTTPClient(clnt)(c); err != nil {
		return nil, err
	}
	return c, nil
}
