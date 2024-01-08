// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package cpi

import (
	"encoding/base64"
	"encoding/json"

	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/utils"
)

type RawData []byte

var _ json.Unmarshaler = (*RawData)(nil)

func (r RawData) MarshalJSON() ([]byte, error) {
	return json.Marshal(base64.StdEncoding.EncodeToString(r))
}

func (r *RawData) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}
	*r, err = base64.StdEncoding.DecodeString(s)
	return err
}

type ContentSpec struct {
	Data       RawData        `json:"data,omitempty"`
	StringData string         `json:"stringdata,omitempty"`
	Path       string         `json:"path,omitempty"`
	Parsed     interface{}    `json:"-"`
	FileSystem vfs.FileSystem `json:"-"`
}

func (k *ContentSpec) Get() (interface{}, error) {
	if k.Parsed != nil {
		return k.Parsed, nil
	}
	if k.Data != nil {
		if k.StringData != "" || k.Path != "" {
			return nil, errors.Newf("only one of data, stringdata or path may be set")
		}
		return []byte(k.Data), nil
	}
	if k.StringData != "" {
		if k.Path != "" {
			return nil, errors.Newf("only one of data, stringdata or path may be set")
		}
		return []byte(k.StringData), nil
	}
	fs := k.FileSystem
	if fs == nil {
		fs = osfs.New()
	}

	return utils.ReadFile(k.Path, fs)
}
