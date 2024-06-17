package logging

import (
	"sigs.k8s.io/yaml"

	"github.com/open-component-model/ocm/pkg/cobrautils/logopts/logging"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/ppi/cmds"
)

func init() {
	cmds.RegisterLoggingConfigHandler(&loggingConfigHandler{})
}

type loggingConfigHandler struct{}

func (l loggingConfigHandler) HandleConfig(data []byte) error {
	var cfg logging.LoggingConfiguration

	err := yaml.Unmarshal(data, &cfg)
	if err != nil {
		return err
	}

	return cfg.Apply()
}
