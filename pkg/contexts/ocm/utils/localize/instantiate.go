// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package localize

import (
	"github.com/mandelsoft/vfs/pkg/vfs"

	"github.com/open-component-model/ocm/pkg/common/compression"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/utils"
	"github.com/open-component-model/ocm/pkg/errors"
	utils2 "github.com/open-component-model/ocm/pkg/utils/tarutils"
)

func Instantiate(rules *InstantiationRules, cv ocm.ComponentVersionAccess, resolver ocm.ComponentVersionResolver, config []byte, fs vfs.FileSystem) error {
	subs, err := Localize(rules.LocalizationRules, cv, resolver)
	if err != nil {
		return errors.Wrapf(err, "localization failed")
	}

	subs, err = Configure(rules.ConfigRules, subs, cv, resolver, rules.ConfigTemplate, config, rules.ConfigLibraries, rules.ConfigScheme)
	if err != nil {
		return errors.Wrapf(err, "applying instance configuration")
	}

	template, rcv, err := utils.ResolveResourceReference(cv, rules.Template, resolver)
	if err != nil {
		return errors.Wrapf(err, "resolving template resource %s", rules.Template)
	}
	defer rcv.Close()
	m, err := template.AccessMethod()
	if err != nil {
		return errors.Wrapf(err, "access template resource %s", rules.Template)
	}
	defer m.Close()
	r, err := m.Reader()
	if err != nil {
		return errors.Wrapf(err, "access template resource %s", rules.Template)
	}
	defer r.Close()

	reader, _, err := compression.AutoDecompress(r)
	if err != nil {
		return errors.Wrapf(err, "cannot determine compression for template resource %s", rules.Template)
	}
	defer reader.Close()

	if err = utils2.ExtractTarToFs(fs, reader); err != nil {
		return errors.Wrapf(err, "cannot package template filesystem %s,", rules.Template)
	}

	return errors.Wrapf(Substitute(subs, fs), "applying substitutions to template")
}
