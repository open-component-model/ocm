package tech

// UniformAccessSpecInfo describes a rough uniform specification for
// an access location or an accessed object. It not necessarily
// provided the exact access information required to technically
// access the object, but just some general information usable
// independently of the particular technical access specification
// to figure aut some general information in a formal way about the access.
type UniformAccessSpecInfo struct {
	Kind string
	Host string
	Port string
	Path string

	Info string
}
