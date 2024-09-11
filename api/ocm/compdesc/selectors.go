package compdesc

import (
	"encoding/json"
	"reflect"
	"runtime"

	v1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	resourcetypes "ocm.software/ocm/api/ocm/extensions/artifacttypes"
	"ocm.software/ocm/api/ocm/extraid"
	"ocm.software/ocm/api/ocm/selectors/accessors"
	"ocm.software/ocm/api/utils"
	"ocm.software/ocm/api/utils/selector"
)

// Deprecated: use package selectors and its sub packages.
type IdentityBasedSelector interface {
	IdentitySelector
	ElementSelector
	ResourceSelector
	ReferenceSelector
}

// Deprecated: use package selectors and its sub packages.
type ElementBasedSelector interface {
	ElementSelector
	ResourceSelector
	ReferenceSelector
}

// Deprecated: use package selectors and its sub packages.
type LabelBasedSelector interface {
	LabelSelector
	ElementSelector
	ResourceSelector
	ReferenceSelector
}

////////////////////////////////////////////////////////////////////////////////

// Deprecated: use package selectors and its sub packages.
type IdentitySelector = selector.Interface

type byVersion struct {
	version string
}

var _ IdentityBasedSelector = (*byVersion)(nil)

func (b *byVersion) Match(obj map[string]string) (bool, error) {
	return obj[SystemIdentityVersion] == b.version, nil
}

func (b *byVersion) MatchElement(obj ElementSelectionContext) (bool, error) {
	return obj.GetVersion() == b.version, nil
}

func (b *byVersion) MatchResource(obj ResourceSelectionContext) (bool, error) {
	return obj.GetVersion() == b.version, nil
}

func (b *byVersion) MatchReference(obj ReferenceSelectionContext) (bool, error) {
	return obj.GetVersion() == b.version, nil
}

// ByVersion creates a new selector that
// selects an element based on its version.
//
// Deprecated: use package selectors and its sub packages.
func ByVersion(version string) IdentityBasedSelector {
	return &byVersion{version: version}
}

type byName struct {
	name string
}

var _ IdentityBasedSelector = (*byName)(nil)

func (b *byName) Match(obj map[string]string) (bool, error) {
	return obj[SystemIdentityName] == b.name, nil
}

func (b *byName) MatchElement(obj ElementSelectionContext) (bool, error) {
	return obj.GetName() == b.name, nil
}

func (b *byName) MatchResource(obj ResourceSelectionContext) (bool, error) {
	return obj.GetName() == b.name, nil
}

func (b *byName) MatchReference(obj ReferenceSelectionContext) (bool, error) {
	return obj.GetName() == b.name, nil
}

// ByName creates a new selector that
// selects an element based on its name.
//
// Deprecated: use package selectors and its sub packages.
func ByName(name string) IdentityBasedSelector {
	return &byName{name: name}
}

type byIdentity struct {
	id      v1.Identity
	partial bool
}

var _ IdentityBasedSelector = (*byIdentity)(nil)

func (b *byIdentity) Match(obj map[string]string) (bool, error) {
	if !b.partial && len(b.id) != len(obj) {
		return false, nil
	}
	for k, v := range b.id {
		e, ok := obj[k]
		if !ok || e != v {
			return false, nil
		}
	}
	return true, nil
}

func (b *byIdentity) MatchElement(obj ElementSelectionContext) (bool, error) {
	return b.Match(obj.Identity())
}

func (b *byIdentity) MatchResource(obj ResourceSelectionContext) (bool, error) {
	return b.Match(obj.Identity())
}

func (b *byIdentity) MatchReference(obj ReferenceSelectionContext) (bool, error) {
	return b.Match(obj.Identity())
}

// ByIdentity creates a new resource and identity selector that
// selects a resource based on its identity.
//
// Deprecated: use package selectors and its sub packages.
func ByIdentity(name string, extras ...string) IdentityBasedSelector {
	id := v1.NewIdentity(name, extras...)
	return &byIdentity{id: id}
}

