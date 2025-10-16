package compdesc

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/mandelsoft/goutils/errors"

	"ocm.software/ocm/api/ocm/compdesc/equivalent"
	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/selectors/accessors"
	"ocm.software/ocm/api/utils/errkind"
	"ocm.software/ocm/api/utils/runtime"
	"ocm.software/ocm/api/utils/semverutils"
)

const InternalSchemaVersion = "internal"

// Deprecated: as result of the new select function an empty list is returned instead of an error.
var NotFound = errors.ErrNotFound()

const (
	KIND_COMPONENT        = errkind.KIND_COMPONENT
	KIND_COMPONENTVERSION = "component version"
	KIND_RESOURCE         = "component resource"
	KIND_SOURCE           = "component source"
	KIND_REFERENCE        = "component reference"
)

const ComponentDescriptorFileName = "component-descriptor.yaml"

// Metadata defines the configured metadata of the component descriptor.
// It is taken from the original serialization format. It can be set
// to define a default serialization version.
type Metadata struct {
	ConfiguredVersion string `json:"configuredSchemaVersion"`
}

// ComponentDescriptor defines a versioned component with a source and dependencies.
// +k8s:deepcopy-gen=true
// +k8s:openapi-gen=true
type ComponentDescriptor struct {
	// Metadata specifies the schema version of the component.
	Metadata Metadata `json:"meta"`
	// Spec contains the specification of the component.
	ComponentSpec `json:"component"`
	// Signatures contains a list of signatures for the ComponentDescriptor
	Signatures metav1.Signatures `json:"signatures,omitempty"`

	// NestedDigets describe calculated resource digests for aggregated
	// component versions, which might not be persisted, but incorporated
	// into signatures of the actual component version
	NestedDigests metav1.NestedDigests `json:"nestedDigests,omitempty"`
}

func New(name, version string) *ComponentDescriptor {
	return DefaultComponent(&ComponentDescriptor{
		Metadata: Metadata{
			ConfiguredVersion: "v2",
		},
		ComponentSpec: ComponentSpec{
			ObjectMeta: metav1.ObjectMeta{
				Name:    name,
				Version: version,
				Provider: metav1.Provider{
					Name: "acme",
				},
			},
		},
	})
}

// SchemaVersion returns the scheme version configured in the representation.
func (cd *ComponentDescriptor) SchemaVersion() string {
	return cd.Metadata.ConfiguredVersion
}

func (cd *ComponentDescriptor) Copy() *ComponentDescriptor {
	out := &ComponentDescriptor{
		Metadata: cd.Metadata,
		ComponentSpec: ComponentSpec{
			ObjectMeta:         *cd.ObjectMeta.Copy(),
			RepositoryContexts: cd.RepositoryContexts.Copy(),
			Sources:            cd.Sources.Copy(),
			References:         cd.References.Copy(),
			Resources:          cd.Resources.Copy(),
		},
		Signatures:    cd.Signatures.Copy(),
		NestedDigests: cd.NestedDigests.Copy(),
	}
	return out
}

func (cd *ComponentDescriptor) Reset() {
	cd.Provider.Name = ""
	cd.Provider.Labels = nil
	cd.Resources = nil
	cd.Sources = nil
	cd.References = nil
	cd.RepositoryContexts = nil
	cd.Signatures = nil
	cd.Labels = nil
	DefaultComponent(cd)
}

// ComponentSpec defines a virtual component with
// a repository context, source and dependencies.
// +k8s:deepcopy-gen=true
// +k8s:openapi-gen=true
type ComponentSpec struct {
	metav1.ObjectMeta `json:",inline"`
	// RepositoryContexts defines the previous repositories of the component
	RepositoryContexts runtime.UnstructuredTypedObjectList `json:"repositoryContexts"`
	// Sources defines sources that produced the component
	Sources Sources `json:"sources"`
	// References references component dependencies that can be resolved in the current context.
	References References `json:"componentReferences"`
	// Resources defines all resources that are created by the component and by a third party.
	Resources Resources `json:"resources"`
}

