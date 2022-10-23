//  SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
//  SPDX-License-Identifier: Apache-2.0

package github

import (
	"github.com/open-component-model/ocm/pkg/cobrautils/flagsets"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/options"
)

func ConfigHandler() flagsets.ConfigOptionTypeSetHandler {
	return flagsets.NewConfigOptionTypeSetHandler(
		Type, AddConfig,
		options.RepositoryOption,
		options.HostnameOption,
		options.CommitOption,
	)
}

func AddConfig(opts flagsets.ConfigOptions, config flagsets.Config) error {
	if v, ok := opts.GetValue(options.RepositoryOption.Name()); ok {
		config["repoUrl"] = v
	}
	if v, ok := opts.GetValue(options.CommitOption.Name()); ok {
		config["commit"] = v
	}
	if v, ok := opts.GetValue(options.HostnameOption.Name()); ok {
		config["apiHostname"] = v
	}
	return nil
}
