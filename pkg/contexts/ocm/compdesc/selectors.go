// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package compdesc

import (
	"encoding/json"
	"reflect"
	"runtime"

	v1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/consts"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/resourcetypes"
	"github.com/open-component-model/ocm/pkg/utils"
	"github.com/open-component-model/ocm/pkg/utils/selector"
)

type IdentitySelector = selector.Interface

// ResourceSelectorFunc defines a function to filter a resource.
type ResourceSelectorFunc func(obj Resource) (bool, error)

var _ ResourceSelector = ResourceSelectorFunc(nil)

func (s ResourceSelectorFunc) MatchResource(obj Resource) (bool, error) {
	return s(obj)
}

// ResourceSelector defines a selector bases on resource attributes.
type ResourceSelector interface {
	MatchResource(obj Resource) (bool, error)
}

// IdentityAndResourceSelector is selector, which can act as
// resource and/or identity selector.
type IdentityAndResourceSelector interface {
	IdentitySelector
	ResourceSelector
}

// MatchResourceByResourceSelector applies all resource selector against the given resource object.
func MatchResourceByResourceSelector(obj Resource, resourceSelectors ...ResourceSelector) (bool, error) {
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
func AndR(sel ...ResourceSelector) ResourceSelector {
	return ResourceSelectorFunc(func(obj Resource) (bool, error) {
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
func OrR(sel ...ResourceSelector) ResourceSelector {
	return ResourceSelectorFunc(func(obj Resource) (bool, error) {
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
func NotR(sel ResourceSelector) ResourceSelector {
	return ResourceSelectorFunc(func(obj Resource) (bool, error) {
		ok, err := sel.MatchResource(obj)
		if err != nil {
			return false, err
		}
		return !ok, nil
	})
}

// ByResourceType creates a new resource selector that
// selects a resource based on its type.
func ByResourceType(ttype string) ResourceSelector {
	return ResourceSelectorFunc(func(obj Resource) (bool, error) {
		return ttype == "" || obj.GetType() == ttype, nil
	})
}

// ByRelation creates a new resource selector that
// selects a resource based on its relation type.
func ByRelation(relation v1.ResourceRelation) ResourceSelectorFunc {
	return ResourceSelectorFunc(func(obj Resource) (bool, error) {
		return obj.Relation == relation, nil
	})
}

type byVersion struct {
	version string
}

var (
	_ ResourceSelector = (*byVersion)(nil)
	_ IdentitySelector = (*byVersion)(nil)
)

func (b *byVersion) Match(obj map[string]string) (bool, error) {
	return obj[SystemIdentityVersion] == b.version, nil
}

func (b *byVersion) MatchResource(obj Resource) (bool, error) {
	return obj.GetVersion() == b.version, nil
}

// ByVersion creates a new resource and identity selector that
// selects a resource based on its version.
func ByVersion(version string) IdentityAndResourceSelector {
	return &byVersion{version: version}
}

type byName struct {
	name string
}

var (
	_ ResourceSelector = (*byName)(nil)
	_ IdentitySelector = (*byName)(nil)
)

func (b *byName) Match(obj map[string]string) (bool, error) {
	return obj[SystemIdentityName] == b.name, nil
}

func (b *byName) MatchResource(obj Resource) (bool, error) {
	return obj.GetName() == b.name, nil
}

// ByName creates a new selector that matches a resource name.
func ByName(name string) IdentityAndResourceSelector {
	return &byName{name: name}
}

type withExtraId struct {
	ids v1.Identity
}

var (
	_ ResourceSelector = (*withExtraId)(nil)
	_ IdentitySelector = (*withExtraId)(nil)
)

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

func (b *withExtraId) MatchResource(obj Resource) (bool, error) {
	return b.Match(obj.ExtraIdentity)
}

// WithExtraIdentity creates a new resource and identity selector that
// selects a resource based on extra identities.
func WithExtraIdentity(args ...string) IdentityAndResourceSelector {
	ids := v1.Identity{}
	for i := 0; i < len(args); i += 2 {
		if i+1 < len(args) {
			ids[args[i]] = args[i+1]
		}
	}
	return &withExtraId{ids: ids}
}

// ByAccessMethod creates a new selector that matches a resource access method type.
func ByAccessMethod(name string) ResourceSelector {
	return ResourceSelectorFunc(func(obj Resource) (bool, error) {
		if obj.Access == nil {
			return name == "", nil
		}
		return obj.Access.GetType() == name || obj.Access.GetKind() == name, nil
	})
}

// ForExecutable creates a new selector that matches a resource for an executable.
func ForExecutable(name string) ResourceSelector {
	return ResourceSelectorFunc(func(obj Resource) (bool, error) {
		return obj.Name == name && obj.Type == resourcetypes.EXECUTABLE && obj.ExtraIdentity != nil &&
			obj.ExtraIdentity[consts.ExecutableOperatingSystem] == runtime.GOOS &&
			obj.ExtraIdentity[consts.ExecutableArchitecture] == runtime.GOARCH, nil
	})
}

////////////////////////////////////////////////////////////////////////////////

// LabelSelector is used to match a label in a label set.
type LabelSelector interface {
	MatchLabel(l v1.Label) (bool, error)
}

type ResourceAndLabelSelector interface {
	ResourceSelector
	LabelSelector
}

// LabelSelectorFunc is a function used as LabelSelector.
type LabelSelectorFunc func(l v1.Label) (bool, error)

func (l LabelSelectorFunc) MatchLabel(label v1.Label) (bool, error) {
	return l(label)
}

// AndL is an AND label selector.
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

var (
	_ ResourceSelector = (*byLabel)(nil)
	_ LabelSelector    = (*byLabel)(nil)
)

func (b *byLabel) MatchResource(obj Resource) (bool, error) {
	for _, l := range obj.Labels {
		if ok, err := b.selector.MatchLabel(l); ok || err != nil {
			return true, nil
		}
	}
	return false, nil
}

func (b *byLabel) MatchLabel(l v1.Label) (bool, error) {
	return b.selector.MatchLabel(l)
}

// ByLabel matches a resource for a list of given label selectors
// matching the same label.
// If multiple label related selectors should be used, they should
// be grouped into a single label selector to be applied in
// combination. Otherwise, a resource might match if the label
// selectors all match, but different labels.
func ByLabel(sel ...LabelSelector) ResourceSelector {
	return ResourceSelectorFunc(func(obj Resource) (bool, error) {
		return MatchLabels(obj.Labels, sel...)
	})
}

// ByLabelName matches a resource or label by a label name.
func ByLabelName(name string) ResourceAndLabelSelector {
	return &byLabel{selector: LabelSelectorFunc(func(l v1.Label) (bool, error) { return l.Name == name, nil })}
}

// ByLabelValue matches a resource or label by a label value.
// This selector should typically be combined with ByLabelName.
func ByLabelValue(value interface{}) ResourceAndLabelSelector {
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
func ByLabelVersion(version string) ResourceAndLabelSelector {
	return &byLabel{selector: LabelSelectorFunc(func(l v1.Label) (bool, error) { return l.Version == version, nil })}
}

// BySignedLabel matches a resource or label by a label indicated to be signed.
// This selector should typically be combined with ByLabelName.
func BySignedLabel(flags ...bool) ResourceAndLabelSelector {
	flag := utils.OptionalDefaultedBool(true, flags...)
	return &byLabel{selector: LabelSelectorFunc(func(l v1.Label) (bool, error) { return l.Signing == flag, nil })}
}

// MatchLabels checks whether a set of labels matches the given label selectors.
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
