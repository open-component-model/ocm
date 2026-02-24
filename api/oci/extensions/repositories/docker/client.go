// taken from "github.com/containers/image/v5", unfortunately this is private and cannot be used via import

package docker

import (
	"net/http"
	"os"

	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/config"
	cliflags "github.com/docker/cli/cli/flags"
	mlog "github.com/mandelsoft/logging"
	dockerclient "github.com/moby/moby/client"
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
	var opts []dockerclient.Opt
	opts = append(opts, dockerclient.FromEnv)
	opts = append(opts, dockerclient.WithHost(dockerhost))
	url, err := dockerclient.ParseHostURL(dockerhost)
	if err == nil && url.Scheme == "unix" {
		opts = append(opts, dockerclient.WithScheme(url.Scheme))
	}
	clnt := http.Client{}
	clnt.Transport = logging.NewRoundTripper(clnt.Transport, logger)
	opts = append(opts, dockerclient.WithHTTPClient(&clnt))
	c, err := dockerclient.New(opts...)
	if err != nil {
		return nil, err
	}

	return c, nil
}
