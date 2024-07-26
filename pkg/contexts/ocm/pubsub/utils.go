package pubsub

import (
	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/goutils/maputils"

	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/listformat"
	"github.com/open-component-model/ocm/pkg/runtime"
)

func SetForRepo(repo cpi.Repository, spec PubSubSpec) error {
	prov := For(repo.GetContext()).For(repo.GetSpecification().GetKind())
	if prov != nil {
		return prov.SetPubSubSpec(repo, spec)
	}
	return errors.ErrNotSupported("pub/sub config")
}

func SpecForRepo(repo cpi.Repository) (PubSubSpec, error) {
	prov := For(repo.GetContext()).For(repo.GetSpecification().GetKind())
	if prov != nil {
		return prov.GetPubSubSpec(repo)
	}
	return nil, nil
}

func SpecForData(ctx cpi.ContextProvider, data []byte) (PubSubSpec, error) {
	return For(ctx).TypeScheme.Decode(data, runtime.DefaultYAMLEncoding)
}

func PubSubForRepo(repo cpi.Repository) (PubSubMethod, error) {
	spec, err := SpecForRepo(repo)
	if spec == nil || err != nil {
		return nil, err
	}
	return spec.PubSubMethod(repo)
}

func Notify(repo cpi.Repository, nv common.NameVersion) error {
	m, err := PubSubForRepo(repo)
	if m == nil || err != nil {
		return err
	}
	return m.NotifyComponentVersion(nv)
}

func PubSubUsage(scheme TypeScheme, providers ProviderRegistry, cli bool) string {
	s := `
The following list describes the supported publish/subscribe system types, their
specification versions, and formats:
`
	if len(scheme.KnownTypes()) == 0 {
		s += "There are currently no known pub/sub types!"
	} else {
		s += scheme.Describe()
	}

	list := maputils.OrderedKeys(providers.KnownProviders())
	if len(list) == 0 {
		s += "There are currently no persistence providers!"
	} else {
		s += "There are persistence providers for the following repository types:\n"
		s += listformat.FormatList("", list...)
	}
	return s
}
