// Copyright 2022 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package localize

import (
	"github.com/mandelsoft/vfs/pkg/vfs"

	"github.com/open-component-model/ocm/pkg/common/compression"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/utils"
	"github.com/open-component-model/ocm/pkg/errors"
	utils2 "github.com/open-component-model/ocm/pkg/utils"
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
		return errors.Wrapf(err, "access template resource %s", rules.Template)
	}
	defer reader.Close()
	err = utils2.ExtractTarToFs(fs, reader)
	if err != nil {
		return errors.Wrapf(err, "cannot extract template filesystem %s,", rules.Template)
	}

	return errors.Wrapf(Substitute(subs, fs), "applying substitutions to template")
}
