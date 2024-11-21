package cpi

import (
	"strconv"

	"github.com/mandelsoft/goutils/errors"

	"ocm.software/ocm/api/ocm/compdesc"
	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/selectors/refsel"
	"ocm.software/ocm/api/ocm/selectors/rscsel"
	"ocm.software/ocm/api/ocm/selectors/srcsel"
)

type DummyComponentVersionAccess struct {
	Context Context
}

var _ ComponentVersionAccess = (*DummyComponentVersionAccess)(nil)

func (d *DummyComponentVersionAccess) GetContext() Context {
	return d.Context
}

func (d *DummyComponentVersionAccess) Close() error {
	return nil
}

func (d *DummyComponentVersionAccess) IsClosed() bool {
	return false
}

func (d *DummyComponentVersionAccess) IsReadOnly() bool {
	return true
}

func (d *DummyComponentVersionAccess) SetReadOnly() {
}

func (d *DummyComponentVersionAccess) Dup() (ComponentVersionAccess, error) {
	return d, nil
}

func (d *DummyComponentVersionAccess) GetProvider() *compdesc.Provider {
	return nil
}

func (d *DummyComponentVersionAccess) SetProvider(p *compdesc.Provider) error {
	return errors.ErrNotSupported()
}

func (d *DummyComponentVersionAccess) AdjustSourceAccess(meta *SourceMeta, acc compdesc.AccessSpec) error {
	return errors.ErrNotSupported()
}

func (c *DummyComponentVersionAccess) Repository() Repository {
	return nil
}

func (d *DummyComponentVersionAccess) GetName() string {
	return ""
}

func (d *DummyComponentVersionAccess) GetVersion() string {
	return ""
}

func (d *DummyComponentVersionAccess) GetDescriptor() *compdesc.ComponentDescriptor {
	return nil
}

func (d *DummyComponentVersionAccess) SelectResources(sel ...rscsel.Selector) ([]ResourceAccess, error) {
	return nil, nil
}

func (d *DummyComponentVersionAccess) GetResources() []ResourceAccess {
	return nil
}

func (d *DummyComponentVersionAccess) GetResource(id metav1.Identity) (ResourceAccess, error) {
	return nil, errors.ErrNotFound("resource", id.String())
}

func (d *DummyComponentVersionAccess) GetResourceIndex(metav1.Identity) int {
	return -1
}

func (d *DummyComponentVersionAccess) GetResourceByIndex(i int) (ResourceAccess, error) {
	return nil, errors.ErrInvalid("resource index", strconv.Itoa(i))
}

func (d *DummyComponentVersionAccess) SelectSources(sel ...srcsel.Selector) ([]SourceAccess, error) {
	return nil, nil
}

func (d *DummyComponentVersionAccess) GetSources() []SourceAccess {
	return nil
}

func (d *DummyComponentVersionAccess) GetSource(id metav1.Identity) (SourceAccess, error) {
	return nil, errors.ErrNotFound(KIND_SOURCE, id.String())
}

func (d *DummyComponentVersionAccess) GetSourceIndex(metav1.Identity) int {
	return -1
}

func (d *DummyComponentVersionAccess) GetSourceByIndex(i int) (SourceAccess, error) {
	return nil, errors.ErrInvalid("source index", strconv.Itoa(i))
}

func (d *DummyComponentVersionAccess) GetReference(meta metav1.Identity) (ComponentReference, error) {
	return ComponentReference{}, errors.ErrNotFound("reference", meta.String())
}

func (d *DummyComponentVersionAccess) SelectReferences(sel ...refsel.Selector) ([]ComponentReference, error) {
	return nil, nil
}

func (d *DummyComponentVersionAccess) GetReferences() []ComponentReference {
	return nil
}

func (d *DummyComponentVersionAccess) GetReferenceIndex(metav1.Identity) int {
	return -1
}

func (d *DummyComponentVersionAccess) GetReferenceByIndex(i int) (ComponentReference, error) {
	return ComponentReference{}, errors.ErrInvalid("reference index", strconv.Itoa(i))
}

func (d *DummyComponentVersionAccess) AccessMethod(spec AccessSpec) (AccessMethod, error) {
	if spec.IsLocal(d.Context) {
		return nil, errors.ErrNotSupported("local access method")
	}
	return spec.AccessMethod(d)
}

func (d *DummyComponentVersionAccess) Update() error {
	return errors.ErrNotSupported("update")
}

func (d *DummyComponentVersionAccess) Execute(f func() error) error {
	return f()
}

func (d *DummyComponentVersionAccess) AddBlob(blob BlobAccess, arttype, refName string, global AccessSpec, opts ...BlobUploadOption) (AccessSpec, error) {
	return nil, errors.ErrNotSupported("adding blobs")
}

func (d *DummyComponentVersionAccess) SetResourceBlob(meta *ResourceMeta, blob BlobAccess, refname string, global AccessSpec, opts ...BlobModificationOption) error {
	return errors.ErrNotSupported("adding blobs")
}

func (d *DummyComponentVersionAccess) AdjustResourceAccess(meta *ResourceMeta, acc compdesc.AccessSpec, opts ...ModificationOption) error {
	return errors.ErrNotSupported("resource modification")
}

func (d *DummyComponentVersionAccess) SetResource(meta *ResourceMeta, spec compdesc.AccessSpec, opts ...ModificationOption) error {
	return errors.ErrNotSupported("resource modification")
}

func (d *DummyComponentVersionAccess) SetResourceByAccess(art ResourceAccess, modopts ...BlobModificationOption) error {
	return errors.ErrNotSupported("resource modification")
}

func (d *DummyComponentVersionAccess) SetSourceBlob(meta *SourceMeta, blob BlobAccess, refname string, global AccessSpec, opts ...TargetElementOption) error {
	return errors.ErrNotSupported("source modification")
}

func (d *DummyComponentVersionAccess) SetSource(meta *SourceMeta, spec compdesc.AccessSpec, opts ...TargetElementOption) error {
	return errors.ErrNotSupported("source modification")
}

func (d *DummyComponentVersionAccess) SetSourceByAccess(art SourceAccess, opts ...TargetElementOption) error {
	return errors.ErrNotSupported()
}

func (d *DummyComponentVersionAccess) SetReference(ref *ComponentReference, opts ...ElementModificationOption) error {
	return errors.ErrNotSupported()
}

func (d *DummyComponentVersionAccess) DiscardChanges() {
}

func (d *DummyComponentVersionAccess) IsPersistent() bool {
	return false
}

func (d *DummyComponentVersionAccess) UseDirectAccess() bool {
	return true
}
