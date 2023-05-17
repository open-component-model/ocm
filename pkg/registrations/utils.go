// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package registrations

import (
	"encoding/json"
	"fmt"

	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/runtime"
	"github.com/open-component-model/ocm/pkg/utils"
)

type Decoder func(data []byte, unmarshaller runtime.Unmarshaler) (interface{}, error)

func DecodeConfig[T any](config interface{}, d ...Decoder) (*T, error) {
	var err error

	var obj T
	cfg := &obj

	if config != nil {
		switch a := config.(type) {
		case string:
			cfg, err = decodeConfig[T]([]byte(a), d...)
		case json.RawMessage:
			cfg, err = decodeConfig[T](a, d...)
		case []byte:
			cfg, err = decodeConfig[T](a, d...)
		case *T:
			cfg = a
		case T:
			cfg = &a
		default:
			return nil, fmt.Errorf("unexpected type %T", a)
		}
		if err != nil {
			return nil, errors.Wrapf(err, "cannot unmarshal config")
		}
	}
	return cfg, nil
}

func decodeConfig[T any](data []byte, dec ...Decoder) (*T, error) {
	if d := utils.Optional(dec...); d != nil {
		r, err := d(data, runtime.DefaultYAMLEncoding)
		if err != nil {
			return nil, err
		}
		if eff, ok := r.(*T); ok {
			return eff, nil
		}
		return nil, errors.Newf("invalid decoded type %T ", r)
	}

	var c T
	err := runtime.DefaultYAMLEncoding.Unmarshal(data, &c)
	if err != nil {
		return nil, err
	}
	return &c, nil
}