// ByPartialIdentity creates a new resource and identity selector that
// selects a resource based on its partial identity.
// All given attributes must match, but potential additional attributes
// of a resource identity are ignored.
//
// Deprecated: use package selectors and its sub packages.
func ByPartialIdentity(name string, extras ...string) IdentityBasedSelector {
	id := v1.NewIdentity(name, extras...)
	return &byIdentity{id: id, partial: true}
}

type withExtraId struct {
	ids v1.Identity
}

var _ ElementBasedSelector = (*withExtraId)(nil)

func (b *withExtraId) Match(obj map[string]string) (bool, error) {
	if len(obj) == 0 {
		return len(b.ids) == 0, nil
	}
	for id, v := range b.ids {
		if obj[id] != v {
			return false, nil
		}
	}
	return true, nil
}

func (b *withExtraId) MatchElement(obj ElementSelectionContext) (bool, error) {
	return b.Match(obj.GetExtraIdentity())
}

func (b *withExtraId) MatchResource(obj ResourceSelectionContext) (bool, error) {
	return b.Match(obj.ExtraIdentity)
}

func (b *withExtraId) MatchReference(obj ReferenceSelectionContext) (bool, error) {
	return b.Match(obj.ExtraIdentity)
}

// WithExtraIdentity creates a new selector that
// selects an element based on extra identities.
//
// Deprecated: use package selectors and its sub packages.
func WithExtraIdentity(args ...string) IdentityBasedSelector {
	ids := v1.Identity{}
	for i := 0; i < len(args); i += 2 {
		if i+1 < len(args) {
			ids[args[i]] = args[i+1]
		}
	}
	return &withExtraId{ids: ids}
}

////////////////////////////////////////////////////////////////////////////////

// ResourceSelectorFunc defines a function to filter a resource.
//
// Deprecated: use package selectors and its sub packages.
type ResourceSelectorFunc func(obj ResourceSelectionContext) (bool, error)

var _ ResourceSelector = ResourceSelectorFunc(nil)

func (s ResourceSelectorFunc) MatchResource(obj ResourceSelectionContext) (bool, error) {
	return s(obj)
}

type resourceSelectionContext struct {
	*Resource
	identity
}

// Deprecated: use package selectors and its sub packages.
func NewResourceSelectionContext(index int, rscs Resources) ResourceSelectionContext {
	return &resourceSelectionContext{
		Resource: &rscs[index],
		identity: identity{
			accessor: rscs,
			index:    index,
		},
	}
}

// ResourceSelectionContext describes the selction context for a resource
// selector. It contains the resource and provides access to its
// identity in the context of its component descriptor.
//
// Deprecated: use package selectors and its sub packages.
type ResourceSelectionContext = *resourceSelectionContext

// ResourceSelector defines a selector based on resource attributes.
//
// Deprecated: use package selectors and its sub packages.
type ResourceSelector interface {
	MatchResource(obj ResourceSelectionContext) (bool, error)
}

// MatchResourceByResourceSelector applies all resource selector against the given resource object.
//
// Deprecated: use package selectors and its sub packages.
func MatchResourceByResourceSelector(obj ResourceSelectionContext, resourceSelectors ...ResourceSelector) (bool, error) {
	for _, sel := range resourceSelectors {
		ok, err := sel.MatchResource(obj)
		if err != nil {
			return false, err
		}
		if !ok {
			return false, nil
		}
	}
	return true, nil
}

// AndR is an AND resource selector.
//
// Deprecated: use package labelsel.
func AndR(sel ...ResourceSelector) ResourceSelector {
	return ResourceSelectorFunc(func(obj ResourceSelectionContext) (bool, error) {
		for _, s := range sel {
			ok, err := s.MatchResource(obj)
			if !ok || err != nil {
				return ok, err
			}
		}
		return true, nil
	})
}

// OrR is an OR resource selector.
//
// Deprecated: use package labelsel.
func OrR(sel ...ResourceSelector) ResourceSelector {
	return ResourceSelectorFunc(func(obj ResourceSelectionContext) (bool, error) {
		for _, s := range sel {
			ok, err := s.MatchResource(obj)
			if ok || err != nil {
				return ok, err
			}
		}
		return false, nil
	})
}

