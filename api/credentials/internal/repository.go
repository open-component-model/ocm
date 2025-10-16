package internal

import (
	"github.com/mandelsoft/goutils/set"

	common "ocm.software/ocm/api/utils/misc"
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
	Properties() common.Properties
}

type DirectCredentials common.Properties

var _ Credentials = (*DirectCredentials)(nil)

func NewCredentials(props common.Properties) DirectCredentials {
	if props == nil {
		props = common.Properties{}
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
	return common.Properties(c).Names()
}

func (c DirectCredentials) Properties() common.Properties {
	return common.Properties(c).Copy()
}

func (c DirectCredentials) Credentials(Context, ...CredentialsSource) (Credentials, error) {
	return c, nil
}

func (c DirectCredentials) Copy() DirectCredentials {
	return DirectCredentials(common.Properties(c).Copy())
}

func (c DirectCredentials) String() string {
	return common.Properties(c).String()
}
