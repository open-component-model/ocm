package config

import (
	"github.com/open-component-model/ocm/api/config"
	"github.com/open-component-model/ocm/api/utils/cobrautils/logopts"
	"github.com/open-component-model/ocm/cmds/ocm/commands/common/options/keyoption"
)

type EvaluatedOptions struct {
	LogOpts       *logopts.EvaluatedOptions
	Keys          *keyoption.EvaluatedOptions
	ConfigForward config.Config
}
