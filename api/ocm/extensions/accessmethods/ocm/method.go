package ocm

import (
	"fmt"
	"io"
	"sync"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/goutils/finalizer"

	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/ocm/cpi/accspeccpi"
	"ocm.software/ocm/api/ocm/resourcerefs"
	"ocm.software/ocm/api/tech/helm"
	"ocm.software/ocm/api/utils/blobaccess/blobaccess"
	"ocm.software/ocm/api/utils/refmgmt"
	"ocm.software/ocm/api/utils/runtime"
)

// Type is the access type for a blob in an OCI repository.
const (
	Type   = "ocm"
	TypeV1 = Type + runtime.VersionSeparator + "v1"
)

func init() {
	accspeccpi.RegisterAccessType(accspeccpi.NewAccessSpecType[*AccessSpec](Type, accspeccpi.WithDescription(usage)))
	accspeccpi.RegisterAccessType(accspeccpi.NewAccessSpecType[*AccessSpec](TypeV1, accspeccpi.WithFormatSpec(formatV1), accspeccpi.WithConfigHandler(ConfigHandler())))
}

// New creates a new Helm Chart accessor for helm repositories.
func New(comp, vers string, repo cpi.RepositorySpec, id metav1.Identity, path ...metav1.Identity) (*AccessSpec, error) {
	spec, err := cpi.ToGenericRepositorySpec(repo)
	if err != nil {
		return nil, err
	}
	return &AccessSpec{
		ObjectVersionedType: runtime.NewVersionedTypedObject(Type),
		OCMRepository:       spec,
		Component:           comp,
		Version:             vers,
		ResourceRef:         metav1.NewNestedResourceRef(id, path),
	}, nil
}

// AccessSpec describes the access for an OCM repository.
type AccessSpec struct {
	runtime.ObjectVersionedType `json:",inline"`

	// OCMRepository is the URL of the OCM repository to load the chart from.
	OCMRepository *cpi.GenericRepositorySpec `json:"ocmRepository,omitempty"`

	// Component if the name of the root component used to lookup the resource.
	Component string `json:"component,omitempty"`

	// Version is the version og the root component.
	Version string `json:"version,omitempty,"`

	ResourceRef metav1.ResourceReference `json:"resourceRef"`
}

var _ accspeccpi.AccessSpec = (*AccessSpec)(nil)

func (a *AccessSpec) Describe(ctx accspeccpi.Context) string {
	comp := a.Component
	if a.Version != "" {
		comp += ":" + a.Version
	}
	if comp != "" {
		comp = " in " + comp
	}
	raw, _ := a.OCMRepository.GetRaw()
	if a.OCMRepository != nil {
		return fmt.Sprintf("OCM resource %s%s in repository %s", a.ResourceRef.String(), comp, string(raw))
	}
	return fmt.Sprintf("OCM resource %s%s", a.ResourceRef.String(), comp)
}

func (a *AccessSpec) IsLocal(ctx accspeccpi.Context) bool {
	return false
}

func (a *AccessSpec) GlobalAccessSpec(ctx accspeccpi.Context) accspeccpi.AccessSpec {
	return a
}

func (a *AccessSpec) AccessMethod(access accspeccpi.ComponentVersionAccess) (accspeccpi.AccessMethod, error) {
	return accspeccpi.AccessMethodForImplementation(&accessMethod{comp: access, spec: a}, nil)
}

///////////////////

func (a *AccessSpec) GetVersion() string {
	return a.Version
}

func (a *AccessSpec) GetComponent() string {
	return a.Component
}

////////////////////////////////////////////////////////////////////////////////

type accessMethod struct {
	lock sync.Mutex
	blob blobaccess.BlobAccess
	repo cpi.Repository
	comp accspeccpi.ComponentVersionAccess
	spec *AccessSpec
	acc  cpi.ResourceAccess
}

var (
	_ accspeccpi.AccessMethodImpl   = (*accessMethod)(nil)
	_ accspeccpi.DigestSpecProvider = (*accessMethod)(nil)
)

func (_ *accessMethod) IsLocal() bool {
	return false
}

func (m *accessMethod) GetKind() string {
	return Type
}

func (m *accessMethod) AccessSpec() accspeccpi.AccessSpec {
	return m.spec
}

func (m *accessMethod) Close() error {
	m.lock.Lock()
	defer m.lock.Unlock()
	if m.blob != nil {
		m.blob.Close()
		m.blob = nil
	}
	if m.repo != nil {
		m.repo.Close()
		m.repo = nil
	}
	return nil
}

func (m *accessMethod) Get() ([]byte, error) {
	return blobaccess.BlobData(m.getBlob())
}

func (m *accessMethod) Reader() (io.ReadCloser, error) {
	return blobaccess.BlobReader(m.getBlob())
}

func (m *accessMethod) MimeType() string {
	return helm.ChartMediaType
}

func (m *accessMethod) getBlob() (bacc blobaccess.BlobAccess, efferr error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	var finalize finalizer.Finalizer
	defer finalize.FinalizeWithErrorPropagation(&efferr)

	if m.blob != nil {
		return m.blob, nil
	}

	vers := m.spec.GetVersion()
	name := m.spec.GetComponent()

	if vers == "" {
		vers = m.comp.GetVersion()
	}
	if name == "" {
		vers = m.comp.GetName()
	}

	var err error

	var cv cpi.ComponentVersionAccess
	if name == m.comp.GetName() && vers == m.comp.GetVersion() {
		cv = m.comp
		if m.repo == nil {
			m.repo, err = cv.Repository().Dup()
			if err != nil {
				return nil, err
			}
		}
	} else {
		if m.repo == nil {
			if m.spec.OCMRepository == nil {
				m.repo, err = m.comp.Repository().Dup()
				if err != nil {
					return nil, err
				}
			} else {
				m.repo, err = refmgmt.ToLazy(m.comp.GetContext().RepositoryForSpec(m.spec.OCMRepository))
				if err != nil {
					return nil, err
				}
			}
		}

		cv, err = refmgmt.ToLazy(m.repo.LookupComponentVersion(name, vers))
		if errors.IsErrNotFound(err) || cv == nil {
			r := m.comp.GetContext().GetResolver()
			if r != nil {
				cv, err = refmgmt.ToLazy(r.LookupComponentVersion(name, vers))
			}
		}
		if err != nil {
			return nil, err
		}
		finalize.Close(cv)
	}
	if cv == nil {
		return nil, errors.ErrNotFound(cpi.KIND_COMPONENTVERSION, name+":"+vers)
	}

	r, eff, err := resourcerefs.ResolveResourceReference(cv, m.spec.ResourceRef, m.comp.GetContext().GetResolver())
	if err != nil {
		return nil, err
	}
	finalize.Close(refmgmt.AsLazy(eff))

	m.blob, err = r.BlobAccess()
	m.acc = r
	return m.blob, err
}

func (m *accessMethod) GetDigestSpec() (*metav1.DigestSpec, error) {
	_, err := m.getBlob()
	if err != nil {
		return nil, err
	}
	return m.acc.Meta().Digest, nil
}