// NotR is a negated resource selector.
//
// Deprecated: use package labelsel.
func NotR(sel ResourceSelector) ResourceSelector {
	return ResourceSelectorFunc(func(obj ResourceSelectionContext) (bool, error) {
		ok, err := sel.MatchResource(obj)
		if err != nil {
			return false, err
		}
		return !ok, nil
	})
}

// ByResourceType creates a new resource selector that
// selects a resource based on its type.
//
// Deprecated: use package selectors and its sub packages.
func ByResourceType(ttype string) ResourceSelector {
	return ResourceSelectorFunc(func(obj ResourceSelectionContext) (bool, error) {
		return ttype == "" || obj.GetType() == ttype, nil
	})
}

// ByRelation creates a new resource selector that
// selects a resource based on its relation type.
//
// Deprecated: use package selectors and its sub packages.
func ByRelation(relation v1.ResourceRelation) ResourceSelectorFunc {
	return ResourceSelectorFunc(func(obj ResourceSelectionContext) (bool, error) {
		return obj.Relation == relation, nil
	})
}

// ByAccessMethod creates a new selector that matches a resource access method type.
//
// Deprecated: use package selectors and its sub packages.
func ByAccessMethod(name string) ResourceSelector {
	return ResourceSelectorFunc(func(obj ResourceSelectionContext) (bool, error) {
		if obj.Access == nil {
			return name == "", nil
		}
		return obj.Access.GetType() == name || obj.Access.GetKind() == name, nil
	})
}

// ForExecutable creates a new selector that matches a resource for an executable.
//
// Deprecated: use package selectors and its sub packages.
func ForExecutable(name string) ResourceSelector {
	return ResourceSelectorFunc(func(obj ResourceSelectionContext) (bool, error) {
		return obj.Name == name && obj.Type == resourcetypes.EXECUTABLE && obj.ExtraIdentity != nil &&
			obj.ExtraIdentity[extraid.ExecutableOperatingSystem] == runtime.GOOS &&
			obj.ExtraIdentity[extraid.ExecutableArchitecture] == runtime.GOARCH, nil
	})
}

////////////////////////////////////////////////////////////////////////////////

// LabelSelector is used to match a label in a label set.
//
// Deprecated: use package selectors and its sub packages.
type LabelSelector interface {
	MatchLabel(l v1.Label) (bool, error)
}

// LabelSelectorFunc is a function used as LabelSelector.
//
// Deprecated: use package selectors and its sub packages.
type LabelSelectorFunc func(l v1.Label) (bool, error)

func (l LabelSelectorFunc) MatchLabel(label v1.Label) (bool, error) {
	return l(label)
}

// AndL is an AND label selector.
//
// Deprecated: use package selectors and its sub packages.
func AndL(sel ...LabelSelector) LabelSelector {
	return LabelSelectorFunc(func(obj v1.Label) (bool, error) {
		for _, s := range sel {
			ok, err := s.MatchLabel(obj)
			if !ok || err != nil {
				return ok, err
			}
		}
		return true, nil
	})
}

// OrL is an OR label selector.
//
// Deprecated: use package selectors and its sub packages.
func OrL(sel ...LabelSelector) LabelSelector {
	return LabelSelectorFunc(func(obj v1.Label) (bool, error) {
		for _, s := range sel {
			ok, err := s.MatchLabel(obj)
			if ok || err != nil {
				return ok, err
			}
		}
		return false, nil
	})
}

// NotL is a negated label selector.
//
// Deprecated: use package selectors and its sub packages.
func NotL(sel LabelSelector) LabelSelector {
	return LabelSelectorFunc(func(obj v1.Label) (bool, error) {
		ok, err := sel.MatchLabel(obj)
		if err != nil {
			return false, err
		}
		return !ok, nil
	})
}

type byLabel struct {
	selector LabelSelector
}

var _ LabelBasedSelector = (*byLabel)(nil)

func (b *byLabel) MatchElement(obj ElementSelectionContext) (bool, error) {
	return b.MatchLabels(obj.GetLabels())
}

