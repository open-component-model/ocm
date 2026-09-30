package wget

import (
	"ocm.software/ocm/api/utils/mime"
	"ocm.software/ocm/api/utils/runtime"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs"
)

const (
	TYPE          = "wget"
	TypeV1        = TYPE + runtime.VersionSeparator + "v1"
	UPPER_TYPE    = "Wget"
	UPPER_TYPE_V1 = UPPER_TYPE + runtime.VersionSeparator + "v1"
)

// Alias type names for the wget input type.
const (
	HTTP_TYPE          = "http"
	HTTP_TYPE_V1       = HTTP_TYPE + runtime.VersionSeparator + "v1"
	UPPER_HTTP_TYPE    = "HTTP"
	UPPER_HTTP_TYPE_V1 = UPPER_HTTP_TYPE + runtime.VersionSeparator + "v1"
)

func init() {
	inputs.DefaultInputTypeScheme.Register(inputs.NewInputType(TYPE, &Spec{}, usage, ConfigHandler()))
	inputs.DefaultInputTypeScheme.Register(inputs.NewInputType(TypeV1, &Spec{}, "", ConfigHandler()))
	inputs.DefaultInputTypeScheme.Register(inputs.NewInputType(UPPER_TYPE, &Spec{}, "", ConfigHandler()))
	inputs.DefaultInputTypeScheme.Register(inputs.NewInputType(UPPER_TYPE_V1, &Spec{}, "", ConfigHandler()))

	inputs.DefaultInputTypeScheme.Register(inputs.NewInputType(HTTP_TYPE, &Spec{}, "", HTTPConfigHandler()))
	inputs.DefaultInputTypeScheme.Register(inputs.NewInputType(HTTP_TYPE_V1, &Spec{}, "", HTTPConfigHandler()))
	inputs.DefaultInputTypeScheme.Register(inputs.NewInputType(UPPER_HTTP_TYPE, &Spec{}, "", HTTPConfigHandler()))
	inputs.DefaultInputTypeScheme.Register(inputs.NewInputType(UPPER_HTTP_TYPE_V1, &Spec{}, "", HTTPConfigHandler()))
}

const usage = `
The <code>url</code> is the url pointing to the http endpoint from which a resource is 
downloaded. The <code>mimeType</code> can be used to specify the MIME type of the 
resource. The input type can also be declared as <code>http</code> or <code>HTTP</code>.

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
