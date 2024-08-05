package wget

import (
	"ocm.software/ocm/api/ocm/extensions/accessmethods/options"
	"ocm.software/ocm/api/utils/cobrautils/flagsets"
	"ocm.software/ocm/api/utils/mime"
)

func ConfigHandler() flagsets.ConfigOptionTypeSetHandler {
	return flagsets.NewConfigOptionTypeSetHandler(
		Type, AddConfig,
		options.URLOption,
		options.MediatypeOption,
		options.HTTPHeaderOption,
		options.HTTPVerbOption,
		options.HTTPBodyOption,
		options.HTTPRedirectOption,
	)
}

func AddConfig(opts flagsets.ConfigOptions, config flagsets.Config) error {
	flagsets.AddFieldByOptionP(opts, options.URLOption, config, "url")
	flagsets.AddFieldByOptionP(opts, options.MediatypeOption, config, "mediaType")
	flagsets.AddFieldByOptionP(opts, options.HTTPHeaderOption, config, "header")
	flagsets.AddFieldByOptionP(opts, options.HTTPVerbOption, config, "verb")
	flagsets.AddFieldByOptionP(opts, options.HTTPBodyOption, config, "body")
	flagsets.AddFieldByOptionP(opts, options.HTTPRedirectOption, config, "noredirect")
	return nil
}

var usage = `
This method implements access to resources stored on an http server.
`

var formatV1 = `
The <code>url</code> is the url pointing to the http endpoint from which a resource is
downloaded. The <code>mimeType</code> can be used to specify the MIME type of the
resource.

This blob type specification supports the following fields:
- **<code>url</code>** *string*

This REQUIRED property describes the url from which the resource is to be
downloaded.

- **<code>mediaType</code>** *string*

This OPTIONAL property describes the media type of the resource to be
downloaded. If omitted, ocm tries to read the mediaType from the Content-Type header
of the http response. If the mediaType cannot be set from the Content-Type header as well,
ocm tries to deduct the mediaType from the URL. If that is not possible either, the default
media type is defaulted to ` + mime.MIME_OCTET + `.

- **<code>header</code>** *map[string][]string*

This OPTIONAL property describes the http headers to be set in the http request to the server.

- **<code>verb</code>** *string*

This OPTIONAL property describes the http verb (also known as http request method) for the http
request. If omitted, the http verb is defaulted to GET.

- **<code>body</code>** *[]byte*

This OPTIONAL property describes the http body to be included in the request.

- **<code>noredirect</code>** *bool*

This OPTIONAL property describes whether http redirects should be disabled. If omitted,
it is defaulted to false (so, per default, redirects are enabled).
`