const (
	SystemIdentityName    = metav1.SystemIdentityName
	SystemIdentityVersion = metav1.SystemIdentityVersion
)

type ElementMetaAccess interface {
	GetName() string
	GetVersion() string
	GetIdentity(accessor ElementListAccessor) metav1.Identity
	GetLabels() metav1.Labels
}

type ArtifactMetaAccess interface {
	ElementMetaAccess
	GetType() string
	SetType(string)
}

// ArtifactMetaPointer is a pointer to an artifact meta object.
type ArtifactMetaPointer[P any] interface {
	ArtifactMetaAccess
	*P
}

// ElementMeta defines a object that is uniquely identified by its identity.
// +k8s:deepcopy-gen=true
// +k8s:openapi-gen=true
type ElementMeta struct {
	// Name is the context unique name of the object.
	Name string `json:"name"`
	// Version is the semver version of the object.
	Version string `json:"version"`
	// ExtraIdentity is the identity of an object.
	// An additional label with key "name" is not allowed
	ExtraIdentity metav1.Identity `json:"extraIdentity,omitempty"`
	// Labels defines an optional set of additional labels
	// describing the object.
	// +optional
	Labels metav1.Labels `json:"labels,omitempty"`
}

// GetName returns the name of the object.
func (o *ElementMeta) GetName() string {
	return o.Name
}

// GetMeta returns the element meta.
func (r *ElementMeta) GetMeta() accessors.ElementMeta {
	return r
}

// GetExtraIdentity returns the extra identity of the object.
func (o *ElementMeta) GetExtraIdentity() metav1.Identity {
	if o.ExtraIdentity == nil {
		return metav1.Identity{}
	}
	return o.ExtraIdentity.Copy()
}

// SetName sets the name of the object.
func (o *ElementMeta) SetName(name string) {
	o.Name = name
}

// GetVersion returns the version of the object.
func (o *ElementMeta) GetVersion() string {
	return o.Version
}

// SetVersion sets the version of the object.
func (o *ElementMeta) SetVersion(version string) {
	o.Version = version
}

// GetLabels returns the label of the object.
func (o *ElementMeta) GetLabels() metav1.Labels {
	return o.Labels
}

// SetLabels sets the labels of the object.
func (o *ElementMeta) SetLabels(labels []metav1.Label) {
	o.Labels = labels
}

// SetLabel sets a single label to an effective value.
// If the value is no byte slice, it is marshaled.
func (o *ElementMeta) SetLabel(name string, value interface{}, opts ...metav1.LabelOption) error {
	return o.Labels.Set(name, value, opts...)
}

// RemoveLabel removes a single label.
func (o *ElementMeta) RemoveLabel(name string) bool {
	return o.Labels.Remove(name)
}

// SetExtraIdentity sets the identity of the object.
func (o *ElementMeta) SetExtraIdentity(identity metav1.Identity) {
	o.ExtraIdentity = identity
}

func (o *ElementMeta) AddExtraIdentity(identity metav1.Identity) {
	if o.ExtraIdentity == nil {
		o.ExtraIdentity = identity
	} else {
		o.ExtraIdentity = o.ExtraIdentity.Copy()
		for k, v := range identity {
			o.ExtraIdentity[k] = v
		}
	}
}

// GetIdentity returns the identity of the object.
func (o *ElementMeta) GetIdentity(accessor ElementListAccessor) metav1.Identity {
	identity := o.ExtraIdentity.Copy()
	if identity == nil {
		identity = metav1.Identity{}
	}
	identity[SystemIdentityName] = o.Name
	if identity.Get(SystemIdentityVersion) == "" && accessor != nil {
		found := false
		l := accessor.Len()
		for i := 0; i < l; i++ {
			m := accessor.Get(i).GetMeta()
			if m.GetName() == o.Name {
				mid := m.GetExtraIdentity()
				mid.Remove(SystemIdentityVersion)
				if mid.Equals(o.ExtraIdentity) {
					if found {
						identity[SystemIdentityVersion] = o.Version
						break
					}
					found = true
				}
			}
		}
	}
	return identity
}

