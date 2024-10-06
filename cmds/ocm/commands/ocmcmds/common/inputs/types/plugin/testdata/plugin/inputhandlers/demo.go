package inputhandlers

import (
	out "fmt"
	"io"
	"os"
	"strings"

	"github.com/mandelsoft/filepath/pkg/filepath"
	"github.com/mandelsoft/goutils/errors"

	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/credentials/cpi"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/options"
	"ocm.software/ocm/api/ocm/plugin/ppi"
	"ocm.software/ocm/api/tech/oci/identity"
	"ocm.software/ocm/api/utils/cobrautils/flagsets"
	"ocm.software/ocm/api/utils/runtime"
	"ocm.software/ocm/cmds/demoplugin/common"
)

const (
	NAME    = "demo"
	VERSION = "v1"
)

type InputSpec struct {
	runtime.ObjectVersionedType `json:",inline"`

	Path      string `json:"path"`
	MediaType string `json:"mediaType,omitempty"`
}

type InputType struct {
	ppi.InputTypeBase
}

var PathOption = options.NewStringOptionType("inputPath", "path in temp file")

var _ ppi.InputType = (*InputType)(nil)

func New() ppi.InputType {
	return &InputType{
		InputTypeBase: ppi.MustNewInputTypeBase(NAME, &InputSpec{}, "demo access to temp files", ""),
	}
}

func (a *InputType) Options() []options.OptionType {
	return []options.OptionType{
		options.MediatypeOption,
		PathOption,
	}
}

func (a *InputType) Decode(data []byte, unmarshaler runtime.Unmarshaler) (runtime.TypedObject, error) {
	if unmarshaler == nil {
		unmarshaler = runtime.DefaultYAMLEncoding
	}
	var spec InputSpec
	err := unmarshaler.Unmarshal(data, &spec)
	if err != nil {
		return nil, err
	}
	return &spec, nil
}

func (a *InputType) ValidateSpecification(p ppi.Plugin, spec ppi.InputSpec) (*ppi.InputSpecInfo, error) {
	var info ppi.InputSpecInfo

	my := spec.(*InputSpec)

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
	info.Hint = my.Path

	return &info, nil
}

func (a *InputType) ComposeSpecification(p ppi.Plugin, opts ppi.Config, config ppi.Config) error {
	list := errors.ErrListf("configuring options")
	list.Add(flagsets.AddFieldByOptionP(opts, PathOption, config, "path"))
	list.Add(flagsets.AddFieldByOptionP(opts, options.MediatypeOption, config, "mediaType"))
	return list.Result()
}

func (a *InputType) Reader(p ppi.Plugin, spec ppi.InputSpec, creds credentials.Credentials) (io.ReadCloser, error) {
	my := spec.(*InputSpec)

	root := os.TempDir()
	return os.Open(filepath.Join(root, my.Path))
}
