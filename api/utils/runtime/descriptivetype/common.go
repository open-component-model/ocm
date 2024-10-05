package descriptivetype

// TypeInfo is the interface extension for descriptive types.
type TypeInfo interface {
	Description() string
	Format() string
}

type typeInfoImpl struct {
	description string
	format      string
}

var _ TypeInfo = (*typeInfoImpl)(nil)

func (i *typeInfoImpl) Description() string {
	return i.description
}

func (i *typeInfoImpl) Format() string {
	return i.format
}

// DescriptionExtender provides an additional description for a type object
// which is appended to the format description in the scheme description
// for the type in question.
type DescriptionExtender[T any] func(t T) string
