package localize

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mandelsoft/goutils/errors"

	"ocm.software/ocm/api/ocm"
	v1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	utils "ocm.software/ocm/api/ocm/ocmutils"
	"ocm.software/ocm/api/ocm/resourcerefs"
	"ocm.software/ocm/api/utils/runtime"
)

// Definition of inbound substitution requests.
// - Localizations  image locations substitution resolved using a component version
// - Configurations configuration substitution resolved by provided value data.
//
// Such requests can be given to merge externally provided data into
// some filesystem template.
// The evaluation of such requests results in a list
// of resolved substitution requests that can be applied without
// further value context to a filesystem structure.

// ImageMapping describes a dedicated substitution of parts
// of container image names based on a relative OCM resource reference.
type ImageMapping struct {
	// The optional but unique(!) name of the mapping to support referencing mapping entries
	Name string `json:"name,omitempty"`

	// The resource reference used to resolve the substitution
	v1.ResourceReference `json:",inline"`

	// The optional variants for the value determination

	// Path in target to substitute by the image tag/digest
	Tag string `json:"tag,omitempty"`
	// Path in target to substitute the image repository
	Repository string `json:"repository,omitempty"`
	// Path in target to substitute the complete image
	Image string `json:"image,omitempty"`
}

type ImageMappings []ImageMapping

func (m *ImageMapping) Evaluate(idx int, cv ocm.ComponentVersionAccess, resolver ocm.ComponentVersionResolver) (ValueMappings, error) {
	name := "image mapping"
	if m.Name != "" {
		name = fmt.Sprintf("%s %q", name, m.Name)
	} else {
		if idx >= 0 {
			name = fmt.Sprintf("%s %d", name, idx+1)
		}
	}
	acc, rcv, err := resourcerefs.ResolveResourceReference(cv, m.ResourceReference, resolver)
	if err != nil {
		return nil, errors.Wrapf(err, "mapping", fmt.Sprintf("%s (%s)", name, &m.ResourceReference))
	}
	defer rcv.Close()
	ref, err := utils.GetOCIArtifactRef(cv.GetContext(), acc)
	if err != nil {
		return nil, errors.Wrapf(err, "mapping %s: cannot resolve resource %s to an OCI Reference", name, &m.ResourceReference)
	}
	ix := strings.Index(ref, ":")
	if ix < 0 {
		ix = strings.Index(ref, "@")
		if ix < 0 {
			return nil, errors.Wrapf(err, "mapping %s: image tag or digest missing (%s)", name, ref)
		}
	}
	repo := ref[:ix]
	tag := ref[ix+1:]

	cnt := 0
	if m.Repository != "" {
		cnt++
	}
	if m.Tag != "" {
		cnt++
	}
	if m.Image != "" {
		cnt++
	}
	if cnt == 0 {
		return nil, fmt.Errorf("no substitution target given for %s", name)
	}

	var result ValueMappings
	var r *ValueMapping
	if m.Repository != "" {
		if r, err = NewValueMapping(substitutionName(name, "repository", cnt), m.Repository, repo); err != nil {
			return nil, errors.Wrapf(err, "setting repository for %s", substitutionName(name, "repository", cnt))
		}
		result = append(result, *r)
	}
	if m.Tag != "" {
		if r, err = NewValueMapping(substitutionName(name, "tag", cnt), m.Tag, tag); err != nil {
			return nil, errors.Wrapf(err, "setting tag for %s", substitutionName(name, "tag", cnt))
		}
		result = append(result, *r)
	}
	if m.Image != "" {
		if r, err = NewValueMapping(substitutionName(name, "image", cnt), m.Image, ref); err != nil {
			return nil, errors.Wrapf(err, "setting image for %s", substitutionName(name, "image", cnt))
		}
		result = append(result, *r)
	}
	return result, nil
}

// Localization is a request to substitute an image location.
// The specification describes substitution targets given by the file path and
// the YAML/JSON value paths of the elements in this file.
// The substitution value is calculated
// from the access specification of the given resource provided by the actual
// component version.
type Localization struct {
	// The path of the file for the substitution
	FilePath string `json:"file"`
	// The image mapping request
	ImageMapping `json:",inline"`
}

// Configuration is a request to substitute a configuration value.
// The specification describes substitution targets given by the file path and
// the YAML/JSON value paths of the elements in this file.
// The substitution value is calculated
// by the value expression (spiff) based on given config data.
// It has the same structure as Substitution, but is a request based
// on external configuration data, while a Substitution describes a fixed target
// value.
type Configuration Substitution

type ValueMapping struct {
	// The optional but unique(!) name of the mapping to support referencing mapping entries
	Name string `json:"name,omitempty"`
	// The target path for the value substitution
	ValuePath string `json:"path"`
	// The value to set
	Value json.RawMessage `json:"value"`
}

func NewValueMapping(name, path string, value interface{}) (*ValueMapping, error) {
	var (
		v   []byte
		err error
	)

	if value != nil {
		v, err = runtime.DefaultJSONEncoding.Marshal(value)
		if err != nil {
			return nil, fmt.Errorf("cannot marshal substitution value: %w", err)
		}
	}
	return &ValueMapping{
		Name:      name,
		ValuePath: path,
		Value:     v,
	}, nil
}

type ValueMappings []ValueMapping

func (s *ValueMappings) Add(name, path string, value interface{}) error {
	m, err := NewValueMapping(name, path, value)
	if err != nil {
		return err
	}
	*s = append(*s, *m)
	return nil
}

// Here comes the structure used for resolved execution requests.
// They can be applied to a filesystem content without further external information.
// It basically has the same structure as the configuration request, but
// the given value is just the target value without any further interpretation.
// This way configuration requests could just provide dedicated values, also

// Substitution is a request to substitute the YAML/JSON
// element given by the value path in the given file path by the given
// direct value.
type Substitution struct {
	// The path of the file for the substitution
	FilePath string `json:"file"`
	// The field mapping toapply to given file path
	ValueMapping `json:",inline"`
}

func (s *Substitution) GetValue() (interface{}, error) {
	var value interface{}
	err := runtime.DefaultYAMLEncoding.Unmarshal(s.Value, &value)
	return value, err
}

type Substitutions []Substitution

func (s *Substitutions) AddValueMapping(m *ValueMapping, file string) {
	*s = append(*s, Substitution{
		FilePath:     file,
		ValueMapping: *m,
	})
}

func (s *Substitutions) Add(name, file, path string, value interface{}) error {
	m, err := NewValueMapping(name, path, value)
	if err != nil {
		return err
	}
	s.AddValueMapping(m, file)
	return nil
}

// InstantiationRules bundle the localization of a filesystem resource
// covering image localization and applying instance configuration.
type InstantiationRules struct {
	Template          v1.ResourceReference   `json:"templateResource,omitempty"`
	LocalizationRules []Localization         `json:"localizationRules,omitempty"`
	ConfigRules       []Configuration        `json:"configRules,omitempty"`
	ConfigScheme      json.RawMessage        `json:"configScheme,omitempty"`
	ConfigTemplate    json.RawMessage        `json:"configTemplate,omitempty"`
	ConfigLibraries   []v1.ResourceReference `json:"configLibraries,omitempty"`
}
