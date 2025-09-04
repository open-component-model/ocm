package relativeociref

import (
	"fmt"

	"ocm.software/ocm/api/oci"
	"ocm.software/ocm/api/ocm/cpi/accspeccpi"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/ociartifact"
	"ocm.software/ocm/api/ocm/internal"
	"ocm.software/ocm/api/utils/runtime"
)

// Type describes the access of an OCI artifact stored as OCI artifact in the OCI
// registry hosting the actual component version.
const (
	Type   = "relativeOciReference"
	TypeV1 = Type + runtime.VersionSeparator + "v1"
)

func init() {
	accspeccpi.RegisterAccessType(accspeccpi.NewAccessSpecType[*AccessSpec](Type))
	accspeccpi.RegisterAccessType(accspeccpi.NewAccessSpecType[*AccessSpec](TypeV1))
}

var _ accspeccpi.HintProvider = (*AccessSpec)(nil)

// New creates a new relativeOciReference accessor.
func New(ref string) *AccessSpec {
	return &AccessSpec{
		ObjectVersionedType: runtime.NewVersionedObjectType(Type),
		Reference:           ref,
	}
}

// AccessSpec describes the access of an OCI artifact stored as OCI artifact in
// the OCI registry hosting the actual component version.
type AccessSpec struct {
	runtime.ObjectVersionedType `json:",inline"`
	// Reference is the OCI repository name and version separated by a colon.
	Reference string `json:"reference"`
}

func (a *AccessSpec) Describe(context accspeccpi.Context) string {
	return fmt.Sprintf("local OCI artifact %s", a.Reference)
}

func (a *AccessSpec) IsLocal(context accspeccpi.Context) bool {
	return true
}

func (a *AccessSpec) GlobalAccessSpec(context accspeccpi.Context) accspeccpi.AccessSpec {
	return nil
}

func (a *AccessSpec) AccessMethod(access accspeccpi.ComponentVersionAccess) (accspeccpi.AccessMethod, error) {
	return access.AccessMethod(a)
}

func (a *AccessSpec) GetDigest() (string, bool) {
	ref, err := oci.ParseRef(a.Reference)
	if err != nil {
		return "", true
	}
	if ref.Digest != nil {
		return ref.Digest.String(), true
	}
	return "", false
}

func (a *AccessSpec) GetReferenceHint(cv internal.ComponentVersionAccess) string {
	return a.Reference
}

func (a *AccessSpec) GetOCIReference(cv accspeccpi.ComponentVersionAccess) (string, error) {
	if cv == nil {
		return "", fmt.Errorf("component version required to determine OCI reference")
	}
	m, err := a.AccessMethod(cv)
	if err != nil {
		return "", err
	}
	defer m.Close()

	if o, ok := accspeccpi.GetAccessMethodImplementation(m).(ociartifact.OCIArtifactReferenceProvider); ok {
		im, err := o.GetOCIReference(cv)
		if err == nil {
			spec := cv.Repository().GetSpecification().AsUniformSpec(cv.GetContext())
			// not supported for fake repos
			if spec.Host != "" {
				im = spec.Host + "/" + im
			}
		}
		return im, err
	}
	return "", nil
}
