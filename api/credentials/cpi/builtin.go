package cpi

import (
	"ocm.software/ocm/api/credentials/internal"
)

const AliasRepositoryType = internal.AliasRepositoryType

type AliasRegistry = internal.AliasRegistry

type aliasRegistry struct {
	RepositoryType
	setter internal.SetAliasFunction
}

var _ AliasRegistry = &aliasRegistry{}

func NewAliasRegistry(t RepositoryType, setter internal.SetAliasFunction) RepositoryType {
	return &aliasRegistry{
		RepositoryType: t,
		setter:         setter,
	}
}

func (a *aliasRegistry) SetAlias(ctx Context, name string, spec RepositorySpec, creds CredentialsSource) error {
	return a.setter(ctx, name, spec, creds)
}
