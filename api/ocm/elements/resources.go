package elements

import (
	"github.com/mandelsoft/goutils/errors"

	"ocm.software/ocm/api/ocm/compdesc"
	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/utils"
)

type ResourceMetaOption interface {
	ApplyToResourceMeta(*compdesc.ResourceMeta) error
}

func ResourceMeta(name, typ string, opts ...ResourceMetaOption) (*compdesc.ResourceMeta, error) {
	m := compdesc.NewResourceMeta(name, typ, metav1.LocalRelation)
	list := errors.ErrList()
	for _, o := range opts {
		if o != nil {
			list.Add(o.ApplyToResourceMeta(m))
		}
	}
	return m, list.Result()
}

////////////////////////////////////////////////////////////////////////////////

type local bool

func (o local) ApplyToResourceMeta(m *compdesc.ResourceMeta) error {
	if o {
		m.Relation = metav1.LocalRelation
	} else {
		m.Relation = metav1.ExternalRelation
	}
	return nil
}

// WithLocalRelation sets the resource relation to metav1.LocalRelation.
func WithLocalRelation(flag ...bool) ResourceMetaOption {
	return local(utils.OptionalDefaultedBool(true, flag...))
}

// WithExternalRelation sets the resource relation to metav1.ExternalRelation.
func WithExternalRelation(flag ...bool) ResourceMetaOption {
	return local(!utils.OptionalDefaultedBool(true, flag...))
}

////////////////////////////////////////////////////////////////////////////////

type srcref struct {
	ref     metav1.StringMap
	labels  metav1.Labels
	errlist errors.ErrorList
}

var _ ResourceMetaOption = (*srcref)(nil)

func (o *srcref) ApplyToResourceMeta(m *compdesc.ResourceMeta) error {
	if err := o.errlist.Result(); err != nil {
		return err
	}
	m.SourceRefs = append(m.SourceRefs, compdesc.SourceRef{IdentitySelector: o.ref.Copy(), Labels: o.labels.Copy()})
	return nil
}

func (o *srcref) WithLabel(name string, value interface{}, opts ...metav1.LabelOption) *srcref {
	r := &srcref{ref: o.ref, labels: o.labels.Copy()}
	r.errlist.Add(r.labels.Set(name, value, opts...))
	return r
}

// WithSourceRef adds a source reference to a resource meta object.
// this is a sequence of name/value pairs.
// Optionally, additional labels can be added with srcref.WithLabel.
func WithSourceRef(sel ...string) *srcref {
	return &srcref{ref: metav1.StringMap(metav1.NewExtraIdentity(sel...))}
}
