// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package localize

import (
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
)

// Localize maps a list of filesystem related localization requests to an
// appropriate set of substitution requests.
func Localize(mappings []Localization, cv ocm.ComponentVersionAccess, resolver ocm.ComponentVersionResolver) (Substitutions, error) {
	var result Substitutions

	for i, v := range mappings {
		m, err := v.Evaluate(i, cv, resolver)
		if err != nil {
			return nil, err
		}
		for _, r := range m {
			result.AddValueMapping(&r, v.FilePath)
		}
	}
	return result, nil
}

// LocalizeMappings maps a set of pure image mappings into
// an appropriate set of value mapping request for a single data object.
func LocalizeMappings(mappings ImageMappings, cv ocm.ComponentVersionAccess, resolver ocm.ComponentVersionResolver) (ValueMappings, error) {
	var result ValueMappings

	for i, v := range mappings {
		m, err := v.Evaluate(i, cv, resolver)
		if err != nil {
			return nil, err
		}
		result = append(result, m...)
	}
	return result, nil
}

func substitutionName(name, sub string, cnt int) string {
	if name == "" {
		return sub
	}
	if cnt <= 1 {
		return name
	}
	return name + "-" + sub
}
