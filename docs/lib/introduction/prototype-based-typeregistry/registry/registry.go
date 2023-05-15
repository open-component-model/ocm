package registry

import (
	"example/prototype-based-typeregistry/runtime"
	"fmt"
	"reflect"
	"sigs.k8s.io/yaml"
)

// Interfaces

type Message interface {
	Print()
}

//######################################################################################################################
// MessageTypeRegistry Implementation

// MessageTypeRegistry is the simplest form of a registry type.
// A prototype-based-typeregistry maps one model type to a dedicated go type ([model type]1:1[go type])
type MessageTypeRegistry map[string]Message

// DefaultMessageRegistry is a global variable that serves as default registry (thus, commonly all types register
// themselves at this registry)
// If you want to use a different configuration, therefore, only support a subset of available types (here, e.g. only
// simplemessage), you would have to create another instance of the MessageTypeRegistry type and "manually" call the
// Register on it.
var DefaultMessageRegistry = MessageTypeRegistry{}

// Register just adds a prototype object to the map.
// We call this parameter prototype as the only purpose of this object is to construct further objects of the same type.
func (m MessageTypeRegistry) Register(name string, prototype Message) {
	m[name] = prototype
}

// DecodeMessage takes the bytes representing the serialization of a certain type of Message as input and returns a
// Message interface as static type with the corresponding message object type as dynamic type
func (m MessageTypeRegistry) DecodeMessage(data []byte) (Message, error) {
	// Unmarshaling the data into a ArbitraryTypedObject.
	// As this object only has a Type Field (with the json tag "type"), all key:value-pairs within the serialized
	// representation but "type:..." are ignored.
	arbitraryTypedObject := runtime.ArbitraryTypedObject{}
	if err := yaml.Unmarshal(data, &arbitraryTypedObject); err != nil {
		return nil, fmt.Errorf("error unmarshaling content into typed object: ", err)
	}
	if arbitraryTypedObject.Type == "" {
		return nil, fmt.Errorf("error no type found")
	}

	// As the type attribute can now be accessed, it can be used to retrieve the corresponding Go type (here, the
	// corresponding Message struct type, or rather a ("prototype") object of that type.
	messagePrototypeObject, ok := m[arbitraryTypedObject.Type]
	if !ok {
		return nil, fmt.Errorf("error type unknown", arbitraryTypedObject.Type)
	}

	// Get the type of the ("prototype") object.
	objectType := reflect.TypeOf(messagePrototypeObject)
	for objectType.Kind() == reflect.Pointer {
		objectType = objectType.Elem()
	}
	// The reflect-library can be used to create a new instance of this ("prototype") object
	messageObject := reflect.New(objectType).Interface()

	// Unmarshaling the data again, this time into the suitable object
	if err := yaml.Unmarshal(data, messageObject); err != nil {
		return nil, fmt.Errorf("error unmarshaling content of into corresponding object: ", err)
	}

	return messageObject.(Message), nil
}
