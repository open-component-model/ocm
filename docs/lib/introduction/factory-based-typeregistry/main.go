package main

// Arbitrary example program to illustrate the basic architectural concept of the ocm-lib.
//
// The usage of interfaces and functions in this example is reduced to the necessary minimum to avoid distraction
// through implementation details.
//
// The factory-based-typeregistry offers much more flexibility than the prototype-based-typeregistry.

import (
	"example/factory-based-typeregistry/registry"

	// This package is imported for side effects!
	// It has to be imported so that the imports within types/init.go and consequently the init-functions within
	// types/simplemessage/message.go and types/complexmessage/message.go are executed and their respective types are
	// added (or rather registered) to the registy.DefaultMessageRegistry.
	_ "example/factory-based-typeregistry/types"

	"fmt"
	"os"
)

const SIMPLEFILE = "serializedmessages/simplemessage.yaml"
const COMPLEXFILE = "serializedmessages/complexmessage.yaml"

func main() {
	// You can switch between SIMPLEFILE and COMPLEXFILE to experiment with the behaviour.
	filepath := SIMPLEFILE
	data, err := os.ReadFile(filepath)
	if err != nil {
		fmt.Println("error reading ", filepath, ":", err)
		os.Exit(1)
	}

	//Decoding Logic
	messageObject, err := registry.DefaultMessageRegistry.DecodeMessage(data)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Calling the Print function of the interface (which will use the implementation of the respective dynamic type)
	messageObject.Print()
}
