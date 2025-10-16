package logging

import (
	"sigs.k8s.io/yaml"

	"ocm.software/ocm/api/ocm/plugin/ppi/cmds"
	"ocm.software/ocm/api/utils/cobrautils/logopts/logging"
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
