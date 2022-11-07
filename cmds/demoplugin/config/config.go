// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"encoding/json"
)

type Config struct {
	AccessMethods Values `json:"accessMethods"`
	Uploaders     Values `json:"uploaders"`
}

type Values struct {
	Path string `json:"path"`
}

func GetConfig(raw json.RawMessage) (interface{}, error) {
	var cfg Config

	err := json.Unmarshal(raw, &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}
