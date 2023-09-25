// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package accessmethods

import (
	out "fmt"
	"io"
	"os"
	"strings"

	"github.com/mandelsoft/filepath/pkg/filepath"

	"github.com/open-component-model/ocm/cmds/demoplugin/common"
	"github.com/open-component-model/ocm/cmds/demoplugin/config"
	"github.com/open-component-model/ocm/pkg/cobrautils/flagsets"
	"github.com/open-component-model/ocm/pkg/contexts/credentials"
	"github.com/open-component-model/ocm/pkg/contexts/credentials/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/oci/identity"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/options"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/ppi"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/runtime"
)

const (
	NAME    = "demo"
	VERSION = "v1"
)

type AccessSpec struct {
	runtime.ObjectVersionedType `json:",inline"`

	Path      string `json:"path"`
	MediaType string `json:"mediaType,omitempty"`
}

type AccessMethod struct {
	ppi.AccessMethodBase
}

var PathOption = options.NewStringOptionType("accessPath", "path in temp repository")

var _ ppi.AccessMethod = (*AccessMethod)(nil)

func New() ppi.AccessMethod {
	return &AccessMethod{
		AccessMethodBase: ppi.MustNewAccessMethodBase(NAME, "", &AccessSpec{}, "demo access to temp files", ""),
	}
}

func (a *AccessMethod) Options() []options.OptionType {
	return []options.OptionType{
		options.MediatypeOption,
		PathOption,
	}
}

func (a *AccessMethod) Decode(data []byte, unmarshaler runtime.Unmarshaler) (runtime.TypedObject, error) {
	if unmarshaler == nil {
		unmarshaler = runtime.DefaultYAMLEncoding
	}
	var spec AccessSpec
	err := unmarshaler.Unmarshal(data, &spec)
	if err != nil {
		return nil, err
	}
	return &spec, nil
}

func (a *AccessMethod) ValidateSpecification(p ppi.Plugin, spec ppi.AccessSpec) (*ppi.AccessSpecInfo, error) {
	var info ppi.AccessSpecInfo

	my := spec.(*AccessSpec)

	if my.Path == "" {
		return nil, out.Errorf("path not specified")
	}
	if strings.HasPrefix(my.Path, "/") {
		return nil, out.Errorf("path must be relative (%s)", my.Path)
	}
	if my.MediaType == "" {
		return nil, out.Errorf("mediaType not specified")
	}
	info.MediaType = my.MediaType
	info.ConsumerId = credentials.ConsumerIdentity{
		cpi.ID_TYPE:            common.CONSUMER_TYPE,
		identity.ID_HOSTNAME:   "localhost",
		identity.ID_PATHPREFIX: my.Path,
	}
	info.Short = "temp file " + my.Path
	info.Hint = "temp file " + my.Path
	return &info, nil
}

func (a *AccessMethod) ComposeAccessSpecification(p ppi.Plugin, opts ppi.Config, config ppi.Config) error {
	list := errors.ErrListf("configuring options")
	list.Add(flagsets.AddFieldByOptionP(opts, PathOption, config, "path"))
	list.Add(flagsets.AddFieldByOptionP(opts, options.MediatypeOption, config, "mediaType"))
	return list.Result()
}

func (a *AccessMethod) Reader(p ppi.Plugin, spec ppi.AccessSpec, creds credentials.Credentials) (io.ReadCloser, error) {
	my := spec.(*AccessSpec)

	cfg, _ := p.GetConfig()
	root := os.TempDir()
	if cfg != nil && cfg.(*config.Config).AccessMethods.Path != "" {
		root = cfg.(*config.Config).Uploaders.Path
		err := os.MkdirAll(root, 0o700)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot create root dir")
		}
	}

	return os.Open(filepath.Join(root, my.Path))
}