// GetRawIdentity returns the identity plus version, if set.
func (o *ElementMeta) GetRawIdentity() metav1.Identity {
	identity := o.ExtraIdentity.Copy()
	if identity == nil {
		identity = metav1.Identity{}
	}
	identity[SystemIdentityName] = o.Name
	if o.Version != "" {
		identity[SystemIdentityVersion] = o.Version
	}
	return identity
}

// GetMatchBaseIdentity returns all possible identity attributes for resource matching.
func (o *ElementMeta) GetMatchBaseIdentity() metav1.Identity {
	identity := o.ExtraIdentity.Copy()
	if identity == nil {
		identity = metav1.Identity{}
	}
	identity[SystemIdentityName] = o.Name
	identity[SystemIdentityVersion] = o.Version

	return identity
}

// GetIdentityDigest returns the digest of the object's identity.
func (o *ElementMeta) GetIdentityDigest(accessor ElementListAccessor) []byte {
	return o.GetIdentity(accessor).Digest()
}

func (o *ElementMeta) Copy() *ElementMeta {
	if o == nil {
		return nil
	}
	return &ElementMeta{
		Name:          o.Name,
		Version:       o.Version,
		ExtraIdentity: o.ExtraIdentity.Copy(),
		Labels:        o.Labels.Copy(),
	}
}

func (o *ElementMeta) Equivalent(a *ElementMeta) equivalent.EqualState {
	if o == a {
		return equivalent.StateEquivalent()
	}
	if o == nil {
		o, a = a, o
	}
	if a == nil {
		return o.Labels.Equivalent(nil)
	}

	state := equivalent.StateLocalHashEqual(a.Name == o.Name && a.Version == o.Version && reflect.DeepEqual(a.ExtraIdentity, o.ExtraIdentity))
	return state.Apply(o.Labels.Equivalent(a.Labels))
}

func GetByIdentity(a ElementListAccessor, id metav1.Identity) ElementMetaAccessor {
	l := a.Len()
	for i := 0; i < l; i++ {
		e := a.Get(i)
		if e.GetMeta().GetIdentity(a).Equals(id) {
			return e
		}
	}
	return nil
}

func GetIndexByIdentity(a ElementListAccessor, id metav1.Identity) int {
	l := a.Len()
	for i := 0; i < l; i++ {
		e := a.Get(i)
		if e.GetMeta().GetIdentity(a).Equals(id) {
			return i
		}
	}
	return -1
}

// ArtifactAccess provides access to a dedicated kind of artifact set
// in the component descriptor (resources or sources).
type ArtifactAccess func(cd *ComponentDescriptor) ArtifactAccessor

// GenericAccessSpec returns a generic AccessSpec implementation for an unstructured object.
// It can always be used instead of a dedicated access spec implementation. The core
// methods will map these spec into effective ones before an access is returned to the caller.
func GenericAccessSpec(un *runtime.UnstructuredTypedObject) AccessSpec {
	return &runtime.UnstructuredVersionedTypedObject{
		*un.DeepCopy(),
	}
}

// Sources describes a set of source specifications.
type Sources []Source

var _ ElementListAccessor = Sources{}

func SourceArtifacts(cd *ComponentDescriptor) ArtifactAccessor {
	return cd.Sources
}

func (r Sources) Equivalent(o Sources) equivalent.EqualState {
	return EquivalentElems(r, o)
}

func (s Sources) Len() int {
	return len(s)
}

func (s Sources) Get(i int) ElementMetaAccessor {
	return &s[i]
}

func (s Sources) GetArtifact(i int) ElementArtifactAccessor {
	return &s[i]
}

func (s Sources) Copy() Sources {
	if s == nil {
		return nil
	}
	out := make(Sources, len(s))
	for i, v := range s {
		out[i] = *v.Copy()
	}
	return out
}

