package main

// Arbitrary example program to illustrate the basic architectural concept of the ocm-lib.

// The usage of interfaces in this example is reduced to the necessary minimum to avoid distraction
// through implementation details.

import (
	"example/context/registry"

	// This package is imported for side effects!
	// It has to be imported so that the imports within types/init.go and consequently the init-functions within
	// types/simplemessage/message.go and types/complexmessage/message.go are executed and their respective types are
	// added (or rather registered) to the registy.DefaultMessageRegistry.
	_ "example/context/types"

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
	// Get the Default Message Context with the Default Message Registry (that contains all implemented types) and the
	// Default Setting of Uppercase = false
	messageCtx := registry.DefaultContext

	// Optionally edit the Settings of the Message Context
	messageCtx.PrintSettings.Uppercase = true

	// This context method encapsulates the dynamic unmarshaling based on the types registered in its local Message
	// Registry and returns a Message Spec. A Message Spec is an object with serializable attributes describing a
	// message. This Spec object also provides a factory method to create a corresponding message object that implements
	// the actual functionality.
	messageSpec, err := messageCtx.MessageSpecForConfig(data)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// This context method encapsulates the call of the factory method of message spec, implicitly passing itself as
	// a container for settings (PrintSettings.Uppercase) to the factory method
	messageObject := messageCtx.MessageForSpec(messageSpec)

	// Instead of calling the previous two methods, one could also simply call this one
	//messageObject, err := messageCtx.MessageForConfig(data)
	//if err != nil {
	//	fmt.Println(err)
	//	os.Exit(1)
	//}

	// Calling the Print function of the interface (which will use the implementation of the respective dynamic type)
	messageObject.Print()
}
