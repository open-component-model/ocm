// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package cpi

import (
	"github.com/open-component-model/ocm/pkg/contexts/credentials/core"
)

const AliasRepositoryType = core.AliasRepositoryType

type AliasRegistry = core.AliasRegistry

type aliasRegistry struct {
	RepositoryType
	setter core.SetAliasFunction
}

var _ AliasRegistry = &aliasRegistry{}

func NewAliasRegistry(t RepositoryType, setter core.SetAliasFunction) RepositoryType {
	return &aliasRegistry{
		RepositoryType: t,
		setter:         setter,
	}
}

func (a *aliasRegistry) SetAlias(ctx Context, name string, spec RepositorySpec, creds CredentialsSource) error {
	return a.setter(ctx, name, spec, creds)
}
