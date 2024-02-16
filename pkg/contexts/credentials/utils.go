// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package credentials

import (
	"strings"

	"github.com/texttheater/golang-levenshtein/levenshtein"

	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/utils"
)

func GetProvidedConsumerId(obj interface{}, uctx ...UsageContext) ConsumerIdentity {
	return utils.UnwrappingCall(obj, func(provider ConsumerIdentityProvider) ConsumerIdentity {
		return provider.GetConsumerId(uctx...)
	})
}

func GetProvidedIdentityMatcher(obj interface{}) string {
	return utils.UnwrappingCall(obj, func(provider ConsumerIdentityProvider) string {
		return provider.GetIdentityMatcher()
	})
}

func CredentialsFor(ctx ContextProvider, obj interface{}, uctx ...UsageContext) (Credentials, error) {
	id := GetProvidedConsumerId(obj, uctx...)
	if id == nil {
		return nil, errors.ErrNotSupported(KIND_CONSUMER)
	}
	return CredentialsForConsumer(ctx, id)
}

func GuessConsumerType(ctxp ContextProvider, spec string) string {
	matchers := ctxp.CredentialsContext().ConsumerIdentityMatchers()
	lspec := strings.ToLower(spec)

	if matchers.Get(spec) == nil {
		fix := ""
		for _, i := range matchers.List() {
			idx := strings.Index(i.Type, ".")
			if idx > 0 && i.Type[:idx] == spec {
				fix = i.Type
				break
			}
		}
		if fix == "" {
			for _, i := range matchers.List() {
				if strings.ToLower(i.Type) == lspec {
					fix = i.Type
					break
				}
			}
		}
		if fix == "" {
			for _, i := range matchers.List() {
				idx := strings.Index(i.Type, ".")
				if idx > 0 && strings.ToLower(i.Type[:idx]) == lspec {
					fix = i.Type
					break
				}
			}
		}
		if fix == "" {
			min := -1
			for _, i := range matchers.List() {
				idx := strings.Index(i.Type, ".")
				if idx > 0 {
					d := levenshtein.DistanceForStrings([]rune(lspec), []rune(strings.ToLower(i.Type[:idx])), levenshtein.DefaultOptions)
					if d < 5 && fix == "" || min > d {
						fix = i.Type
						min = d
					}
				}
			}
		}
		if fix == "" {
			min := -1
			for _, i := range matchers.List() {
				d := levenshtein.DistanceForStrings([]rune(lspec), []rune(strings.ToLower(i.Type)), levenshtein.DefaultOptions)
				if d < 5 && fix == "" || min > d {
					fix = i.Type
					min = d
				}
			}
		}
		if fix != "" {
			return fix
		}
	}
	return spec
}
