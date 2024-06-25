package internal

const AliasRepositoryType = "Alias"

type SetAliasFunction func(ctx Context, name string, spec RepositorySpec, creds CredentialsSource) error

type AliasRegistry interface {
	SetAlias(ctx Context, name string, spec RepositorySpec, creds CredentialsSource) error
}
