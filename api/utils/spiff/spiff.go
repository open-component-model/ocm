package spiff

import (
	"encoding/json"
	"fmt"

	"github.com/ghodss/yaml"
	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/spiff/features"
	"github.com/mandelsoft/spiff/spiffing"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/modern-go/reflect2"
	"ocm.software/ocm/api/utils"
)

type Request struct {
	Template   spiffing.Source
	Stubs      []spiffing.Source
	ValuesNode string
	Values     interface{}
	Mode       int
	FileSystem vfs.FileSystem
	Functions  spiffing.Functions
}

func (r Request) GetValues() (map[string]interface{}, error) {
	if reflect2.IsNil(r.Values) {
		return nil, nil
	}

	var (
		err  error
		data []byte
	)

	if b, ok := r.Values.([]byte); ok {
		data = b
	} else {
		data, err = json.Marshal(r.Values)
		if err != nil {
			return nil, errors.ErrInvalidWrap(err, "values", fmt.Sprintf("%T", r.Values))
		}
	}

	var values interface{}
	err = yaml.Unmarshal(data, &values)
	if err != nil {
		return nil, errors.ErrInvalidWrap(err, "values", fmt.Sprintf("%T", r.Values))
	}
	if r.ValuesNode != "" {
		return map[string]interface{}{r.ValuesNode: values}, nil
	}
	if v, ok := values.(map[string]interface{}); ok {
		return v, nil
	}
	return nil, errors.ErrInvalid("values", fmt.Sprintf("%T", values))
}

func (r *Request) GetSpiff() (spiffing.Spiff, error) {
	spiff := spiffing.New().WithFeatures(features.CONTROL, features.INTERPOLATION).WithFileSystem(utils.FileSystem(r.FileSystem)).WithMode(r.Mode).WithFunctions(r.Functions)
	values, err := r.GetValues()
	if err != nil {
		return nil, err
	}
	if values != nil {
		spiff, err = spiff.WithValues(values)
	}
	if err != nil {
		return nil, err
	}
	return spiff, nil
}

func Cascade(req *Request) ([]byte, error) {
	if req.Template == nil {
		return nil, nil
	}
	spiff, err := req.GetSpiff()
	if err != nil {
		return nil, err
	}
	stubs := []spiffing.Node{}

	data, err := req.Template.Data()
	if err != nil {
		return nil, errors.ErrInvalidWrap(err, "template", req.Template.Name())
	}
	templ, err := spiff.Unmarshal("template "+req.Template.Name(), data)
	if err != nil {
		return nil, errors.Wrapf(err, "template: %s", req.Template.Name())
	}

	for i, s := range req.Stubs {
		data, err := s.Data()
		if err != nil {
			return nil, errors.ErrInvalidWrap(err, "stub", s.Name())
		}
		stub, err := spiff.Unmarshal(s.Name(), data)
		if err != nil {
			return nil, errors.Wrapf(err, "stub %d (%s)", i+1, s.Name())
		}
		stubs = append(stubs, stub)
	}

	node, err := spiff.Cascade(templ, stubs)
	if err != nil {
		return nil, errors.Wrapf(err, "processing template %s", req.Template.Name())
	}
	return spiff.Marshal(node)
}

func CascadeWith(opts ...Option) ([]byte, error) {
	req, err := GetRequest(opts...)
	if err != nil {
		return nil, err
	}
	return Cascade(req)
}
