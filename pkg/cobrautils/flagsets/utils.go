//  SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
//  SPDX-License-Identifier: Apache-2.0

package flagsets

import (
	"fmt"
	"strings"
)

func GetField(config Config, names ...string) (interface{}, error) {
	var cur interface{} = config

	for i, n := range names {
		if cur == nil {
			return nil, nil
		}
		m, ok := cur.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("%s is no map", strings.Join(names[:i], "."))
		}
		cur = m[n]
	}
	return cur, nil
}

func SetField(config Config, value interface{}, names ...string) error {
	var last Config
	var cur interface{} = config

	if config == nil {
		return fmt.Errorf("no map given")
	}
	for i, n := range names {
		if cur == nil {
			cur = Config{}
			last[names[i-1]] = cur
		}
		m, ok := cur.(map[string]interface{})
		if !ok {
			return fmt.Errorf("%s is no map", strings.Join(names[:i], "."))
		}
		if i == len(names)-1 {
			m[n] = value
			return nil
		}
		last = m
		cur = m[n]
	}
	return fmt.Errorf("oops")
}
