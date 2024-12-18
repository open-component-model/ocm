package npm

import (
	"fmt"
	"net/url"

	"ocm.software/ocm/api/ocm/cpi/accspeccpi"
	"ocm.software/ocm/api/utils/blobaccess/blobaccess"
	npmblob "ocm.software/ocm/api/utils/blobaccess/npm"
	"ocm.software/ocm/api/utils/mime"
	"ocm.software/ocm/api/utils/runtime"
)

// Type is the access type of NPM registry.
const (
	Type   = "npm"
	TypeV1 = Type + runtime.VersionSeparator + "v1"
)

func init() {
	accspeccpi.RegisterAccessType(accspeccpi.NewAccessSpecType[*AccessSpec](Type, accspeccpi.WithDescription(usage)))
	accspeccpi.RegisterAccessType(accspeccpi.NewAccessSpecType[*AccessSpec](TypeV1, accspeccpi.WithFormatSpec(formatV1), accspeccpi.WithConfigHandler(ConfigHandler())))
}

// AccessSpec describes the access for a NPM registry.
type AccessSpec struct {
	runtime.ObjectVersionedType `json:",inline"`

	// Registry is the base URL of the NPM registry
	Registry string `json:"registry"`
	// Package is the name of NPM package
	Package string `json:"package"`
	// Version of the NPM package.
	Version string `json:"version"`
}

var _ accspeccpi.AccessSpec = (*AccessSpec)(nil)

// New creates a new NPM registry access spec version v1.
func New(registry, pkg, version string) *AccessSpec {
	return &AccessSpec{
		ObjectVersionedType: runtime.NewVersionedTypedObject(Type),
		Registry:            registry,
		Package:             pkg,
		Version:             version,
	}
}

func (a *AccessSpec) Describe(_ accspeccpi.Context) string {
	return fmt.Sprintf("NPM package %s:%s in registry %s", a.Package, a.Version, a.Registry)
}

func (a *AccessSpec) Info(ctx accspeccpi.Context) *accspeccpi.UniformAccessSpecInfo {
	u, err := url.Parse(a.Registry)
	if err != nil {
		u = &url.URL{}
	}
	return &accspeccpi.UniformAccessSpecInfo{
		Kind: Type,
		Host: u.Hostname(),
		Port: u.Port(),
		Path: u.Path,
		Info: a.GetReferenceHint(nil),
	}
}

func (_ *AccessSpec) IsLocal(accspeccpi.Context) bool {
	return false
}

func (a *AccessSpec) GlobalAccessSpec(_ accspeccpi.Context) accspeccpi.AccessSpec {
	return a
}

func (a *AccessSpec) GetReferenceHint(_ accspeccpi.ComponentVersionAccess) string {
	return a.Package + ":" + a.Version
}

func (_ *AccessSpec) GetType() string {
	return Type
}

func (a *AccessSpec) AccessMethod(c accspeccpi.ComponentVersionAccess) (accspeccpi.AccessMethod, error) {
	return accspeccpi.AccessMethodForImplementation(newMethod(c, a))
}

////////////////////////////////////////////////////////////////////////////////

func newMethod(c accspeccpi.ComponentVersionAccess, a *AccessSpec) (accspeccpi.AccessMethodImpl, error) {
	factory := func() (blobaccess.BlobAccess, error) {
		return npmblob.BlobAccess(a.Registry, a.Package, a.Version, npmblob.WithDataContext(c.GetContext()))
	}
	return accspeccpi.NewDefaultMethodImpl(c, a, "", mime.MIME_TGZ, factory), nil
}
