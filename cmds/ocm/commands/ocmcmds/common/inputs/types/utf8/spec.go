package utf8

import (
	"encoding/json"

	"k8s.io/apimachinery/pkg/util/validation/field"

	"ocm.software/ocm/api/utils/blobaccess"
	"ocm.software/ocm/api/utils/runtime"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs/cpi"
)

type Spec struct {
	inputs.InputSpecBase `json:",inline"`
	cpi.ProcessSpec      `json:",inline"`

	// Text is an utf8 string
	Text          string          `json:"text,omitempty"`
	Json          json.RawMessage `json:"json,omitempty"`
	FormattedJson json.RawMessage `json:"formattedJson,omitempty"`
	Yaml          json.RawMessage `json:"yaml,omitempty"`
}

var _ inputs.InputSpec = (*Spec)(nil)

func New(text string, mediatype string, compress bool) *Spec {
	return &Spec{
		InputSpecBase: inputs.InputSpecBase{
			ObjectVersionedType: runtime.ObjectVersionedType{
				Type: TYPE,
			},
		},
		ProcessSpec: cpi.NewProcessSpec(mediatype, compress),
		Text:        text,
	}
}

func NewJson(data interface{}, mediatype string, compress bool) (*Spec, error) {
	raw, err := runtime.DefaultJSONEncoding.Marshal(data)
	if err != nil {
		return nil, err
	}
	return &Spec{
		InputSpecBase: inputs.InputSpecBase{
			ObjectVersionedType: runtime.ObjectVersionedType{
				Type: TYPE,
			},
		},
		ProcessSpec: cpi.NewProcessSpec(mediatype, compress),
		Json:        json.RawMessage(raw),
	}, nil
}

func NewFormattedJson(data interface{}, mediatype string, compress bool) (*Spec, error) {
	raw, err := runtime.DefaultJSONEncoding.Marshal(data)
	if err != nil {
		return nil, err
	}
	return &Spec{
		InputSpecBase: inputs.InputSpecBase{
			ObjectVersionedType: runtime.ObjectVersionedType{
				Type: TYPE,
			},
		},
		ProcessSpec:   cpi.NewProcessSpec(mediatype, compress),
		FormattedJson: json.RawMessage(raw),
	}, nil
}

func NewYaml(data interface{}, mediatype string, compress bool) (*Spec, error) {
	raw, err := runtime.DefaultJSONEncoding.Marshal(data)
	if err != nil {
		return nil, err
	}
	return &Spec{
		InputSpecBase: inputs.InputSpecBase{
			ObjectVersionedType: runtime.ObjectVersionedType{
				Type: TYPE,
			},
		},
		ProcessSpec: cpi.NewProcessSpec(mediatype, compress),
		Yaml:        json.RawMessage(raw),
	}, nil
}

func (s *Spec) Validate(fldPath *field.Path, ctx inputs.Context, inputFilePath string) field.ErrorList {
	cnt := 0
	if s.Text != "" {
		cnt++
	}
	if s.Json != nil {
		cnt++
	}
	if s.FormattedJson != nil {
		cnt++
	}
	if s.Yaml != nil {
		cnt++
	}
	if cnt > 1 {
		allErrs := field.ErrorList{}
		allErrs = append(allErrs, field.Forbidden(fldPath, "only one of the fields text, json, formattedJson or yaml can be set"))
		return allErrs
	}
	return nil
}

func (s *Spec) GetBlob(ctx inputs.Context, info inputs.InputResourceInfo) (blobaccess.BlobAccess, string, error) {
	data, err := Plain([]byte(s.Text))

	if s.Json != nil {
		data, err = Json(s.Json)
	}
	if s.FormattedJson != nil {
		data, err = FormattedJson(s.FormattedJson)
	}
	if s.Yaml != nil {
		data, err = Yaml(s.Yaml)
	}
	if err != nil {
		return nil, "", err
	}
	return s.ProcessBlob(ctx, blobaccess.DataAccessForData(data), ctx.FileSystem())
}

func Prepare(raw []byte) (interface{}, error) {
	var v interface{}
	err := runtime.DefaultJSONEncoding.Unmarshal(raw, &v)
	if err != nil {
		return nil, err
	}
	if s, ok := v.(string); ok {
		v = nil
		err := runtime.DefaultYAMLEncoding.Unmarshal([]byte(s), &v)
		if err != nil {
			return nil, err
		}
	}
	return v, nil
}

type OutputFormat func(data []byte) ([]byte, error)

func Plain(data []byte) ([]byte, error) {
	return data, nil
}

func Json(raw []byte) ([]byte, error) {
	v, err := Prepare(raw)
	if err != nil {
		return nil, err
	}
	return json.Marshal(v)
}

func FormattedJson(raw []byte) ([]byte, error) {
	v, err := Prepare(raw)
	if err != nil {
		return nil, err
	}
	return json.MarshalIndent(v, "", "  ")
}

func Yaml(raw []byte) ([]byte, error) {
	v, err := Prepare(raw)
	if err != nil {
		return nil, err
	}
	return runtime.DefaultYAMLEncoding.Marshal(v)
}
