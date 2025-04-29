package internal

import (
	"github.com/mandelsoft/goutils/set"

	"ocm.software/ocm/api/utils/misc"
)

type Repository interface {
	ExistsCredentials(name string) (bool, error)
	LookupCredentials(name string) (Credentials, error)
	WriteCredentials(name string, creds Credentials) (Credentials, error)
}

type Credentials interface {
	CredentialsSource
	ExistsProperty(name string) bool
	GetProperty(name string) string
	PropertyNames() set.Set[string]
	Properties() misc.Properties
}

type DirectCredentials misc.Properties

var _ Credentials = (*DirectCredentials)(nil)

func NewCredentials(props misc.Properties) DirectCredentials {
	if props == nil {
		props = misc.Properties{}
	} else {
		props = props.Copy()
	}
	return DirectCredentials(props)
}

func (c DirectCredentials) ExistsProperty(name string) bool {
	_, ok := c[name]
	return ok
}

func (c DirectCredentials) GetProperty(name string) string {
	return c[name]
}

func (c DirectCredentials) PropertyNames() set.Set[string] {
	return misc.Properties(c).Names()
}

func (c DirectCredentials) Properties() misc.Properties {
	return misc.Properties(c).Copy()
}

func (c DirectCredentials) Credentials(Context, ...CredentialsSource) (Credentials, error) {
	return c, nil
}

func (c DirectCredentials) Copy() DirectCredentials {
	return DirectCredentials(misc.Properties(c).Copy())
}

func (c DirectCredentials) String() string {
	return misc.Properties(c).String()
}
