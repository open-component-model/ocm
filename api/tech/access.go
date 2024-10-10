package tech

// UniformAccessSpecInfo describes a rough uniform specification for
// an access location or an accessed object. It not necessarily
// provided the exact access information required to technically
// access the object, but just some general information usable
// independently of the particular technical access specification
// to figure aut some general information in a formal way about the access.
type UniformAccessSpecInfo struct {
	Kind string `json:"kind"`
	Host string `json:"host,omitempty"`
	Port string `json:"port,omitempty"`
	Path string `json:"path,omitempty"`

	Info string `json:"info,omitempty"`
}
