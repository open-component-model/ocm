// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"strings"

	"github.com/open-component-model/ocm/pkg/errors"
)

func ParseSettings(args []string, kinds ...string) (map[string]string, error) {
	kind := "setting"
	if len(kinds) > 0 {
		kind = kinds[0]
	}

	settings := map[string]string{}
	for _, arg := range args {
		if i := strings.Index(arg, "="); i > 0 {
			value := arg[i+1:]
			name := strings.TrimSpace(arg[0:i])
			settings[name] = value
		} else {
			return nil, errors.Newf("invalid %s %q (assignment required)", kind, arg)
		}
	}
	return settings, nil
}

func FilterSettings(args ...string) (attrs map[string]string, addArgs []string) {
	for _, arg := range args {
		if i := strings.Index(arg, "="); i > 0 {
			if attrs == nil {
				attrs = map[string]string{}
			}
			value := arg[i+1:]
			name := strings.TrimSpace(arg[0:i])
			attrs[name] = value
			continue
		}
		addArgs = append(addArgs, arg)
	}
	return attrs, addArgs
}
