package simplemessage

// This package defines and registers a type that implements the Message interface.

import (
	"example/prototype-based-typeregistry/registry"
	"fmt"
)

const TYPE = "simplemessage"

func init() {
	registry.DefaultMessageRegistry.Register(TYPE, &Message{})
}

type Message struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func (m *Message) Print() {
	fmt.Println(m.Text)
}
