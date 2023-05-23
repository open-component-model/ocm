package registry

import (
	"example/factory-based-typeregistry/runtime"
	"fmt"
	"reflect"
	"sigs.k8s.io/yaml"
)

// Interfaces

type Message interface {
	Print()
}

type MessageFactory interface {
	Decode(data []byte) (Message, error)
}

// ######################################################################################################################
// MessageTypeRegistry Implementation

// MessageTypeRegistry is a factory-based-registry.
// In case of the PrototypeBasedMessageFactory, a factory-based-typeregistry maps n model types to 1 factory type
// ([model types]n:1[go type]).
// In case of the type specific implementations of the MessageFactory (Factories within the simplemessage and
// complexmessage directories), a factory-based-typeregistry maps 1 model type to 1 factory type
// ([model type]1:1[go type]).
type MessageTypeRegistry map[string]MessageFactory

// DefaultMessageRegistry is a global variable that serves as default registry (thus, commonly all types register
// themselves at this registry)
// If you want to use a different configuration, therefore, only support a subset of available types (here, e.g. only
// simplemessage), you would have to create another instance of the MessageTypeRegistry type and "manually" call the
// Register on it.
var DefaultMessageRegistry = MessageTypeRegistry{}

// Register just adds a prototype object to the map.
// We call this parameter prototype as the only purpose of this object is to construct further objects of the same type.
func (m MessageTypeRegistry) Register(name string, factory MessageFactory) {
	m[name] = factory
}

// DecodeMessage takes the bytes representing the serialization of a certain type of Message as input and returns a
// Message interface as static type with the corresponding message object type as dynamic type
func (m MessageTypeRegistry) DecodeMessage(data []byte) (Message, error) {
	// Unmarshaling the data into a ArbitraryTypedObject.
	// As this object only has a Type Field (with the json tag "type"), all key:value-pairs within the serialized
	// representation but "type:..." are ignored.
	arbitraryTypedObject := runtime.ArbitraryTypedObject{}
	if err := yaml.Unmarshal(data, &arbitraryTypedObject); err != nil {
		return nil, fmt.Errorf("error unmarshaling content into typed object: %w", err)
	}
	if arbitraryTypedObject.Type == "" {
		return nil, fmt.Errorf("error no type found")
	}

	// As the type attribute can now be accessed, it can be used to retrieve the corresponding registered factory.
	messageFactory, ok := m[arbitraryTypedObject.Type]
	if !ok {
		return nil, fmt.Errorf("error type unknown %v", arbitraryTypedObject.Type)
	}

	// Unmarshaling the data again, this time into the suitable object
	messageObject, err := messageFactory.Decode(data)
	if err != nil {
		return nil, err
	}
	return messageObject, nil
}

//######################################################################################################################
// Generic Factory Implementation
// (Generic, as it is a factory that can produce objects of arbitrary dynamic type)

// PrototypeBasedMessageFactory is quite similar to the prototype-based-typeregistry. A PrototypeBasedMessageFactory
// also assumes that there exists a dedicated go type for each model type AND that it is sufficient to directly decode
// the serialized data into an empty object of that type without requiring further processing.
type PrototypeBasedMessageFactory struct {
	Prototype Message
}

func (p *PrototypeBasedMessageFactory) Decode(data []byte) (Message, error) {
	// Get the type of the ("prototype") object.
	objectType := reflect.TypeOf(p.Prototype)
	for objectType.Kind() == reflect.Pointer {
		objectType = objectType.Elem()
	}
	// The reflect-library can be used to create a new instance of this ("prototype") object
	messageObject := reflect.New(objectType).Interface()

	// Unmarshaling the data again, this time into the suitable object
	if err := yaml.Unmarshal(data, messageObject); err != nil {
		return nil, fmt.Errorf("error unmarshaling content of into corresponding object: %w", err)
	}

	return messageObject.(Message), nil
}
