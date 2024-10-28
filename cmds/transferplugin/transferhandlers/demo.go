package transferhandlers

import (
	"slices"

	"github.com/mandelsoft/goutils/general"
	"github.com/mandelsoft/goutils/optionutils"
	"ocm.software/ocm/api/ocm/plugin/ppi"
	"ocm.software/ocm/cmds/transferplugin/config"
)

const (
	NAME = "demo"
)

func New() ppi.TransferHandler {
	h := ppi.NewTransferHandler(NAME, "enable value transport for dedicated external repositories")
	h.RegisterDecision(NewDecision(ppi.Q_TRANSFER_RESOURCE, func(options *ppi.TransferOptions) bool { return optionutils.AsBool(options.ResourcesByValue) },
		`value transport only for dedicated access types and service hosts`))
	h.RegisterDecision(NewDecision(ppi.Q_TRANSFER_SOURCE, func(options *ppi.TransferOptions) bool { return optionutils.AsBool(options.SourcesByValue) },
		`value transport only for dedicated access types and service hosts`))
	return h
}

type DecisionHandler struct {
	ppi.DecisionHandlerBase
	optfunc func(opts *ppi.TransferOptions) bool
}

var _ ppi.DecisionHandler = (*DecisionHandler)(nil)

func NewDecision(typ string, optfunc func(opts *ppi.TransferOptions) bool, desc ...string) ppi.DecisionHandler {
	return &DecisionHandler{
		DecisionHandlerBase: ppi.NewDecisionHandlerBase(typ, general.Optional(desc...)),
		optfunc:             optfunc,
	}
}

func (d DecisionHandler) DecideOn(p ppi.Plugin, question ppi.Question) (bool, error) {
	q := question.(*ppi.ArtifactQuestion)

	var cfg *config.Config

	if q.Options.Special == nil {
		c, err := p.GetConfig()
		if c == nil || err != nil {
			return false, err
		}
		cfg = c.(*config.Config)
	} else {
		c, err := config.GetConfig(*q.Options.Special)
		if err != nil {
			return false, err
		}
		if c != nil {
			cfg = c.(*config.Config)
		}
	}

	if list := cfg.TransferRepositories.Types[q.Artifact.AccessInfo.Kind]; list != nil {
		host := q.Artifact.AccessInfo.Host
		return slices.Contains(list, host), nil
	}
	return d.optfunc(&q.Options), nil
}
