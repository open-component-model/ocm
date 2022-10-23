//  SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
//  SPDX-License-Identifier: Apache-2.0

package ociblob

import (
	"github.com/open-component-model/ocm/pkg/cobrautils/flagsets"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/options"
)

func ConfigHandler() flagsets.ConfigOptionTypeSetHandler {
	return flagsets.NewConfigOptionTypeSetHandler(
		Type, AddConfig,
		options.ReferenceOption,
		options.MediatypeOption,
		options.SizeOption,
		options.DigestOption,
	)
}

func AddConfig(opts flagsets.ConfigOptions, config flagsets.Config) error {
	if v, ok := opts.GetValue(options.ReferenceOption.Name()); ok {
		config["ref"] = v
	}
	if v, ok := opts.GetValue(options.MediatypeOption.Name()); ok {
		config["mediaType"] = v
	}
	if v, ok := opts.GetValue(options.SizeOption.Name()); ok {
		config["size"] = v
	}
	if v, ok := opts.GetValue(options.DigestOption.Name()); ok {
		config["digest"] = v
	}
	return nil
}