// Source is the definition of a component's source.
// +k8s:deepcopy-gen=true
// +k8s:openapi-gen=true
type Source struct {
	SourceMeta `json:",inline"`
	Access     AccessSpec `json:"access"`
}

func (s *Source) GetAccess() AccessSpec {
	return s.Access
}

func (r *Source) SetAccess(a AccessSpec) {
	r.Access = a
}

func (r *Source) Equivalent(e ElementMetaAccessor) equivalent.EqualState {
	if o, ok := e.(*Source); !ok {
		return equivalent.StateNotLocalHashEqual()
	} else {
		state := equivalent.StateLocalHashEqual(r.Type == o.Type)
		return state.Apply(
			r.ElementMeta.Equivalent(&o.ElementMeta),
		)
	}
}

func (s *Source) Copy() *Source {
	return &Source{
		SourceMeta: *s.SourceMeta.Copy(),
		Access:     s.Access,
	}
}

// SourceMeta is the definition of the meta data of a source.
// +k8s:deepcopy-gen=true
// +k8s:openapi-gen=true
type SourceMeta struct {
	ElementMeta
	// Type describes the type of the object.
	Type string `json:"type"`
}

// GetType returns the type of the object.
func (o *SourceMeta) GetType() string {
	return o.Type
}

// SetType sets the type of the object.
func (o *SourceMeta) SetType(ttype string) {
	o.Type = ttype
}

// Copy copies a source meta.
func (o *SourceMeta) Copy() *SourceMeta {
	if o == nil {
		return nil
	}
	return &SourceMeta{
		ElementMeta: *o.ElementMeta.Copy(),
		Type:        o.Type,
	}
}

func (o *SourceMeta) WithVersion(v string) *SourceMeta {
	r := *o
	r.Version = v
	return &r
}

func (o *SourceMeta) WithExtraIdentity(extras ...string) *SourceMeta {
	r := *o
	r.AddExtraIdentity(NewExtraIdentity(extras...))
	return &r
}

func (o *SourceMeta) WithLabel(l *Label) *SourceMeta {
	r := *o
	r.Labels.SetDef(l.Name, l)
	return &r
}

func NewSourceMeta(name, typ string) *SourceMeta {
	return &SourceMeta{
		ElementMeta: ElementMeta{Name: name},
		Type:        typ,
	}
}

// SourceRef defines a reference to a source
// +k8s:deepcopy-gen=true
// +k8s:openapi-gen=true
type SourceRef struct {
	// IdentitySelector defines the identity that is used to match a source.
	IdentitySelector metav1.StringMap `json:"identitySelector,omitempty"`
	// Labels defines an optional set of additional labels
	// describing the object.
	// +optional
	Labels metav1.Labels `json:"labels,omitempty"`
}

// Copy copy a source ref.
func (r *SourceRef) Copy() *SourceRef {
	if r == nil {
		return nil
	}
	return &SourceRef{
		IdentitySelector: r.IdentitySelector.Copy(),
		Labels:           r.Labels.Copy(),
	}
}

type SourceRefs []SourceRef

// Copy copies a list of source refs.
func (r SourceRefs) Copy() SourceRefs {
	if r == nil {
		return nil
	}

	result := make(SourceRefs, len(r))
	for i, v := range r {
		result[i] = *v.Copy()
	}
	return result
}

// Resources describes a set of resource specifications.
type Resources []Resource

var _ ElementListAccessor = Resources{}

func ResourceArtifacts(cd *ComponentDescriptor) ArtifactAccessor {
	return cd.Resources
}

func (r Resources) Equivalent(o Resources) equivalent.EqualState {
	return EquivalentElems(r, o)
}

func (r Resources) Len() int {
	return len(r)
}

func (r Resources) Get(i int) ElementMetaAccessor {
	return &r[i]
}

func (r Resources) GetArtifact(i int) ElementArtifactAccessor {
	return &r[i]
}

func (r Resources) Copy() Resources {
	if r == nil {
		return nil
	}
	out := make(Resources, len(r))
	for i, v := range r {
		out[i] = *v.Copy()
	}
	return out
}

