//  SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
//  SPDX-License-Identifier: Apache-2.0

package s3

import (
	"github.com/open-component-model/ocm/pkg/cobrautils/flagsets"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/options"
)

func ConfigHandler() flagsets.ConfigOptionTypeSetHandler {
	return flagsets.NewConfigOptionTypeSetHandler(
		Type, AddConfig,
		options.RegionOption,
		options.BucketOption,
		options.ReferenceOption,
		options.MediatypeOption,
		options.VersionOption,
	)
}

func AddConfig(opts flagsets.ConfigOptions, config flagsets.Config) error {
	if v, ok := opts.GetValue(options.ReferenceOption.Name()); ok {
		config["key"] = v
	}
	if v, ok := opts.GetValue(options.MediatypeOption.Name()); ok {
		config["mediaType"] = v
	}
	if v, ok := opts.GetValue(options.RegionOption.Name()); ok {
		config["region"] = v
	}
	if v, ok := opts.GetValue(options.BucketOption.Name()); ok {
		config["bucket"] = v
	}
	if v, ok := opts.GetValue(options.VersionOption.Name()); ok {
		config["version"] = v
	}
	return nil
}
