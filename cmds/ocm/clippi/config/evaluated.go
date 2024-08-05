package config

import (
	"ocm.software/ocm/api/config"
	"ocm.software/ocm/api/utils/cobrautils/logopts"
	"ocm.software/ocm/cmds/ocm/commands/common/options/keyoption"
)

type EvaluatedOptions struct {
	LogOpts       *logopts.EvaluatedOptions
	Keys          *keyoption.EvaluatedOptions
	ConfigForward config.Config
}
