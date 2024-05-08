package ocm

func AccessUsage(scheme AccessTypeScheme, cli bool) string {
	s := `
The following list describes the supported access methods, their versions
and specification formats.
Typically there is special support for the CLI artifact add commands.
The access method specification can be put below the <code>access</code> field.
If always requires the field <code>type</code> describing the kind and version
shown below.
`
	return s + scheme.Describe()
}
