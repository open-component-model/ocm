package repocpi

import (
	"io"

	"github.com/mandelsoft/goutils/errors"

	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/utils"
	"ocm.software/ocm/api/utils/refmgmt"
	"ocm.software/ocm/api/utils/refmgmt/resource"
)

type ComponentAccessInfo struct {
	Impl ComponentAccessImpl
	Kind string
	Main bool
}

type RepositoryImpl interface {
	SetBridge(bridge RepositoryBridge)

	GetContext() cpi.Context

	IsReadOnly() bool
	SetReadOnly()

	GetSpecification() cpi.RepositorySpec
	ComponentLister() cpi.ComponentLister

	ExistsComponentVersion(name string, version string) (bool, error)
	LookupComponent(name string) (*ComponentAccessInfo, error)

	io.Closer
}

// Chunked is an optional interface, which
// may be implemented to accept a blob limit for mapping
// local blobs to an external storage system.
type Chunked interface {
	// SetBlobLimit sets the blob limit if possible.
	// It returns true, if this was successful.
	SetBlobLimit(s int64) bool
}

// SetBlobLimit tries to set a blob limit for a repository
// implementation. It returns true, if this was possible.
func SetBlobLimit(i RepositoryImpl, s int64) bool {
	if c, ok := i.(Chunked); ok {
		return c.SetBlobLimit(s)
	}
	return false
}

type _repositoryBridgeBase = resource.ResourceImplBase[cpi.Repository]

type repositoryBridge struct {
	*_repositoryBridgeBase
	ctx  cpi.Context
	kind string
	impl RepositoryImpl
}

var _ utils.Unwrappable = (*repositoryBridge)(nil)

func newRepositoryBridge(impl RepositoryImpl, kind string, closer ...io.Closer) RepositoryBridge {
	base := resource.NewSimpleResourceImplBase[cpi.Repository](closer...)
	b := &repositoryBridge{
		_repositoryBridgeBase: base,
		ctx:                   impl.GetContext(),
		impl:                  impl,
	}
	impl.SetBridge(b)
	return b
}

func (b *repositoryBridge) Close() error {
	list := errors.ErrListf("closing %s", b.kind)
	refmgmt.AllocLog.Trace("closing repository bridge", "kind", b.kind)
	list.Add(b.impl.Close())
	list.Add(b._repositoryBridgeBase.Close())
	refmgmt.AllocLog.Trace("closed repository bridge", "kind", b.kind)
	return list.Result()
}

func (b *repositoryBridge) GetContext() cpi.Context {
	return b.ctx
}

func (b *repositoryBridge) IsReadOnly() bool {
	return b.impl.IsReadOnly()
}

func (b *repositoryBridge) SetReadOnly() {
	b.impl.SetReadOnly()
}

func (b *repositoryBridge) Unwrap() interface{} {
	return b.impl
}

func (b *repositoryBridge) GetSpecification() cpi.RepositorySpec {
	return b.impl.GetSpecification()
}

func (b *repositoryBridge) ComponentLister() cpi.ComponentLister {
	return b.impl.ComponentLister()
}

func (b *repositoryBridge) ExistsComponentVersion(name string, version string) (bool, error) {
	return b.impl.ExistsComponentVersion(name, version)
}

func (b *repositoryBridge) LookupComponentVersion(name string, version string) (cv cpi.ComponentVersionAccess, rerr error) {
	c, err := b.LookupComponent(name)
	if err != nil {
		return nil, err
	}
	defer refmgmt.PropagateCloseTemporary(&rerr, c) // temporary component object not exposed.
	refmgmt.AllocLog.Trace("lookup version for temporary component ref", "component", name, "version", version)
	return c.LookupVersion(version)
}

func (b *repositoryBridge) LookupComponent(name string) (cpi.ComponentAccess, error) {
	i, err := b.impl.LookupComponent(name)
	if err != nil {
		return nil, err
	}
	return NewComponentAccess(i.Impl, i.Kind, i.Main)
}