func (r Resources) HaveDigests() bool {
	for _, e := range r {
		if e.Digest == nil {
			return false
		}
	}
	return true
}

// Resource describes a resource dependency of a component.
// +k8s:deepcopy-gen=true
// +k8s:openapi-gen=true
type Resource struct {
	ResourceMeta `json:",inline"`
	// Access describes the type specific method to
	// access the defined resource.
	Access AccessSpec `json:"access"`
}

func (r *Resource) GetAccess() AccessSpec {
	return r.Access
}

func (r *Resource) SetAccess(a AccessSpec) {
	r.Access = a
}

func (r *Resource) GetDigest() *metav1.DigestSpec {
	return r.Digest
}

func (r *Resource) SetDigest(d *metav1.DigestSpec) {
	r.Digest = d
}

func (r *Resource) GetRelation() metav1.ResourceRelation {
	return r.Relation
}

func (r *Resource) Equivalent(e ElementMetaAccessor) equivalent.EqualState {
	if o, ok := e.(*Resource); !ok {
		state := equivalent.StateNotLocalHashEqual()
		if r.Digest.IsExcluded() || IsNoneAccess(r.Access) {
			return state
		} else {
			state = state.Apply(equivalent.StateNotArtifactEqual(r.Digest != nil))
		}
		return state
	} else {
		// not delegated to ResourceMeta, because the significance of digests can only be determined at the Resource level.
		state := equivalent.StateLocalHashEqual(r.Type == o.Type && r.Relation == o.Relation && reflect.DeepEqual(r.SourceRefs, o.SourceRefs))

		if !IsNoneAccess(r.Access) || !IsNoneAccess(o.Access) {
			state = state.Apply(r.Digest.Equivalent(o.Digest))
		}
		return state.Apply(r.ElementMeta.Equivalent(&o.ElementMeta))
	}
}

func (r *Resource) Copy() *Resource {
	return &Resource{
		ResourceMeta: *r.ResourceMeta.Copy(),
		Access:       r.Access,
	}
}

// ResourceMeta describes the meta data of a resource.
// +k8s:deepcopy-gen=true
// +k8s:openapi-gen=true
type ResourceMeta struct {
	ElementMeta `json:",inline"`

	// Type describes the type of the object.
	Type string `json:"type"`

	// Relation describes the relation of the resource to the component.
	// Can be a local or external resource
	Relation metav1.ResourceRelation `json:"relation,omitempty"`

	// SourceRefs defines a list of source names.
	// These entries reference the sources defined in the
	// component.sources.
	SourceRefs SourceRefs `json:"srcRefs,omitempty"`

	// Digest is the optional digest of the referenced resource.
	// +optional
	Digest *metav1.DigestSpec `json:"digest,omitempty"`
}

// Fresh returns a digest-free copy.
func (o *ResourceMeta) Fresh() *ResourceMeta {
	n := o.Copy()
	n.Digest = nil
	return n
}

// GetType returns the type of the object.
func (o *ResourceMeta) GetType() string {
	return o.Type
}

// SetType sets the type of the object.
func (o *ResourceMeta) SetType(ttype string) {
	o.Type = ttype
}

// SetDigest sets the digest of the object.
func (o *ResourceMeta) SetDigest(d *metav1.DigestSpec) {
	o.Digest = d
}

// Copy copies a resource meta.
func (o *ResourceMeta) Copy() *ResourceMeta {
	if o == nil {
		return nil
	}
	r := &ResourceMeta{
		ElementMeta: *o.ElementMeta.Copy(),
		Type:        o.Type,
		Relation:    o.Relation,
		SourceRefs:  o.SourceRefs.Copy(),
		Digest:      o.Digest.Copy(),
	}
	return r
}

func (o *ResourceMeta) WithVersion(v string) *ResourceMeta {
	r := *o
	r.Version = v
	return &r
}

func (o *ResourceMeta) WithExtraIdentity(extras ...string) *ResourceMeta {
	r := *o
	r.AddExtraIdentity(NewExtraIdentity(extras...))
	return &r
}

