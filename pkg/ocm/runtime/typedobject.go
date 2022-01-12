package runtime

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
