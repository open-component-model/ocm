package config

import (
	"github.com/open-component-model/ocm/cmds/ocm/commands/common/options/keyoption"
	"github.com/open-component-model/ocm/pkg/cobrautils/logopts"
	"github.com/open-component-model/ocm/pkg/contexts/config"
)

type EvaluatedOptions struct {
	LogOpts       *logopts.EvaluatedOptions
	Keys          *keyoption.EvaluatedOptions
	ConfigForward config.Config
}
