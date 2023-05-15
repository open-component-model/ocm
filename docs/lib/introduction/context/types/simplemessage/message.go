package simplemessage

// This package defines and registers a type that implements the Message interface.

import (
	"example/context/registry"
	"example/context/util"
	"fmt"
	"sigs.k8s.io/yaml"
	"strings"
)

const TYPE = "simplemessage"

func init() {
	registry.DefaultMessageRegistry.Register(TYPE, &Factory{})
	// Alternatively, if the post-processing is not necessary the prototype-based message factory could be used for both
	// types (in this case, the type registry maps different model types to different objects of the same factory type).
	// registry.DefaultMessageRegistry.Register(TYPE, &registry.PrototypeBasedMessageSpecFactory{Prototype: &MessageSpec{}})
}

type MessageSpec struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func (spec *MessageSpec) Message(ctx *registry.Context) registry.Message {
	return &Message{
		ctx:  ctx,
		spec: spec,
	}
}

//######################################################################################################################

// Factory is an implementation of the MessageSpecFactory interface (Factory for MessageSpecs).
// A justified question could be, why each Message does not simply provide its own Decode method (and therefore be a
// Factory itself). This would be fine as long as one Decode method per Message is sufficient. Once you need a second
// one (here, e.g. a Decode method that also checks the spelling of the title), you would need methods with different
// name and therefore, you would have to extend the interface.
type Factory struct{}

func (f *Factory) Decode(data []byte) (registry.MessageSpec, error) {
	messageSpec := MessageSpec{}
	if err := yaml.Unmarshal(data, &messageSpec); err != nil {
		return nil, fmt.Errorf("error unmarshaling content of into corresponding object: %w", err)
	}
	// This method just demonstrates the possible necessity of type-specific post-processing activities that require
	// access to the structs actual fields
	util.CheckSpelling(&messageSpec.Text)
	return &messageSpec, nil
}

//######################################################################################################################

type Message struct {
	ctx  *registry.Context
	spec *MessageSpec
}

func (m *Message) Print() {
	text := m.spec.Text
	if m.ctx.PrintSettings.Uppercase {
		text = strings.ToUpper(text)
	}
	fmt.Println(text)
}
