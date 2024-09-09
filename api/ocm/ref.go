package ocm

import (
	"fmt"
	"strings"

	"github.com/mandelsoft/goutils/errors"

	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/ocm/grammar"
	common "ocm.software/ocm/api/utils/misc"
)

// ParseRepo parses a standard ocm repository reference into a internal representation.
func ParseRepo(ref string) (UniformRepositorySpec, error) {
	return cpi.ParseRepo(ref)
}

func ParseRepoToSpec(ctx Context, ref string, create ...bool) (RepositorySpec, error) {
	return cpi.ParseRepoToSpec(ctx, ref, create...)
}

// RefSpec is a go internal representation of a oci reference.
type RefSpec struct {
	UniformRepositorySpec
	CompSpec
}

// ParseRef parses a standard ocm reference into an internal representation.
func ParseRef(ref string) (RefSpec, error) {
	create := false
	if strings.HasPrefix(ref, "+") {
		create = true
		ref = ref[1:]
	}

	var spec RefSpec
	v := ""
	match := grammar.AnchoredReferenceRegexp.FindSubmatch([]byte(ref))
	if match != nil {
		v = string(match[6])
		s := string(match[2])
		h := string(match[1])
		t, _ := grammar.SplitTypeSpec(h)
		spec = RefSpec{
			UniformRepositorySpec{
				Type:            t,
				TypeHint:        h,
				Scheme:          s,
				Host:            string(match[3]),
				SubPath:         string(match[4]),
				CreateIfMissing: create,
			},
			CompSpec{
				Component: string(match[5]),
				Version:   nil,
			},
		}
	}

	if match == nil {
		match = grammar.AnchoredSchemedHostPortReferenceRegexp.FindSubmatch([]byte(ref))
		if match != nil {
			v = string(match[6])
			s := string(match[2])
			h := string(match[1])
			t, _ := grammar.SplitTypeSpec(h)
			spec = RefSpec{
				UniformRepositorySpec{
					Type:            t,
					TypeHint:        h,
					Scheme:          s,
					Host:            string(match[3]),
					SubPath:         string(match[4]),
					CreateIfMissing: create,
				},
				CompSpec{
					Component: string(match[5]),
					Version:   nil,
				},
			}
		}
	}

	if match == nil {
		match = grammar.AnchoredHostWithPortReferenceRegexp.FindSubmatch([]byte(ref))
		if match != nil {
			v = string(match[6])
			s := string(match[2])
			h := string(match[1])
			t, _ := grammar.SplitTypeSpec(h)
			spec = RefSpec{
				UniformRepositorySpec{
					Type:            t,
					TypeHint:        h,
					Scheme:          s,
					Host:            string(match[3]),
					SubPath:         string(match[4]),
					CreateIfMissing: create,
				},
				CompSpec{
					Component: string(match[5]),
					Version:   nil,
				},
			}
		}
	}

	if match == nil {
		match = grammar.AnchoredGenericReferenceRegexp.FindSubmatch([]byte(ref))
		if match == nil {
			return RefSpec{}, errors.ErrInvalid(KIND_OCM_REFERENCE, ref)
		}
		v = string(match[4])
		h := string(match[1])
		t, _ := grammar.SplitTypeSpec(h)
		spec = RefSpec{
			UniformRepositorySpec{
				Type:            t,
				TypeHint:        h,
				Info:            string(match[2]),
				CreateIfMissing: create,
			},
			CompSpec{
				Component: string(match[3]),
				Version:   nil,
			},
		}
	}

	if v != "" {
		spec.Version = &v
	}
	var err error
	if spec.Info == "" || !(string(spec.Info[0]) == "{") {
		spec.UniformRepositorySpec, err = cpi.HandleRef(spec.UniformRepositorySpec)
	}
	return spec, err
}

func (r *RefSpec) Name() string {
	if r.SubPath == "" {
		return fmt.Sprintf("%s//%s", r.Host, r.Component)
	}
	return fmt.Sprintf("%s/%s//%s", r.Host, r.SubPath, r.Component)
}

func (r *RefSpec) HostPort() (string, string) {
	i := strings.Index(r.Host, ":")
	if i < 0 {
		return r.Host, ""
	}
	return r.Host[:i], r.Host[i+1:]
}

func (r *RefSpec) Reference() string {
	t := r.Type
	if t != "" {
		t += "::"
	}
	s := r.SubPath
	if s != "" {
		s = "/" + s
	}
	v := ""
	if r.Version != nil && *r.Version != "" {
		v = ":" + *r.Version
	}
	return fmt.Sprintf("%s%s%s//%s%s", t, r.Host, s, r.Component, v)
}

func (r *RefSpec) IsVersion() bool {
	return r.Version != nil
}

func (r *RefSpec) String() string {
	return r.Reference()
}

func (r RefSpec) DeepCopy() RefSpec {
	if r.Version != nil {
		v := *r.Version
		r.Version = &v
	}
	return r
}

////////////////////////////////////////////////////////////////////////////////

func ParseComp(ref string) (CompSpec, error) {
	match := grammar.AnchoredComponentVersionRegexp.FindSubmatch([]byte(ref))

	if match == nil {
		return CompSpec{}, errors.ErrInvalid(KIND_COMPONENTVERSION, ref)
	}

	v := string(match[2])
	r := CompSpec{
		Component: string(match[1]),
		Version:   nil,
	}
	if v != "" {
		r.Version = &v
	}
	return r, nil
}

// CompSpec is a go internal representation of a ocm component version name.
type CompSpec struct {
	// Component is the component name part of a component version
	Component string
	// +optional
	Version *string
}

func (r *CompSpec) IsVersion() bool {
	return r.Version != nil
}

func (r *CompSpec) NameVersion() common.NameVersion {
	if r.Version != nil {
		return common.NewNameVersion(r.Component, *r.Version)
	}
	return common.NewNameVersion(r.Component, "-")
}

func (r *CompSpec) Reference() string {
	v := ""
	if r.Version != nil && *r.Version != "" {
		v = ":" + *r.Version
	}
	return fmt.Sprintf("%s%s", r.Component, v)
}

func (r *CompSpec) String() string {
	return r.Reference()
}
