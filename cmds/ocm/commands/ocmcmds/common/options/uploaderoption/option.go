// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package uploaderoption

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/spf13/pflag"
	"sigs.k8s.io/yaml"

	"github.com/open-component-model/ocm/cmds/ocm/pkg/options"
	"github.com/open-component-model/ocm/pkg/cobrautils/flag"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/registration"
	"github.com/open-component-model/ocm/pkg/errors"
)

func From(o options.OptionSetProvider) *Option {
	var opt *Option
	o.AsOptionSet().Get(&opt)
	return opt
}

type Registration struct {
	Name         string
	ArtifactType string
	MediaType    string
	Config       json.RawMessage
}

func New() *Option {
	return &Option{}
}

type Option struct {
	spec          map[string]string
	Registrations []*Registration
}

func (o *Option) AddFlags(fs *pflag.FlagSet) {
	flag.StringToStringVarP(fs, &o.spec, "uploader", "", nil, "repository uploader (<name>:<artifact type>[:<media type>]=<JSON target config)")
}

func (o *Option) Configure(ctx clictx.Context) error {
	desc := "<name>[:<artifact type>[:<media type>]]"
	for n, v := range o.spec {
		nam := n
		art := ""
		med := ""
		i := strings.Index(nam, ":")
		if i >= 0 {
			art = nam[i+1:]
			nam = nam[:i]
			i = strings.Index(art, ":")
			if i >= 0 {
				med = art[i+1:]
				art = art[:i]
				i = strings.Index(med, ":")
				if i >= 0 {
					return fmt.Errorf("invalid uploader registration %s must be of %s", n, desc)
				}
			}
		}

		var data json.RawMessage
		var err error
		if strings.HasPrefix(v, "@") {
			data, err = vfs.ReadFile(ctx.FileSystem(), v[1:])
			if err != nil {
				return errors.Wrapf(err, "cannot read upload target specification from %q", v[1:])
			}
		} else {
			var values interface{}
			if err = yaml.Unmarshal([]byte(v), &values); err != nil {
				return errors.Wrapf(err, "invalid target specification %q", v)
			}
			data, err = json.Marshal(values)
			if err != nil {
				return errors.Wrapf(err, "cannot marshal target specification")
			}
		}
		o.Registrations = append(o.Registrations, &Registration{
			Name:         nam,
			ArtifactType: art,
			MediaType:    med,
			Config:       data,
		})
	}
	return nil
}

func (o *Option) Register(ctx ocm.ContextProvider) error {
	for _, s := range o.Registrations {
		err := registration.RegisterBlobHandlerByName(ctx.OCMContext(), s.Name, s.Config,
			registration.ForArtifactType(s.ArtifactType), registration.ForMimeType(s.MediaType))
		if err != nil {
			return err
		}
	}
	return nil
}

func (o *Option) Usage() string {
	s := `
If the <code>--uploader</code> option is specified, appropriate uploaders
are configured for the transport target. It has the following format

<center>
    <pre>&lt;name>:&lt;artifact type>:&lt;media type>=&lt;yaml target config></pre>
</center>

The uploader name may be a path expression with the following possibilities:
- <code>ocm/ociRegistry</code>: oci Registry upload for local OCI artifact blobs.
  The media type is optional. If given ist must be an OCI artifact media type.
- <code>plugin/<plugin name>[/<uploader name]</code>: uploader provided by plugin.
`
	return s
}
