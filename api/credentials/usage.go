package credentials

func RepositoryUsage(scheme RepositoryTypeScheme) string {
	s := `
The following list describes the supported credential providers
(credential repositories), their specification versions
and formats. Because of the extensible nature of the OCM model,
credential consum
`
	return s + scheme.Describe()
}
