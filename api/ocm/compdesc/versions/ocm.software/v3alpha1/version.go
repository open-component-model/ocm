package v3alpha1

import (
	"github.com/mandelsoft/goutils/errors"
	"ocm.software/ocm/api/ocm/compdesc"
	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/compdesc/versions/ocm.software/v3alpha1/jsonscheme"
	"ocm.software/ocm/api/utils/runtime"
)

const (
	SchemaVersion = GroupVersion

	VersionName  = "v3alpha1"
	GroupVersion = metav1.GROUP + "/" + VersionName
	Kind         = metav1.KIND
)

func init() {
	compdesc.RegisterScheme(&DescriptorVersion{})
}

type DescriptorVersion struct{}

var _ compdesc.Scheme = (*DescriptorVersion)(nil)

func (v *DescriptorVersion) GetVersion() string {
	return SchemaVersion
}

func (v *DescriptorVersion) Decode(data []byte, opts *compdesc.DecodeOptions) (compdesc.ComponentDescriptorVersion, error) {
	var cd ComponentDescriptor
	if !opts.DisableValidation {
		if err := jsonscheme.Validate(data); err != nil {
			return nil, err
		}
	}
	var err error
	if opts.StrictMode {
		err = opts.Codec.DecodeStrict(data, &cd)
	} else {
		err = opts.Codec.Decode(data, &cd)
	}
	if err != nil {
		return nil, err
	}

	if err := cd.Default(); err != nil {
		return nil, err
	}

	if !opts.DisableValidation {
		err = cd.Validate()
		if err != nil {
			return nil, err
		}
	}
	return &cd, err
}

////////////////////////////////////////////////////////////////////////////////
// convert to internal version
////////////////////////////////////////////////////////////////////////////////

func (v *DescriptorVersion) ConvertTo(obj compdesc.ComponentDescriptorVersion) (out *compdesc.ComponentDescriptor, err error) {
	if obj == nil {
		return nil, nil
	}
	in, ok := obj.(*ComponentDescriptor)
	if !ok {
		return nil, errors.Newf("%T is no version v2 descriptor", obj)
	}
	if in.Kind != Kind {
		return nil, errors.ErrInvalid("kind", in.Kind)
	}

	defer compdesc.CatchConversionError(&err)
	out = &compdesc.ComponentDescriptor{
		Metadata: compdesc.Metadata{ConfiguredVersion: in.APIVersion},
		ComponentSpec: compdesc.ComponentSpec{
			ObjectMeta:         *in.ObjectMeta.Copy(),
			RepositoryContexts: in.RepositoryContexts.Copy(),
			Sources:            convertSourcesTo(in.Spec.Sources),
			Resources:          convertResourcesTo(in.Spec.Resources),
			References:         convertReferencesTo(in.Spec.References),
		},
		Signatures:    in.Signatures.Copy(),
		NestedDigests: in.NestedDigests.Copy(),
	}
	return out, nil
}

func convertReferenceTo(in Reference) compdesc.Reference {
	return compdesc.Reference{
		ElementMeta:   convertElementmetaTo(in.ElementMeta),
		ComponentName: in.ComponentName,
		Digest:        in.Digest.Copy(),
	}
}

func convertReferencesTo(in []Reference) compdesc.References {
	out := make(compdesc.References, len(in))
	for i := range in {
		out[i] = convertReferenceTo(in[i])
	}
	return out
}

func convertArtifactTo(in Artifact) compdesc.Artifact {
	hints := make(metav1.ReferenceHints, len(in.ReferenceHints))

	for i, h := range in.ReferenceHints {
		hints[i] = h
	}
	return compdesc.Artifact{
		Access:         compdesc.GenericAccessSpec(in.Access.DeepCopy()),
		ReferenceHints: hints,
	}
}

func convertSourceTo(in Source) compdesc.Source {
	return compdesc.Source{
		SourceMeta: compdesc.SourceMeta{
			ElementMeta: convertElementmetaTo(in.ElementMeta),
			Type:        in.Type,
		},
		Artifact: convertArtifactTo(in.Artifact),
	}
}

func convertSourcesTo(in Sources) compdesc.Sources {
	if in == nil {
		return nil
	}
	out := make(compdesc.Sources, len(in))
	for i := range in {
		out[i] = convertSourceTo(in[i])
	}
	return out
}

func convertElementmetaTo(in ElementMeta) compdesc.ElementMeta {
	return compdesc.ElementMeta{
		Name:          in.Name,
		Version:       in.Version,
		ExtraIdentity: in.ExtraIdentity.Copy(),
		Labels:        in.Labels.Copy(),
	}
}

func convertResourceTo(in Resource) compdesc.Resource {
	srcRefs := ConvertSourcerefsTo(in.SourceRefs)
	if srcRefs == nil {
		srcRefs = ConvertSourcerefsTo(in.SourceRef)
	}
	return compdesc.Resource{
		ResourceMeta: compdesc.ResourceMeta{
			ElementMeta: convertElementmetaTo(in.ElementMeta),
			Type:        in.Type,
			Relation:    in.Relation,
			SourceRefs:  srcRefs,
			Digest:      in.Digest.Copy(),
		},
		Artifact: convertArtifactTo(in.Artifact),
	}
}

