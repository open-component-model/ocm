//go:build !ignore_autogenerated
// +build !ignore_autogenerated

// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

// Code generated by controller-gen. DO NOT EDIT.

package v3alpha1

import (
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ComponentVersionSpec) DeepCopyInto(out *ComponentVersionSpec) {
	*out = *in
	if in.Sources != nil {
		in, out := &in.Sources, &out.Sources
		*out = make(Sources, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.References != nil {
		in, out := &in.References, &out.References
		*out = make(References, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Resources != nil {
		in, out := &in.Resources, &out.Resources
		*out = make(Resources, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ComponentVersionSpec.
func (in *ComponentVersionSpec) DeepCopy() *ComponentVersionSpec {
	if in == nil {
		return nil
	}
	out := new(ComponentVersionSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ElementMeta) DeepCopyInto(out *ElementMeta) {
	*out = *in
	if in.ExtraIdentity != nil {
		in, out := &in.ExtraIdentity, &out.ExtraIdentity
		*out = make(v1.Identity, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.Labels != nil {
		in, out := &in.Labels, &out.Labels
		*out = make(v1.Labels, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ElementMeta.
func (in *ElementMeta) DeepCopy() *ElementMeta {
	if in == nil {
		return nil
	}
	out := new(ElementMeta)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Reference) DeepCopyInto(out *Reference) {
	*out = *in
	in.ElementMeta.DeepCopyInto(&out.ElementMeta)
	if in.Digest != nil {
		in, out := &in.Digest, &out.Digest
		*out = new(v1.DigestSpec)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Reference.
func (in *Reference) DeepCopy() *Reference {
	if in == nil {
		return nil
	}
	out := new(Reference)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Resource) DeepCopyInto(out *Resource) {
	*out = *in
	in.ElementMeta.DeepCopyInto(&out.ElementMeta)
	if in.SourceRefs != nil {
		in, out := &in.SourceRefs, &out.SourceRefs
		*out = make([]SourceRef, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.SourceRef != nil {
		in, out := &in.SourceRef, &out.SourceRef
		*out = make([]SourceRef, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Access != nil {
		in, out := &in.Access, &out.Access
		*out = (*in).DeepCopy()
	}
	if in.Digest != nil {
		in, out := &in.Digest, &out.Digest
		*out = new(v1.DigestSpec)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Resource.
func (in *Resource) DeepCopy() *Resource {
	if in == nil {
		return nil
	}
	out := new(Resource)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Source) DeepCopyInto(out *Source) {
	*out = *in
	in.SourceMeta.DeepCopyInto(&out.SourceMeta)
	if in.Access != nil {
		in, out := &in.Access, &out.Access
		*out = (*in).DeepCopy()
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Source.
func (in *Source) DeepCopy() *Source {
	if in == nil {
		return nil
	}
	out := new(Source)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SourceMeta) DeepCopyInto(out *SourceMeta) {
	*out = *in
	in.ElementMeta.DeepCopyInto(&out.ElementMeta)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SourceMeta.
func (in *SourceMeta) DeepCopy() *SourceMeta {
	if in == nil {
		return nil
	}
	out := new(SourceMeta)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SourceRef) DeepCopyInto(out *SourceRef) {
	*out = *in
	if in.IdentitySelector != nil {
		in, out := &in.IdentitySelector, &out.IdentitySelector
		*out = make(v1.StringMap, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.Labels != nil {
		in, out := &in.Labels, &out.Labels
		*out = make(v1.Labels, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SourceRef.
func (in *SourceRef) DeepCopy() *SourceRef {
	if in == nil {
		return nil
	}
	out := new(SourceRef)
	in.DeepCopyInto(out)
	return out
}
