package transferhandlers

import (
	"slices"

	"github.com/mandelsoft/goutils/optionutils"
	"ocm.software/ocm/api/ocm/plugin/ppi"
	"ocm.software/ocm/cmds/transferplugin/config"
)

const (
	NAME = "demo"
)

func New() ppi.TransferHandler {
	h := ppi.NewTransferHandler(NAME, "enable value transport for dedicated external repositories")

	h.RegisterDecision(ppi.NewTransferResourceDecision(`value transport only for dedicated access types and service hosts`,
		ForOptions(func(options *ppi.TransferOptions) bool { return optionutils.AsBool(options.ResourcesByValue) })))

	h.RegisterDecision(ppi.NewTransferSourceDecision(`value transport only for dedicated access types and service hosts`,
		ForOptions(func(options *ppi.TransferOptions) bool { return optionutils.AsBool(options.SourcesByValue) })))
	return h
}

type OptionFunc func(opts *ppi.TransferOptions) bool

func ForOptions(f OptionFunc) ppi.ArtifactQuestionFunc {
	return func(p ppi.Plugin, question *ppi.ArtifactQuestionArguments) (bool, error) {
		var cfg *config.Config

		if question.Options.Special == nil {
			c, err := p.GetConfig()
			if c == nil || err != nil {
				return false, err
			}
			cfg = c.(*config.Config)
		} else {
			c, err := config.GetConfig(*question.Options.Special)
			if err != nil {
				return false, err
			}
			if c != nil {
				cfg = c.(*config.Config)
			}
		}

		if list := cfg.TransferRepositories.Types[question.Artifact.AccessInfo.Kind]; list != nil {
			host := question.Artifact.AccessInfo.Host
			return slices.Contains(list, host), nil
		}
		return f(&question.Options), nil
	}
}