func (b *byLabel) MatchResource(obj ResourceSelectionContext) (bool, error) {
	return b.MatchLabels(obj.Labels)
}

func (b *byLabel) MatchReference(obj ReferenceSelectionContext) (bool, error) {
	return b.MatchLabels(obj.Labels)
}

func (b *byLabel) MatchLabels(obj v1.Labels) (bool, error) {
	for _, l := range obj {
		if ok, err := b.selector.MatchLabel(l); ok || err != nil {
			return true, nil
		}
	}
	return false, nil
}

func (b *byLabel) MatchLabel(l v1.Label) (bool, error) {
	return b.selector.MatchLabel(l)
}

// ByLabel matches a resource or element for a list of given label selectors
// matching the same label.
// If multiple label related selectors should be used, they should
// be grouped into a single label selector to be applied in
// combination. Otherwise, a resource might match if the label
// selectors all match, but different labels.
//
// Deprecated: use package selectors and its sub packages.
func ByLabel(sel ...LabelSelector) LabelBasedSelector {
	return &byLabel{selector: LabelSelectorFunc(func(l v1.Label) (bool, error) {
		return MatchLabels(v1.Labels{l}, sel...)
	})}
}

// ByLabelName matches an element by a label name.
//
// Deprecated: use package selectors and its sub packages.
func ByLabelName(name string) LabelBasedSelector {
	return &byLabel{selector: LabelSelectorFunc(func(l v1.Label) (bool, error) { return l.Name == name, nil })}
}

// ByLabelValue matches a resource or label by a label value.
// This selector should typically be combined with ByLabelName.
//
// Deprecated: use package selectors and its sub packages.
func ByLabelValue(value interface{}) LabelBasedSelector {
	return &byLabel{selector: LabelSelectorFunc(func(l v1.Label) (bool, error) {
		var data interface{}
		if err := json.Unmarshal(l.Value, &data); err != nil {
			return false, err
		}
		return reflect.DeepEqual(data, value), nil
	})}
}

// ByLabelVersion matches a resource or label by a label version.
// This selector should typically be combined with ByLabelName.
//
// Deprecated: use package selectors and its sub packages.
func ByLabelVersion(version string) LabelBasedSelector {
	return &byLabel{selector: LabelSelectorFunc(func(l v1.Label) (bool, error) { return l.Version == version, nil })}
}

// BySignedLabel matches a resource or label by a label indicated to be signed.
// This selector should typically be combined with ByLabelName.
//
// Deprecated: use package selectors and its sub packages.
func BySignedLabel(flags ...bool) LabelBasedSelector {
	flag := utils.OptionalDefaultedBool(true, flags...)
	return &byLabel{selector: LabelSelectorFunc(func(l v1.Label) (bool, error) { return l.Signing == flag, nil })}
}

// MatchLabels checks whether a set of labels matches the given label selectors.
//
// Deprecated: use package selectors and its sub packages.
func MatchLabels(labels v1.Labels, sel ...LabelSelector) (bool, error) {
	if len(labels) == 0 && len(sel) == 0 {
		return true, nil
	}
	found := false
outer:
	for _, l := range labels {
		for _, s := range sel {
			ok, err := s.MatchLabel(l)
			if err != nil {
				return false, err
			}
			if !ok {
				continue outer
			}
		}
		found = true
		break
	}

	return found, nil
}

// SelectLabels returns labels matching the given label selectors.
//
// Deprecated: use package selectors and its sub packages.
func SelectLabels(labels v1.Labels, sel ...LabelSelector) (v1.Labels, error) {
	list := make(v1.Labels, 0)
outer:
	for _, l := range labels {
		for _, s := range sel {
			ok, err := s.MatchLabel(l)
			if err != nil {
				return nil, err
			}
			if !ok {
				continue outer
			}
		}
		list = append(list, l)
	}

	return list, nil
}

