package localize

import (
	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/vfs/pkg/vfs"

	"ocm.software/ocm/api/ocm"
	resourcetypes "ocm.software/ocm/api/ocm/extensions/artifacttypes"
	"ocm.software/ocm/api/ocm/extensions/download"
	"ocm.software/ocm/api/ocm/resourcerefs"
	common "ocm.software/ocm/api/utils/misc"
)

func Instantiate(rules *InstantiationRules, cv ocm.ComponentVersionAccess, resolver ocm.ComponentVersionResolver, config []byte, fs vfs.FileSystem, types ...string) error {
	subs, err := Localize(rules.LocalizationRules, cv, resolver)
	if err != nil {
		return errors.Wrapf(err, "localization failed")
	}

	subs, err = Configure(rules.ConfigRules, subs, cv, resolver, rules.ConfigTemplate, config, rules.ConfigLibraries, rules.ConfigScheme)
	if err != nil {
		return errors.Wrapf(err, "applying instance configuration")
	}

	template, rcv, err := resourcerefs.ResolveResourceReference(cv, rules.Template, resolver)
	if err != nil {
		return errors.Wrapf(err, "resolving template resource %s", rules.Template)
	}
	defer rcv.Close()

	if len(types) != 0 {
		found := false
		for _, t := range types {
			found = found || (t == template.Meta().Type)
		}
		if !found {
			return errors.ErrInvalid(resourcetypes.KIND_RESOURCE_TYPE, template.Meta().Type)
		}
	}

	ok, _, err := download.For(cv.GetContext()).Download(common.NewPrinter(nil), template, ".", fs)
	if err != nil {
		return errors.Wrapf(err, "cannot download resource %s", rules.Template)
	}
	if !ok {
		return errors.Wrapf(err, "cannot download resource %s: no downloader found", rules.Template)
	}

	return errors.Wrapf(Substitute(subs, fs), "applying substitutions to template")
}