func convertResourcesTo(in Resources) compdesc.Resources {
	if in == nil {
		return nil
	}
	out := make(compdesc.Resources, len(in))
	for i := range in {
		out[i] = convertResourceTo(in[i])
	}
	return out
}

func convertSourcerefTo(in SourceRef) compdesc.SourceRef {
	return compdesc.SourceRef{
		IdentitySelector: in.IdentitySelector.Copy(),
		Labels:           in.Labels.Copy(),
	}
}

func ConvertSourcerefsTo(in []SourceRef) []compdesc.SourceRef {
	if in == nil {
		return nil
	}
	out := make([]compdesc.SourceRef, len(in))
	for i := range in {
		out[i] = convertSourcerefTo(in[i])
	}
	return out
}

////////////////////////////////////////////////////////////////////////////////
// convert from internal version
////////////////////////////////////////////////////////////////////////////////

func (v *DescriptorVersion) ConvertFrom(in *compdesc.ComponentDescriptor) (compdesc.ComponentDescriptorVersion, error) {
	if in == nil {
		return nil, nil
	}
	out := &ComponentDescriptor{
		TypeMeta: metav1.TypeMeta{
			APIVersion: SchemaVersion,
			Kind:       Kind,
		},
		ObjectMeta:         *in.ObjectMeta.Copy(),
		RepositoryContexts: in.RepositoryContexts.Copy(),
		Spec: ComponentVersionSpec{
			Sources:    convertSourcesFrom(in.Sources),
			Resources:  convertResourcesFrom(in.Resources),
			References: convertReferencesFrom(in.References),
		},
		Signatures:    in.Signatures.Copy(),
		NestedDigests: in.NestedDigests.Copy(),
	}
	if err := out.Default(); err != nil {
		return nil, err
	}
	return out, nil
}

func convertReferenceFrom(in compdesc.Reference) Reference {
	return Reference{
		ElementMeta:   convertElementmetaFrom(in.ElementMeta),
		ComponentName: in.ComponentName,
		Digest:        in.Digest.Copy(),
	}
}

func convertReferencesFrom(in []compdesc.Reference) []Reference {
	if in == nil {
		return nil
	}
	out := make([]Reference, len(in))
	for i := range in {
		out[i] = convertReferenceFrom(in[i])
	}
	return out
}

func convertArtifactFrom(in compdesc.Artifact) Artifact {
	acc, err := runtime.ToUnstructuredTypedObject(in.Access)
	if err != nil {
		compdesc.ThrowConversionError(err)
	}
	hints := make([]metav1.DefaultReferenceHint, len(in.ReferenceHints))
	for i, h := range in.ReferenceHints {
		hints[i] = h.AsDefault()
	}
	return Artifact{
		Access:         acc,
		ReferenceHints: hints,
	}
}

func convertSourceFrom(in compdesc.Source) Source {
	return Source{
		SourceMeta: SourceMeta{
			ElementMeta: convertElementmetaFrom(in.ElementMeta),
			Type:        in.Type,
		},
		Artifact: convertArtifactFrom(in.Artifact),
	}
}

func convertSourcesFrom(in compdesc.Sources) Sources {
	if in == nil {
		return nil
	}
	out := make(Sources, len(in))
	for i := range in {
		out[i] = convertSourceFrom(in[i])
	}
	return out
}

func convertElementmetaFrom(in compdesc.ElementMeta) ElementMeta {
	return ElementMeta{
		Name:          in.Name,
		Version:       in.Version,
		ExtraIdentity: in.ExtraIdentity.Copy(),
		Labels:        in.Labels.Copy(),
	}
}

func convertResourceFrom(in compdesc.Resource) Resource {
	return Resource{
		ElementMeta: convertElementmetaFrom(in.ElementMeta),
		Type:        in.Type,
		Relation:    in.Relation,
		SourceRefs:  convertSourcerefsFrom(in.SourceRefs),
		Artifact:    convertArtifactFrom(in.Artifact),
		Digest:      in.Digest.Copy(),
	}
}

func convertResourcesFrom(in compdesc.Resources) Resources {
	if in == nil {
		return nil
	}
	out := make(Resources, len(in))
	for i := range in {
		out[i] = convertResourceFrom(in[i])
	}
	return out
}

func convertSourcerefFrom(in compdesc.SourceRef) SourceRef {
	return SourceRef{
		IdentitySelector: in.IdentitySelector.Copy(),
		Labels:           in.Labels.Copy(),
	}
}

func convertSourcerefsFrom(in []compdesc.SourceRef) []SourceRef {
	if in == nil {
		return nil
	}
	out := make([]SourceRef, len(in))
	for i := range in {
		out[i] = convertSourcerefFrom(in[i])
	}
	return out
}