// MatchReferencesByReferenceSelector applies all resource selector against the given resource object.
//
// Deprecated: use package selectors and its sub packages.
func MatchReferencesByReferenceSelector(obj ReferenceSelectionContext, resourceSelectors ...ReferenceSelector) (bool, error) {
	for _, sel := range resourceSelectors {
		ok, err := sel.MatchReference(obj)
		if err != nil {
			return false, err
		}
		if !ok {
			return false, nil
		}
	}
	return true, nil
}

////////////////////////////////////////////////////////////////////////////////

type elementSelectionContext struct {
	accessors.ElementMeta
	identity
}

// Deprecated: not supported anymore.
type ElementSelectionContext = *elementSelectionContext

// Deprecated: not supported anymore.
func NewElementSelectionContext(index int, elems ElementListAccessor) ElementSelectionContext {
	return &elementSelectionContext{
		ElementMeta: elems.Get(index).GetMeta(),
		identity: identity{
			accessor: elems,
			index:    index,
		},
	}
}

// Deprecated: use package selectors and its sub packages.
type ElementSelector interface {
	MatchElement(obj ElementSelectionContext) (bool, error)
}

////////////////////////////////////////////////////////////////////////////////

// ReferenceSelectorFunc defines a function to filter a resource.
//
// Deprecated: use package selectors and its sub packages.
type ReferenceSelectorFunc func(obj ReferenceSelectionContext) (bool, error)

var _ ReferenceSelector = ReferenceSelectorFunc(nil)

func (s ReferenceSelectorFunc) MatchReference(obj ReferenceSelectionContext) (bool, error) {
	return s(obj)
}

type referenceSelectionContext struct {
	*Reference
	identity
}

// Deprecated: use package selectors and its sub packages.
func NewReferenceSelectionContext(index int, refs References) ReferenceSelectionContext {
	return &referenceSelectionContext{
		Reference: &refs[index],
		identity: identity{
			accessor: refs,
			index:    index,
		},
	}
}

// ReferenceSelectionContext describes the selection context for a reference
// selector. It contains the reference and provides access to its
// identity in the context of its component descriptor.
//
// Deprecated: use package selectors and its sub packages.
type ReferenceSelectionContext = *referenceSelectionContext

// ReferenceSelector defines a selector based on reference attributes.
//
// Deprecated: use package selectors and its sub packages.
type ReferenceSelector interface {
	MatchReference(obj ReferenceSelectionContext) (bool, error)
}

// AndC is an AND reference selector.
//
// Deprecated: use package selectors and its sub packages.
func AndC(sel ...ReferenceSelector) ReferenceSelector {
	return ReferenceSelectorFunc(func(obj ReferenceSelectionContext) (bool, error) {
		for _, s := range sel {
			ok, err := s.MatchReference(obj)
			if !ok || err != nil {
				return ok, err
			}
		}
		return true, nil
	})
}

// OrC is an OR resource selector.
//
// Deprecated: use package selectors and its sub packages.
func OrC(sel ...ReferenceSelector) ReferenceSelector {
	return ReferenceSelectorFunc(func(obj ReferenceSelectionContext) (bool, error) {
		for _, s := range sel {
			ok, err := s.MatchReference(obj)
			if ok || err != nil {
				return ok, err
			}
		}
		return false, nil
	})
}

// NotC is a negated resource selector.
//
// Deprecated: use package selectors and its sub packages.
func NotC(sel ReferenceSelector) ReferenceSelector {
	return ReferenceSelectorFunc(func(obj ReferenceSelectionContext) (bool, error) {
		ok, err := sel.MatchReference(obj)
		if err != nil {
			return false, err
		}
		return !ok, nil
	})
}

// Deprecated: use package selectors and its sub packages.
func ByComponent(name string) ReferenceSelector {
	return ReferenceSelectorFunc(func(obj ReferenceSelectionContext) (bool, error) {
		return obj.ComponentName == name, nil
	})
}

////////////////////////////////////////////////////////////////////////////////

type identity struct {
	identity v1.Identity
	accessor ElementListAccessor
	index    int
}

func (i *identity) Identity() v1.Identity {
	if i.identity == nil {
		i.identity = i.accessor.Get(i.index).GetMeta().GetIdentity(i.accessor)
	}
	return i.identity
}
