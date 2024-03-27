//go:build !ignore_autogenerated

// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

// Code generated by controller-gen. DO NOT EDIT.

package v1

import ()

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ArtefactDigest) DeepCopyInto(out *ArtefactDigest) {
	*out = *in
	if in.ExtraIdentity != nil {
		in, out := &in.ExtraIdentity, &out.ExtraIdentity
		*out = make(Identity, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	out.Digest = in.Digest
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ArtefactDigest.
func (in *ArtefactDigest) DeepCopy() *ArtefactDigest {
	if in == nil {
		return nil
	}
	out := new(ArtefactDigest)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in ArtefactDigests) DeepCopyInto(out *ArtefactDigests) {
	{
		in := &in
		*out = make(ArtefactDigests, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ArtefactDigests.
func (in ArtefactDigests) DeepCopy() ArtefactDigests {
	if in == nil {
		return nil
	}
	out := new(ArtefactDigests)
	in.DeepCopyInto(out)
	return *out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DigestSpec) DeepCopyInto(out *DigestSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DigestSpec.
func (in *DigestSpec) DeepCopy() *DigestSpec {
	if in == nil {
		return nil
	}
	out := new(DigestSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in Identity) DeepCopyInto(out *Identity) {
	{
		in := &in
		*out = make(Identity, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Identity.
func (in Identity) DeepCopy() Identity {
	if in == nil {
		return nil
	}
	out := new(Identity)
	in.DeepCopyInto(out)
	return *out
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Label.
func (in *Label) DeepCopy() *Label {
	if in == nil {
		return nil
	}
	out := new(Label)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in Labels) DeepCopyInto(out *Labels) {
	{
		in := &in
		*out = make(Labels, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Labels.
func (in Labels) DeepCopy() Labels {
	if in == nil {
		return nil
	}
	out := new(Labels)
	in.DeepCopyInto(out)
	return *out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Metadata) DeepCopyInto(out *Metadata) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Metadata.
func (in *Metadata) DeepCopy() *Metadata {
	if in == nil {
		return nil
	}
	out := new(Metadata)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NestedComponentDigests) DeepCopyInto(out *NestedComponentDigests) {
	*out = *in
	if in.Digest != nil {
		in, out := &in.Digest, &out.Digest
		*out = new(DigestSpec)
		**out = **in
	}
	if in.Resources != nil {
		in, out := &in.Resources, &out.Resources
		*out = make(ArtefactDigests, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NestedComponentDigests.
func (in *NestedComponentDigests) DeepCopy() *NestedComponentDigests {
	if in == nil {
		return nil
	}
	out := new(NestedComponentDigests)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in NestedDigests) DeepCopyInto(out *NestedDigests) {
	{
		in := &in
		*out = make(NestedDigests, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NestedDigests.
func (in NestedDigests) DeepCopy() NestedDigests {
	if in == nil {
		return nil
	}
	out := new(NestedDigests)
	in.DeepCopyInto(out)
	return *out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Signature) DeepCopyInto(out *Signature) {
	*out = *in
	out.Digest = in.Digest
	out.Signature = in.Signature
	if in.Timestamp != nil {
		in, out := &in.Timestamp, &out.Timestamp
		*out = new(TimestampSpec)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Signature.
func (in *Signature) DeepCopy() *Signature {
	if in == nil {
		return nil
	}
	out := new(Signature)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SignatureSpec) DeepCopyInto(out *SignatureSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SignatureSpec.
func (in *SignatureSpec) DeepCopy() *SignatureSpec {
	if in == nil {
		return nil
	}
	out := new(SignatureSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Timestamp) DeepCopyInto(out *Timestamp) {
	*out = *in
	in._time.DeepCopyInto(&out._time)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Timestamp.
func (in *Timestamp) DeepCopy() *Timestamp {
	if in == nil {
		return nil
	}
	out := new(Timestamp)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TimestampSpec) DeepCopyInto(out *TimestampSpec) {
	*out = *in
	if in.Time != nil {
		in, out := &in.Time, &out.Time
		*out = new(Timestamp)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TimestampSpec.
func (in *TimestampSpec) DeepCopy() *TimestampSpec {
	if in == nil {
		return nil
	}
	out := new(TimestampSpec)
	in.DeepCopyInto(out)
	return out
}
