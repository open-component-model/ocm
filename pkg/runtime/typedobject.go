package runtime

// ObjectType describes the type of a object
type ObjectType struct {
	// Type describes the type of the object.
	Type string `json:"type"`
}

// NewObjectType creates an ObjectType value
func NewObjectType(typ string) ObjectType {
	return ObjectType{typ}
}

// GetType returns the type of the object.
func (t ObjectType) GetType() string {
	return t.Type
}

// SetType sets the type of the object.
func (t *ObjectType) SetType(typ string) {
	t.Type = typ
}
