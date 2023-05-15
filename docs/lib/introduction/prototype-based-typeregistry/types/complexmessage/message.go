package complexmessage

import (
	"example/prototype-based-typeregistry/registry"
	"fmt"
)

// This package defines and registers a type that implements the Message interface.

const TYPE = "complexmessage"

func init() {
	registry.DefaultMessageRegistry.Register(TYPE, &Message{})
}

type Message struct {
	Type  string `json:"type"`
	Title string `json:"title"`
	Body  string `json:"body"`
}

func (r *Message) Print() {
	fmt.Println("Title: ", r.Title)
	fmt.Println("Body: ", r.Body)
}
