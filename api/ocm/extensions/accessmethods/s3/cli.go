package s3

import (
	"ocm.software/ocm/api/ocm/extensions/accessmethods/options"
	"ocm.software/ocm/api/utils/cobrautils/flagsets"
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
	flagsets.AddFieldByOptionP(opts, options.ReferenceOption, config, "key")
	flagsets.AddFieldByOptionP(opts, options.MediatypeOption, config, "mediaType")
	flagsets.AddFieldByOptionP(opts, options.RegionOption, config, "region")
	flagsets.AddFieldByOptionP(opts, options.BucketOption, config, "bucket")
	flagsets.AddFieldByOptionP(opts, options.VersionOption, config, "version")
	return nil
}

var usage = `
This method implements the access of a blob stored in an S3 bucket.
`
