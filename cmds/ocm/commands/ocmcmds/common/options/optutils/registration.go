package optutils

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mandelsoft/goutils/errors"
	"github.com/spf13/pflag"
	"sigs.k8s.io/yaml"

	clictx "ocm.software/ocm/api/cli"
	"ocm.software/ocm/api/utils"
	"ocm.software/ocm/api/utils/cobrautils/flag"
)

type Registration struct {
	Name         string
	ArtifactType string
	MediaType    string
	Config       interface{}
}

func NewRegistrationOption(name, short, desc, usage string) RegistrationOption {
	return RegistrationOption{name: name, short: short, desc: desc, usage: usage}
}

type RegistrationOption struct {
	name          string
	short         string
	desc          string
	usage         string
	spec          map[string]string
	Registrations []*Registration
}

const RegistrationFormat = "<name>[:<artifact type>[:<media type>]]=<JSON target config"

func (o *RegistrationOption) AddFlags(fs *pflag.FlagSet) {
	flag.StringToStringVarP(fs, &o.spec, o.name, o.short, nil, fmt.Sprintf("%s (%s)", o.desc, RegistrationFormat))
}

func (o *RegistrationOption) HasRegistrations() bool {
	return len(o.spec) > 0 || len(o.Registrations) > 0
}

func (o *RegistrationOption) Configure(ctx clictx.Context) error {
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
					return fmt.Errorf("invalid %s registration %s must be of %s", o.name, n, RegistrationFormat)
				}
			}
		}

		var data interface{}
		var raw []byte
		var err error
		if strings.HasPrefix(v, "@") {
			raw, err = utils.ReadFile(v[1:], ctx.FileSystem())
			if err != nil {
				return errors.Wrapf(err, "cannot read %s config from %q", o.name, v[1:])
			}
		} else {
			if v != "" {
				raw = []byte(v)
			}
		}

		if len(raw) > 0 {
			var values interface{}
			if err = yaml.Unmarshal(raw, &values); err != nil {
				return errors.Wrapf(err, "invalid %s config %q", o.name, string(raw))
			}
			data, err = json.Marshal(values)
			if err != nil {
				return errors.Wrapf(err, "cannot marshal %s config", o.name)
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

func (o *RegistrationOption) Usage() string {
	s := fmt.Sprintf(`

If the <code>--%s</code> option is specified, appropriate %s handlers
are configured for the operation. It has the following format

<center>
    <pre>&lt;name>:&lt;artifact type>:&lt;media type>=&lt;yaml target config></pre>
</center>

The %s name may be a path expression with the following possibilities:
%s`, o.name, o.name, o.name, o.usage)
	return s
}
