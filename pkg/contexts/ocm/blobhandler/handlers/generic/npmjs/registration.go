package npmjs

import (
	"fmt"

	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/listformat"
	"github.com/open-component-model/ocm/pkg/registrations"
)

type Config = Attribute

func init() {
	cpi.RegisterBlobHandlerRegistrationHandler("npmjs/package", &RegistrationHandler{})
}

type RegistrationHandler struct{}

var _ cpi.BlobHandlerRegistrationHandler = (*RegistrationHandler)(nil)

func (r *RegistrationHandler) RegisterByName(handler string, ctx cpi.Context, config cpi.BlobHandlerConfig, olist ...cpi.BlobHandlerOption) (bool, error) {
	if handler != "" {
		return true, fmt.Errorf("invalid npmjsArtifact handler %q", handler)
	}
	if config == nil {
		return true, fmt.Errorf("npmjs target specification required")
	}
	attr, err := registrations.DecodeConfig[Config](config, AttributeType{}.Decode)
	if err != nil {
		return true, errors.Wrapf(err, "blob handler configuration")
	}

	var mimes []string
	opts := cpi.NewBlobHandlerOptions(olist...)
	if opts.MimeType == "npmjs" {
		mimes = append(mimes, opts.MimeType)
	}

	h := NewArtifactHandler(attr)
	for _, m := range mimes {
		opts.MimeType = m
		ctx.BlobHandlers().Register(h, opts)
	}

	return true, nil
}

func (r *RegistrationHandler) GetHandlers(ctx cpi.Context) registrations.HandlerInfos {
	return registrations.NewLeafHandlerInfo("downloading npmjs artifacts", `
The <code>npmjsArtifacts</code> downloader is able to download npmjs artifacts
as artifact archive according to the npmjs package spec.
The following artifact media types are supported: npmjs
By default, it is registered for these mimetypes.

It accepts a config with the following fields:
`+listformat.FormatMapElements("", AttributeDescription())+`
Alternatively, a single string value can be given representing a npmjs repository
reference.`,
	)
}
