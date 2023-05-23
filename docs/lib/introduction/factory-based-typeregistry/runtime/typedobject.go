package runtime

// The name of this package - runtime - is inspired by kubernetes.  It defines generic types and corresponding methods.

// ArbitraryTypedObject only has a Type Field (with the json tag "type"). Thus, when unmarshaling something with more
// than just a type, all key:value-pairs within the serialized representation but "type:..." are ignored. This is
// leveraged in the Decode implementation.
type ArbitraryTypedObject struct {
	Type string `json:"type"`
}