func (o *ResourceMeta) WithLabel(l *Label) *ResourceMeta {
	r := *o
	r.Labels.SetDef(l.Name, l)
	return &r
}

func NewResourceMeta(name string, typ string, relation metav1.ResourceRelation) *ResourceMeta {
	return &ResourceMeta{
		ElementMeta: ElementMeta{Name: name},
		Type:        typ,
		Relation:    relation,
	}
}

type References []Reference

func (r References) Equivalent(o References) equivalent.EqualState {
	return EquivalentElems(r, o)
}

func (r References) Len() int {
	return len(r)
}

func (r References) Get(i int) ElementMetaAccessor {
	return &r[i]
}

func (r References) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

func (r References) Less(i, j int) bool {
	c := strings.Compare(r[i].Name, r[j].Name)
	if c != 0 {
		return c < 0
	}
	return semverutils.Compare(r[i].Version, r[j].Version) < 0
}

func (r References) Copy() References {
	if r == nil {
		return nil
	}
	out := make(References, len(r))
	for i, v := range r {
		out[i] = *v.Copy()
	}
	return out
}

// Reference describes the reference to another component in the registry.
// +k8s:deepcopy-gen=true
// +k8s:openapi-gen=true
type Reference struct {
	ElementMeta `json:",inline"`
	// ComponentName describes the remote name of the referenced object
	ComponentName string `json:"componentName"`
	// Digest is the optional digest of the referenced component.
	// +optional
	Digest *metav1.DigestSpec `json:"digest,omitempty"`
}

func NewComponentReference(name, componentName, version string, extraIdentity metav1.Identity) *Reference {
	return &Reference{
		ElementMeta: ElementMeta{
			Name:          name,
			Version:       version,
			ExtraIdentity: extraIdentity,
		},
		ComponentName: componentName,
	}
}

func (r Reference) String() string {
	return fmt.Sprintf("%s[%s:%s]", r.Name, r.ComponentName, r.Version)
}

// WithVersion returns a new reference with a dedicated version.
func (o *Reference) WithVersion(v string) *Reference {
	n := o.Copy()
	n.Version = v
	return n
}

// WithExtraIdentity returns a new reference with a dedicated version.
func (o *Reference) WithExtraIdentity(extras ...string) *Reference {
	n := o.Copy()
	n.AddExtraIdentity(NewExtraIdentity(extras...))
	return n
}

// Fresh returns a digest-free copy.
func (o *Reference) Fresh() *Reference {
	n := o.Copy()
	n.Digest = nil
	return n
}

func (r *Reference) GetDigest() *metav1.DigestSpec {
	return r.Digest
}

func (r *Reference) SetDigest(d *metav1.DigestSpec) {
	r.Digest = d
}

func (r *Reference) Equivalent(e ElementMetaAccessor) equivalent.EqualState {
	if o, ok := e.(*Reference); !ok {
		state := equivalent.StateNotLocalHashEqual()
		if r.Digest != nil {
			state = state.Apply(equivalent.StateNotArtifactEqual(true))
		}
		return state
	} else {
		state := equivalent.StateLocalHashEqual(r.Name == o.Name && r.Version == o.Version && r.ComponentName == o.ComponentName)
		// TODO: how to handle digests
		if r.Digest != nil && o.Digest != nil { // hmm, digest described more than the local component, should we use it at all?
			state = state.Apply(r.Digest.Equivalent(o.Digest))
		} else if r.Digest != o.Digest { // not both are nil
			state = state.NotEquivalent()
		}

		return state.Apply(
			r.ElementMeta.Equivalent(&o.ElementMeta),
		)
	}
}

func (r *Reference) GetComponentName() string {
	return r.ComponentName
}

func (r *Reference) Copy() *Reference {
	return &Reference{
		ElementMeta:   *r.ElementMeta.Copy(),
		ComponentName: r.ComponentName,
		Digest:        r.Digest.Copy(),
	}
}
