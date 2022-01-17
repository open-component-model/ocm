package runtime

import (
	"strings"
)

// ObjectType describes the type of a object
// +k8s:deepcopy-gen=true
type ObjectType struct {
	// Type describes the type of the object.
	Type string `json:"type"`
}

// GetType returns the type of the object.
func (t ObjectType) GetType() string {
	return t.Type
}

// SetType sets the type of the object.
func (t *ObjectType) SetType(ttype string) {
	t.Type = ttype
}

// ObjectTypeVersion describes the type of a object
// +k8s:deepcopy-gen=true
type ObjectTypeVersion struct {
	ObjectType `json:",inline"`
}

// NewObjectTypeVersion returns a type version object
func NewObjectTypeVersion(t string) ObjectTypeVersion {
	return ObjectTypeVersion{
		ObjectType{
			Type: t,
		},
	}
}

// GetName returns the name part of the type
func (s *ObjectTypeVersion) GetName() string {
	t := s.GetType()
	i := strings.LastIndex(t, "/")
	if i < 0 {
		return t
	}
	return t[:i]
}

// GetVersion returns the version part of the type
func (s *ObjectTypeVersion) GetVersion() string {
	t := s.GetType()
	i := strings.LastIndex(t, "/")
	if i < 0 {
		return "v1"
	}
	return t[i+1:]
}
